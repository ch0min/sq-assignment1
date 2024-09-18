import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 50 }, // Ramp up to 50 users in 1 minute
    { duration: '2m', target: 50 }, // Stay at 50 users for 2 minutes
    { duration: '1m', target: 0 },  // Ramp down to 0 users in 1 minute
  ],
};

export default function () {
  const res = http.get('http://localhost:4000/api/todos'); // Replace with your endpoint
  check(res, {
    'is status 200': (r) => r.status === 200,
  });
  sleep(1);
}
