import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { randomString, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Custom metrics
const ingestionRate = new Rate('ingestion_success');
const ingestionLatency = new Trend('ingestion_latency');

export const options = {
  scenarios: {
    ramp_up: {
      executor: 'ramping-vus',
      startVUs: 10,
      stages: [
        { duration: '2m', target: 50 },    // Warm up
        { duration: '5m', target: 200 },   // Ramp to target
        { duration: '10m', target: 200 },  // Sustained load
        { duration: '2m', target: 500 },   // Spike test
        { duration: '5m', target: 200 },   // Recovery
        { duration: '2m', target: 0 },     // Ramp down
      ],
    },
  },
  thresholds: {
    'ingestion_success': ['rate>0.99'],
    'ingestion_latency': ['p(99)<5000'],
  },
};

const SERVICES = ['api-gateway', 'auth-service', 'user-profile', 'payment-processor', 'inventory-manager'];
const LEVELS = ['INFO', 'WARN', 'ERROR', 'DEBUG'];

function generateLog() {
  const service = SERVICES[randomIntBetween(0, SERVICES.length - 1)];
  const level = LEVELS[randomIntBetween(0, LEVELS.length - 1)];
  return {
    timestamp: new Date().toISOString(),
    service: service,
    level: level,
    message: `This is a simulated ${level} log from ${service} - ${randomString(20)}`,
    trace_id: randomString(16),
    span_id: randomString(8),
    duration_ms: randomIntBetween(1, 1000)
  };
}

export default function () {
  const url = __ENV.INGESTION_URL || 'http://localhost:8080/vector';
  const payload = JSON.stringify(generateLog());
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(url, payload, params);

  const success = check(res, {
    'is status 200 or 202': (r) => r.status === 200 || r.status === 202,
  });

  ingestionRate.add(success);
  ingestionLatency.add(res.timings.duration);

  // Short sleep to simulate real-world request patterns
  sleep(randomIntBetween(0, 10) / 1000.0);
}
