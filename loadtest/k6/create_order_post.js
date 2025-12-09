import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// 自定义指标
const errorRate = new Rate('errors');
const orderCreationTime = new Trend('order_creation_time');
const orderCreationSuccess = new Counter('order_creation_success');
const orderCreationFailure = new Counter('order_creation_failure');

// 配置
export const options = {
  stages: [
    { duration: '30s', target: 50 },   // 预热：30秒内逐步增加到50个并发
    { duration: '1m', target: 100 },   // 稳定：1分钟内保持100个并发
    { duration: '2m', target: 200 },   // 增长：2分钟内增加到200个并发
    { duration: '2m', target: 300 },   // 峰值：2分钟内增加到300个并发
    { duration: '1m', target: 200 },   // 下降：1分钟内降到200个并发
    { duration: '30s', target: 0 },    // 冷却：30秒内降到0
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

export default function () {
  // 生成测试数据
  const outOrderNo = generateOutOrderNo();
  const amount = Math.floor(Math.random() * 100000) + 1000;
  const timestamp = Math.floor(Date.now() / 1000);

  // 构建请求参数
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

  // 生成签名
  const sign = generateSign(params, MERCHANT_KEY);
  params.sign = sign;

  // POST请求体
  const payload = JSON.stringify(params);

  // 发送POST请求（使用正常订单创建接口，通过mock插件模拟真实下单）
  const startTime = Date.now();
  const response = http.post(
    `${BASE_URL}/api/v1/orders`,
    payload,
    {
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': 'k6-loadtest/1.0',
      },
      tags: { name: 'CreateOrderPOST' },
    }
  );

  const duration = Date.now() - startTime;
  orderCreationTime.add(duration);

  // 检查响应
  const success = check(response, {
    'status is 200': (r) => r.status === 200,
    'response has data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.code === 200 && body.data !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  if (success) {
    orderCreationSuccess.add(1);
  } else {
    orderCreationFailure.add(1);
    errorRate.add(1);
    console.error(`Order creation failed: ${response.status} - ${response.body}`);
  }

  // 思考时间
  sleep(Math.random() * 2 + 1);
}

export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'loadtest/results/create_order_post_summary.json': JSON.stringify(data, null, 2),
  };
}

function textSummary(data, options) {
  const indent = options.indent || '  ';
  let summary = '\n' + indent + '压测结果摘要 (POST方式)\n';
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
  
  if (data.metrics.order_creation_success) {
    summary += indent + `成功创建订单数: ${data.metrics.order_creation_success.values.count}\n`;
  }
  
  if (data.metrics.order_creation_failure) {
    summary += indent + `失败订单数: ${data.metrics.order_creation_failure.values.count}\n`;
  }
  
  summary += indent + '='.repeat(50) + '\n';
  return summary;
}
