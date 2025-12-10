import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '5m', target: 50 }, // Ramp up to 50 users over 5 minutes
    { duration: '10m', target: 50 }, // Stay at 50 users for 10 minutes
    { duration: '5m', target: 0 }, // Ramp down to 0 users
  ],
};

export default function () {
  const res = http.get('http://localhost:8080/health');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 1000ms': (r) => r.timings.duration < 1000,
  });
  sleep(1);
}
