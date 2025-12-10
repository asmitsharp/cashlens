import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 }, // Below normal load
    { duration: '5m', target: 100 },
    { duration: '2m', target: 200 }, // Normal load
    { duration: '5m', target: 200 },
    { duration: '2m', target: 300 }, // Around the breaking point
    { duration: '5m', target: 300 },
    { duration: '2m', target: 400 }, // Beyond the breaking point
    { duration: '5m', target: 400 },
    { duration: '10m', target: 0 }, // Scale down. Recovery stage.
  ],
};

export default function () {
  const res = http.get('http://localhost:8080/health');
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
  sleep(1);
}
