// K6 Load Testing Script for Ticketing System
// More sophisticated than hey - supports scenarios, thresholds, and stages

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const eventListDuration = new Trend('event_list_duration');
const searchDuration = new Trend('search_duration');
const totalRequests = new Counter('total_requests');

// Configuration
const API_URL = 'https://ticketingsystem-production-4a1d.up.railway.app';

// Test configuration - choose your scenario
export const options = {
    // Scenario 1: Gradual ramp-up (uncomment to use)
    stages: [
        { duration: '30s', target: 10 },   // Ramp up to 10 users
        { duration: '1m', target: 50 },    // Ramp up to 50 users
        { duration: '2m', target: 100 },   // Ramp up to 100 users
        { duration: '1m', target: 50 },    // Ramp down to 50 users
        { duration: '30s', target: 0 },    // Ramp down to 0 users
    ],
    
    // Performance thresholds
    thresholds: {
        http_req_duration: ['p(95)<2000', 'p(99)<3000'], // 95% of requests should be below 2s
        http_req_failed: ['rate<0.1'],     // Error rate should be less than 10%
        errors: ['rate<0.1'],               // Custom error rate
    },
};

// Scenario 2: Constant load (alternative config)
// export const options = {
//     vus: 50,              // 50 virtual users
//     duration: '5m',       // Run for 5 minutes
// };

// Scenario 3: Spike test (alternative config)
// export const options = {
//     stages: [
//         { duration: '10s', target: 10 },
//         { duration: '10s', target: 200 },  // Sudden spike
//         { duration: '30s', target: 200 },  // Hold spike
//         { duration: '10s', target: 10 },
//     ],
// };

export default function () {
    totalRequests.add(1);
    
    // Test 1: Health check
    let healthRes = http.get(`${API_URL}/health`);
    check(healthRes, {
        'health check status is 200': (r) => r.status === 200,
    }) || errorRate.add(1);
    
    sleep(1);
    
    // Test 2: List events
    let eventsRes = http.get(`${API_URL}/events`);
    eventListDuration.add(eventsRes.timings.duration);
    check(eventsRes, {
        'events list status is 200': (r) => r.status === 200,
        'events list has data': (r) => r.body.length > 0,
    }) || errorRate.add(1);
    
    sleep(2);
    
    // Test 3: Search events
    const searchQueries = ['concert', 'festival', 'sports', 'conference', 'workshop'];
    const randomQuery = searchQueries[Math.floor(Math.random() * searchQueries.length)];
    
    let searchRes = http.get(`${API_URL}/events/search?query=${randomQuery}`);
    searchDuration.add(searchRes.timings.duration);
    check(searchRes, {
        'search status is 200': (r) => r.status === 200,
    }) || errorRate.add(1);
    
    sleep(2);
    
    // Test 4: Get metrics (verify monitoring is working)
    let metricsRes = http.get(`${API_URL}/metrics`);
    check(metricsRes, {
        'metrics available': (r) => r.status === 200,
    }) || errorRate.add(1);
    
    sleep(3);
}

// Setup function (runs once at the beginning)
export function setup() {
    console.log('Starting load test...');
    console.log(`Target: ${API_URL}`);
}

// Teardown function (runs once at the end)
export function teardown(data) {
    console.log('Load test completed!');
    console.log('Check Grafana dashboard for detailed metrics');
}
