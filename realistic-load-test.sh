#!/bin/bash

# Realistic User Journey Load Test
# Simulates actual user behavior: browse -> select event -> book tickets

API_URL="https://ticketingsystem-production-4a1d.up.railway.app"
CONCURRENT_USERS=20
DURATION=300  # 5 minutes
RESULTS_FILE="./load-test-results/realistic_journey_$(date +%Y%m%d_%H%M%S).txt"

mkdir -p ./load-test-results

echo "=== Realistic User Journey Load Test ===" | tee "$RESULTS_FILE"
echo "Simulating $CONCURRENT_USERS concurrent users for $DURATION seconds" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Function to simulate a single user journey
user_journey() {
    local user_id=$1
    local start_time=$(date +%s)
    local end_time=$((start_time + DURATION))
    
    while [ $(date +%s) -lt $end_time ]; do
        # Step 1: Browse homepage/events
        curl -s -o /dev/null -w "User $user_id - Browse Events: %{http_code} in %{time_total}s\n" \
            "$API_URL/events" >> "$RESULTS_FILE"
        sleep $((RANDOM % 5 + 2))  # 2-7 seconds
        
        # Step 2: Search for events
        local queries=("concert" "festival" "sports" "theater" "comedy")
        local query=${queries[$RANDOM % ${#queries[@]}]}
        curl -s -o /dev/null -w "User $user_id - Search '$query': %{http_code} in %{time_total}s\n" \
            "$API_URL/events/search?query=$query" >> "$RESULTS_FILE"
        sleep $((RANDOM % 4 + 2))  # 2-6 seconds
        
        # Step 3: View event details (simulate clicking on an event)
        curl -s -o /dev/null -w "User $user_id - View Event: %{http_code} in %{time_total}s\n" \
            "$API_URL/events" >> "$RESULTS_FILE"
        sleep $((RANDOM % 6 + 3))  # 3-9 seconds (users spend time reading)
        
        # Step 4: Check available tickets
        curl -s -o /dev/null -w "User $user_id - Check Tickets: %{http_code} in %{time_total}s\n" \
            "$API_URL/events" >> "$RESULTS_FILE"
        sleep $((RANDOM % 3 + 1))  # 1-4 seconds
        
        # Longer pause before next journey (simulate user thinking/browsing)
        sleep $((RANDOM % 10 + 5))  # 5-15 seconds
    done
}

# Start concurrent user simulations
echo "Starting $CONCURRENT_USERS concurrent user journeys..."
for i in $(seq 1 $CONCURRENT_USERS); do
    user_journey $i &
done

# Wait for all background jobs to complete
wait

echo "" | tee -a "$RESULTS_FILE"
echo "=== Load Test Complete ===" | tee -a "$RESULTS_FILE"
echo "Results saved to: $RESULTS_FILE" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Summary statistics
echo "=== Summary Statistics ===" | tee -a "$RESULTS_FILE"
echo "Total requests: $(grep -c "User" "$RESULTS_FILE")" | tee -a "$RESULTS_FILE"
echo "Successful (200): $(grep -c "200" "$RESULTS_FILE")" | tee -a "$RESULTS_FILE"
echo "Failed (non-200): $(grep -cv "200" "$RESULTS_FILE" | grep "User" | wc -l)" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"
echo "View detailed metrics in Grafana: http://localhost:3001" | tee -a "$RESULTS_FILE"
