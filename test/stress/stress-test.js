import http from 'k6/http';
import { check, sleep } from 'k6';

// ===============================
// CONFIGURAÇÕES
// ===============================

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';
const EMAIL = __ENV.EMAIL;
const PASSWORD = __ENV.PASSWORD;

// Parametros de configuração de thresholds via variáveis de ambiente
const MAX_VUS = Number(__ENV.MAX_VUS || 100); // número máximo de usuários virtuais
const P95_MS = Number(__ENV.P95_MS || 800); // as requisições devem ser respondidas em até 800ms
const ERROR_RATE = Number(__ENV.ERROR_RATE || 0.01);  // taxa máxima de erro permitida (1%)

export const options = {
  stages: [
    { duration: '30s', target: Math.floor(MAX_VUS * 0.2) },
    { duration: '1m',  target: Math.floor(MAX_VUS * 0.5) },
    { duration: '1m',  target: MAX_VUS },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: [`p(95)<${P95_MS}`],
    http_req_failed: [`rate<${ERROR_RATE}`],
  },
};

export function setup() {
  const payload = JSON.stringify({
    email: EMAIL,
    password: PASSWORD,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(`${BASE_URL}/sessions`, payload, params);

  check(res, {
    'login status 200': (r) => r.status === 200,
    'token exists': (r) => r.json('data.access_token') !== undefined,
  });

  return {
    token: res.json('data.access_token'),
  };
}

export default function (data) {
  const headers = {
    Authorization: `Bearer ${data.token}`,
    'Content-Type': 'application/json',
  };

  // -------- USERS --------
  const usersRes = http.get(
    `${BASE_URL}/users`,
    { headers }
  );

  check(usersRes, {
    'users status 200': (r) => r.status === 200,
  });

  // -------- PRODUCTS --------
  const productsRes = http.get(
    `${BASE_URL}/products`,
    { headers }
  );

  check(productsRes, {
    'products status 200': (r) => r.status === 200,
  });

  // -------- CUSTOMERS --------
  const customersRes = http.get(
    `${BASE_URL}/customers`,
    { headers }
  );

  check(customersRes, {
    'customers status 200': (r) => r.status === 200,
  });

  sleep(1);
}
