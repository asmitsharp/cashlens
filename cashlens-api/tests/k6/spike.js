import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 100 }, // Below normal load
    { duration: '1m', target: 100 },
    { duration: '10s', target: 500 }, // Spike to 500 users
    { duration: '3m', target: 500 }, // Stay at 500 users
    { duration: '10s', target: 100 }, // Scale down. Recovery stage.
    { duration: '3m', target: 100 },
    { duration: '10s', target: 0 },
  ],
};

export default function () {
  const res = http.get('http://localhost:8080/health');
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
  sleep(1);
}
