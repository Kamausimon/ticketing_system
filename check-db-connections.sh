#!/bin/bash

# Check database connections for ticketing_system
# Shows detailed connection status and identifies potential issues

export PGPASSWORD=postgres

echo "📊 Database Connection Status for ticketing_system"
echo "=================================================="
echo ""

echo "🔍 Current Active Connections:"
psql -h localhost -U postgres -d ticketing_system -c "
SELECT 
    pid,
    usename,
    application_name,
    client_addr,
    state,
    wait_event_type,
    wait_event,
    state_change,
    NOW() - state_change as idle_duration,
    LEFT(query, 80) as query_preview
FROM pg_stat_activity
WHERE datname = 'ticketing_system'
ORDER BY state_change;
"

echo ""
echo "📈 Connection Summary by State:"
psql -h localhost -U postgres -d ticketing_system -c "
SELECT 
    state,
    wait_event_type,
    wait_event,
    COUNT(*) as connection_count
FROM pg_stat_activity
WHERE datname = 'ticketing_system'
GROUP BY state, wait_event_type, wait_event
ORDER BY connection_count DESC;
"

echo ""
echo "⚠️  Long-Running/Idle Transactions:"
psql -h localhost -U postgres -d ticketing_system -c "
SELECT 
    pid,
    state,
    xact_start,
    NOW() - xact_start as transaction_duration,
    query_start,
    LEFT(query, 100) as query_preview
FROM pg_stat_activity
WHERE datname = 'ticketing_system'
  AND xact_start IS NOT NULL
  AND NOW() - xact_start > INTERVAL '10 seconds'
ORDER BY xact_start;
"

echo ""
echo "🔒 Blocked Queries:"
psql -h localhost -U postgres -d ticketing_system -c "
SELECT 
    blocked_locks.pid AS blocked_pid,
    blocked_activity.usename AS blocked_user,
    blocking_locks.pid AS blocking_pid,
    blocking_activity.usename AS blocking_user,
    blocked_activity.query AS blocked_statement,
    blocking_activity.query AS current_statement_in_blocking_process
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks 
    ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted
  AND blocked_activity.datname = 'ticketing_system';
"

echo ""
echo "💡 Recommendations:"
echo "  - Idle connections in 'ClientRead' state should be < 5"
echo "  - Long-running transactions (>10s) should be investigated"
echo "  - Blocked queries indicate potential deadlocks"
echo ""
