import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { SharedArray } from 'k6/data';

// 自定义指标
const errorRate = new Rate('errors');
const orderCreationTime = new Trend('order_creation_time');
const orderQueryTime = new Trend('order_query_time');
const healthCheckTime = new Trend('health_check_time');

// 配置 - 混合场景
export const options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '2m', target: 150 },
    { duration: '3m', target: 250 },
    { duration: '2m', target: 150 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],
    errors: ['rate<0.01'],
  },
};

// 测试数据配置
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const MERCHANT_ID = __ENV.MERCHANT_ID || '1';
const CHANNEL_ID = __ENV.CHANNEL_ID || '1';
const MERCHANT_KEY = __ENV.MERCHANT_KEY || 'your_merchant_key';

// 存储已创建的订单号（用于查询测试）
const createdOrders = new SharedArray('orders', function () {
  return [];
});

// 生成签名（与后端一致的签名算法）
function generateSign(params, key) {
  const crypto = require('k6/crypto');
  
  // 1. 移除 sign 字段，过滤掉 null 值
  const filtered = {};
  for (const k in params) {
    if (k !== 'sign' && params[k] != null) {
      filtered[k] = params[k];
    }
  }
  
  // 2. 按 key 排序
  const sortedKeys = Object.keys(filtered).sort();
  
  // 3. 拼接参数：key=value&key=value&key={merchant_key}
  const signStr = sortedKeys.map(k => `${k}=${filtered[k]}`).join('&') + `&key=${key}`;
  
  // 4. MD5 加密并转大写
  const hash = crypto.md5(signStr, 'hex');
  return hash.toUpperCase();
}

// 生成商户订单号
function generateOutOrderNo() {
  const timestamp = Date.now();
  const random = Math.floor(Math.random() * 10000);
  return `LOADTEST_${timestamp}_${random}`;
}

// 创建订单场景
function createOrder() {
  const outOrderNo = generateOutOrderNo();
  const amount = Math.floor(Math.random() * 100000) + 1000;

  const params = {
    mchId: parseInt(MERCHANT_ID),
    channelId: parseInt(CHANNEL_ID),
    mchOrderNo: outOrderNo,
    amount: amount,
    notifyUrl: 'https://example.com/notify',
    jumpUrl: 'https://example.com/jump',
    extra: '{}',
    compatible: 0,
    test: false, // 不使用测试模式，模拟真实下单
  };

  const sign = generateSign(params, MERCHANT_KEY);
  params.sign = sign;

  const payload = JSON.stringify(params);
  const startTime = Date.now();
  
  // 使用正常订单创建接口（通过mock插件模拟真实下单）
  const response = http.post(
    `${BASE_URL}/api/v1/orders`,
    payload,
    {
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': 'k6-loadtest/1.0',
      },
      tags: { name: 'CreateOrder' },
    }
  );

  const duration = Date.now() - startTime;
  orderCreationTime.add(duration);

  const success = check(response, {
    'create order status is 200': (r) => r.status === 200,
    'create order has data': (r) => {
      try {
        const body = JSON.parse(r.body);
        if (body.code === 200 && body.data && body.data.order_no) {
          // 保存订单号用于后续查询
          createdOrders.push(body.data.order_no);
          return true;
        }
        return false;
      } catch (e) {
        return false;
      }
    },
  });

  if (!success) {
    errorRate.add(1);
  }

  return response;
}

// 查询订单场景
function queryOrder(orderNo) {
  if (!orderNo) return;

  const startTime = Date.now();
  const response = http.get(
    `${BASE_URL}/api/v1/orders/${orderNo}`,
    {
      headers: {
        'User-Agent': 'k6-loadtest/1.0',
      },
      tags: { name: 'QueryOrder' },
    }
  );

  const duration = Date.now() - startTime;
  orderQueryTime.add(duration);

  const success = check(response, {
    'query order status is 200': (r) => r.status === 200,
    'query order has data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.code === 200 && body.data !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  if (!success) {
    errorRate.add(1);
  }
}

// 健康检查场景
function healthCheck() {
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
    'health check status is 200': (r) => r.status === 200,
  });

  if (!success) {
    errorRate.add(1);
  }
}

export default function () {
  // 70% 创建订单，20% 查询订单，10% 健康检查
  const rand = Math.random();
  
  if (rand < 0.7) {
  // 创建订单（使用正常接口，通过mock插件模拟真实下单）
  createOrder();
  } else if (rand < 0.9) {
    // 查询订单（如果有已创建的订单）
    if (createdOrders.length > 0) {
      const randomIndex = Math.floor(Math.random() * createdOrders.length);
      queryOrder(createdOrders[randomIndex]);
    } else {
      // 如果没有订单，创建订单
      createOrder();
    }
  } else {
    // 健康检查
    healthCheck();
  }

  sleep(Math.random() * 2 + 1);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'loadtest/results/mixed_scenario_summary.json': JSON.stringify(data, null, 2),
  };
}

function textSummary(data, options) {
  const indent = options.indent || '  ';
  let summary = '\n' + indent + '混合场景压测结果\n';
  summary += indent + '='.repeat(50) + '\n';
  
  if (data.metrics.http_req_duration) {
    summary += indent + `平均响应时间: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
    summary += indent + `P95响应时间: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
    summary += indent + `P99响应时间: ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms\n`;
  }
  
  if (data.metrics.order_creation_time) {
    summary += indent + `订单创建平均时间: ${data.metrics.order_creation_time.values.avg.toFixed(2)}ms\n`;
  }
  
  if (data.metrics.order_query_time) {
    summary += indent + `订单查询平均时间: ${data.metrics.order_query_time.values.avg.toFixed(2)}ms\n`;
  }
  
  if (data.metrics.health_check_time) {
    summary += indent + `健康检查平均时间: ${data.metrics.health_check_time.values.avg.toFixed(2)}ms\n`;
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
