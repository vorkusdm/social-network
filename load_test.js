import http from 'k6/http';
import { check, sleep } from 'k6';

// Конфигурация сценариев: 1, 10, 100 и 1000 одновременных пользователей (VUs)
export const options = {
  scenarios: {
    load_1_vu: {
      executor: 'constant-vus',
      vus: 1,
      duration: '30s',
    },
    load_10_vu: {
      executor: 'constant-vus',
      vus: 10,
      duration: '30s',
      startTime: '30s',
    },
    load_100_vu: {
      executor: 'constant-vus',
      vus: 100,
      duration: '30s',
      startTime: '60s',
    },
    load_1000_vu: {
      executor: 'constant-vus',
      vus: 1000,
      duration: '30s',
      startTime: '90s',
    },
  },
  // Пороговые значения (опционально, для проверки стабильности)
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95% запросов должны быть быстрее 2 секунд
  },
};

export default function () {
  // ВАЖНО: Проверьте через `head -n 20 people.v2.csv`, какие имена там есть!
  const url = 'http://localhost:8080/user/search?first_name=Vorkus&last_name=Иванов&limit=50';
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.get(url, params);
  
  // Проверяем, что сервер ответил успешно
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 2000ms': (r) => r.timings.duration < 2000,
  });

  // Небольшая пауза между запросами, чтобы имитировать реальное поведение
  sleep(0.1);
}