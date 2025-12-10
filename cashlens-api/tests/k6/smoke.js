import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 1, // 1 Virtual User
  duration: '1m', // Run for 1 minute
};

export default function () {
  const res = http.get('http://localhost:8080/health');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  sleep(1);
}
