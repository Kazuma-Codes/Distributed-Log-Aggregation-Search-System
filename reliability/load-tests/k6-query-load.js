import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Rate } from 'k6/metrics';
import { randomString, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const searchLatency = new Trend('search_latency');
const statsLatency = new Trend('stats_latency');
const traceLatency = new Trend('trace_latency');
const querySuccess = new Rate('query_success');

export const options = {
  scenarios: {
    search_queries: {
      executor: 'constant-vus',
      vus: 60, // 60% search
      duration: '10m',
      exec: 'search',
    },
    stats_queries: {
      executor: 'constant-vus',
      vus: 25, // 25% stats
      duration: '10m',
      exec: 'stats',
    },
    trace_queries: {
      executor: 'constant-vus',
      vus: 15, // 15% trace
      duration: '10m',
      exec: 'trace',
    },
  },
  thresholds: {
    'search_latency': ['p(99)<2000'],
    'stats_latency': ['p(99)<500'],
    'query_success': ['rate>0.99'],
  },
};

const BASE_URL = __ENV.QUERY_URL || 'http://localhost:8081/api/v1';
const SERVICES = ['api-gateway', 'auth-service', 'user-profile', 'payment-processor', 'inventory-manager'];

function getRandomService() {
  return SERVICES[randomIntBetween(0, SERVICES.length - 1)];
}

export function search() {
  const service = getRandomService();
  const url = `${BASE_URL}/search?service=${service}&limit=100`;
  const res = http.get(url);
  const success = check(res, { 'status is 200': (r) => r.status === 200 });
  querySuccess.add(success);
  searchLatency.add(res.timings.duration);
  sleep(1);
}

export function stats() {
  const service = getRandomService();
  const url = `${BASE_URL}/stats?service=${service}&interval=1h`;
  const res = http.get(url);
  const success = check(res, { 'status is 200': (r) => r.status === 200 });
  querySuccess.add(success);
  statsLatency.add(res.timings.duration);
  sleep(1);
}

export function trace() {
  const traceId = randomString(16);
  const url = `${BASE_URL}/trace/${traceId}`;
  const res = http.get(url);
  const success = check(res, { 'status is 200': (r) => r.status === 200 });
  querySuccess.add(success);
  traceLatency.add(res.timings.duration);
  sleep(1);
}
