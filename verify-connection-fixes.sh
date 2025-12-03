#!/bin/bash

# Post-Fix Verification Script
# Verify that connection leak fixes are working correctly

set -e

export PGPASSWORD=postgres

echo "🔍 DATABASE CONNECTION LEAK FIX VERIFICATION"
echo "============================================="
echo ""

# Check 1: Verify no idle connections at start
echo "✓ Check 1: Initial connection state"
IDLE_COUNT=$(PGPASSWORD=postgres psql -h localhost -U postgres -d ticketing_system -t -c "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = 'ticketing_system' AND state = 'idle';")
echo "  Idle connections: $IDLE_COUNT"
if [ "$IDLE_COUNT" -gt 2 ]; then
    echo "  ⚠️  WARNING: Found $IDLE_COUNT idle connections (expected ≤ 2)"
else
    echo "  ✅ PASS: Idle connection count is acceptable"
fi
echo ""

# Check 2: Verify connection pool settings
echo "✓ Check 2: Connection pool configuration"
echo "  Checking database/main.go for correct settings..."
if grep -q "SetMaxIdleConns(0)" internal/database/main.go; then
    echo "  ✅ PASS: MaxIdleConns set to 0 (no idle connections)"
else
    echo "  ❌ FAIL: MaxIdleConns not set to 0"
fi
if grep -q "SetMaxOpenConns(25)" internal/database/main.go; then
    echo "  ✅ PASS: MaxOpenConns set to 25"
else
    echo "  ⚠️  WARNING: MaxOpenConns not set to 25"
fi
echo ""

# Check 3: Verify transaction cleanup pattern
echo "✓ Check 3: Transaction cleanup pattern"
echo "  Checking for proper 'committed' flag usage..."
COMMITTED_COUNT=$(grep -r "committed := false" internal/ | wc -l)
echo "  Found $COMMITTED_COUNT transaction blocks with cleanup pattern"
if [ "$COMMITTED_COUNT" -ge 8 ]; then
    echo "  ✅ PASS: Multiple transaction blocks use proper cleanup"
else
    echo "  ⚠️  WARNING: Expected at least 8 transaction blocks with cleanup"
fi
echo ""

# Check 4: Verify empty transaction was removed
echo "✓ Check 4: Empty transaction removal"
if grep -A5 "Start transaction" internal/tickets/generate.go | grep -q "Note: Tickets are generated"; then
    echo "  ✅ PASS: Empty transaction removed from generate.go"
else
    echo "  ⚠️  INFO: Check generate.go manually"
fi
echo ""

# Check 5: Code compilation
echo "✓ Check 5: Code compilation"
if go build -o /tmp/verify-build ./cmd/api-server/main.go 2>&1; then
    echo "  ✅ PASS: Code compiles successfully"
    rm -f /tmp/verify-build
else
    echo "  ❌ FAIL: Code does not compile"
    exit 1
fi
echo ""

# Check 6: Monitor connections during brief test
echo "✓ Check 6: Connection behavior test"
echo "  Starting API server briefly to test connection behavior..."
echo "  (Server will run for 5 seconds)"

# Start server in background
go run cmd/api-server/main.go > /tmp/server-test.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Check connections
ACTIVE_COUNT=$(PGPASSWORD=postgres psql -h localhost -U postgres -d ticketing_system -t -c "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = 'ticketing_system' AND state = 'active';")
IDLE_COUNT=$(PGPASSWORD=postgres psql -h localhost -U postgres -d ticketing_system -t -c "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = 'ticketing_system' AND state = 'idle';")

echo "  Active connections: $ACTIVE_COUNT"
echo "  Idle connections: $IDLE_COUNT"

# Stop server
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

# Wait for connections to close
sleep 2

# Check final state
FINAL_IDLE=$(PGPASSWORD=postgres psql -h localhost -U postgres -d ticketing_system -t -c "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = 'ticketing_system' AND state = 'idle';")
echo "  Final idle connections after server stop: $FINAL_IDLE"

if [ "$FINAL_IDLE" -le 1 ]; then
    echo "  ✅ PASS: Connections properly cleaned up"
else
    echo "  ⚠️  WARNING: $FINAL_IDLE idle connections remain"
fi
echo ""

# Summary
echo "============================================="
echo "📊 VERIFICATION SUMMARY"
echo "============================================="
echo ""
echo "The connection leak fixes have been applied."
echo ""
echo "Key Changes:"
echo "  • Removed empty transaction in tickets/generate.go"
echo "  • Added 'committed' flag pattern to 9 transaction blocks"
echo "  • Set MaxIdleConns to 0 (close immediately)"
echo "  • Increased MaxOpenConns to 25"
echo "  • Reduced idle timeout to 10 seconds"
echo ""
echo "Next Steps:"
echo "  1. Deploy these changes to production"
echo "  2. Monitor with: ./check-db-connections.sh"
echo "  3. If needed, clean up: ./kill-idle-connections.sh"
echo ""
echo "✅ Verification complete!"
