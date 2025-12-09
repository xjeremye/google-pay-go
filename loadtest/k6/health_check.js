import http from 'k6/http';
import { check } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// 自定义指标
const errorRate = new Rate('errors');
const healthCheckTime = new Trend('health_check_time');

// 配置 - 健康检查可以更激进一些
export const options = {
  stages: [
    { duration: '10s', target: 100 },  // 快速预热
    { duration: '1m', target: 500 },   // 稳定500并发
    { duration: '2m', target: 1000 },  // 增长到1000并发
    { duration: '1m', target: 500 },   // 下降
    { duration: '10s', target: 0 },    // 冷却
  ],
  thresholds: {
    http_req_duration: ['p(95)<100', 'p(99)<200'], // 健康检查应该很快
    http_req_failed: ['rate<0.001'],                // 错误率应该极低
    errors: ['rate<0.001'],
  },
};

// 测试数据配置
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const startTime = Date.now();
  const response = http.get(`${BASE_URL}/health`, {
    headers: {
      'User-Agent': 'k6-loadtest/1.0',
    },
    tags: { name: 'HealthCheck' },
  });

  const duration = Date.now() - startTime;
  healthCheckTime.add(duration);

  const success = check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
    'has database status': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.database !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  if (!success) {
    errorRate.add(1);
  }
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'loadtest/results/health_check_summary.json': JSON.stringify(data, null, 2),
  };
}

function textSummary(data, options) {
  const indent = options.indent || '  ';
  let summary = '\n' + indent + '健康检查压测结果\n';
  summary += indent + '='.repeat(50) + '\n';
  
  if (data.metrics.http_req_duration) {
    summary += indent + `平均响应时间: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
    summary += indent + `P95响应时间: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
    summary += indent + `P99响应时间: ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms\n`;
  }
  
  if (data.metrics.http_reqs) {
    summary += indent + `总请求数: ${data.metrics.http_reqs.values.count}\n`;
    summary += indent + `QPS: ${data.metrics.http_reqs.values.rate.toFixed(2)}\n`;
  }
  
  if (data.metrics.http_req_failed) {
    summary += indent + `错误率: ${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%\n`;
  }
  
  summary += indent + '='.repeat(50) + '\n';
  return summary;
}
