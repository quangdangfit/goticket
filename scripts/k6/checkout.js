// k6 script: ramp 0 → 500 vUs over 1m, hold 2m at 500, ramp down.
// Run: k6 run -e BASE=http://localhost:8080 -e TOKEN=<jwt> scripts/k6/checkout.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
  stages: [
    { duration: '60s', target: 500 },
    { duration: '120s', target: 500 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    'checks{tag:checkout}': ['rate>0.99'],
  },
};

const BASE = __ENV.BASE || 'http://localhost:8080';
const TOKEN = __ENV.TOKEN;
const SHOWTIME = __ENV.SHOWTIME;
const TICKET_TYPE = __ENV.TICKET_TYPE;

export default function () {
  const body = JSON.stringify({
    idempotency_key: uuidv4(),
    showtime_id: SHOWTIME,
    items: [{ showtime_id: SHOWTIME, ticket_type_id: TICKET_TYPE, quantity: 1 }],
  });
  const res = http.post(`${BASE}/api/v1/orders`, body, {
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${TOKEN}` },
    tags: { name: 'checkout' },
  });
  check(res, { 'status 201|409|410': (r) => [201, 409, 410].includes(r.status) }, { tag: 'checkout' });
  sleep(0.2);
}
