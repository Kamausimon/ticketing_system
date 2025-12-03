#!/bin/bash

# Kill idle database connections for ticketing_system
# This script terminates idle connections that are in ClientRead state

export PGPASSWORD=postgres

echo "🔍 Finding idle connections in ticketing_system database..."

# Query to find and kill idle connections
psql -h localhost -U postgres -d ticketing_system -c "
SELECT 
    pid,
    state,
    wait_event_type,
    wait_event,
    query_start,
    state_change,
    query
FROM pg_stat_activity
WHERE datname = 'ticketing_system'
  AND state = 'idle'
  AND wait_event_type = 'Client'
  AND wait_event = 'ClientRead'
  AND state_change < NOW() - INTERVAL '30 seconds';
"

echo ""
echo "⚠️  Killing idle connections older than 30 seconds..."

# Kill idle connections
psql -h localhost -U postgres -d ticketing_system -c "
SELECT 
    pg_terminate_backend(pid),
    pid,
    state,
    wait_event
FROM pg_stat_activity
WHERE datname = 'ticketing_system'
  AND state = 'idle'
  AND wait_event_type = 'Client'
  AND wait_event = 'ClientRead'
  AND state_change < NOW() - INTERVAL '30 seconds';
"

echo ""
echo "✅ Idle connections terminated!"
echo ""
echo "📊 Current connection status:"

# Show current connections
psql -h localhost -U postgres -d ticketing_system -c "
SELECT 
    state,
    wait_event_type,
    wait_event,
    COUNT(*) as count
FROM pg_stat_activity
WHERE datname = 'ticketing_system'
GROUP BY state, wait_event_type, wait_event
ORDER BY count DESC;
"
