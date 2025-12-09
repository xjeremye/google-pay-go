import http from 'k6/http';
import { check } from 'k6';

// 测试数据配置
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8888';
const MERCHANT_ID = __ENV.MERCHANT_ID || '20001';      // 商户ID：20001
const CHANNEL_ID = __ENV.CHANNEL_ID || '8008';         // 支付通道ID：8008（使用alipay_mock插件）
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
  return `TEST_${timestamp}_${random}`;
}

// 只执行一次测试
export const options = {
  vus: 1,        // 1个虚拟用户
  iterations: 1, // 只执行1次
};

export default function () {
  console.log('='.repeat(60));
  console.log('开始测试订单创建接口');
  console.log(`BASE_URL: ${BASE_URL}`);
  console.log(`MERCHANT_ID: ${MERCHANT_ID}`);
  console.log(`CHANNEL_ID: ${CHANNEL_ID}`);
  console.log('='.repeat(60));

  // 生成测试数据
  const outOrderNo = generateOutOrderNo();
  const amount = 10000; // 100元（10000分）
  console.log(`订单号: ${outOrderNo}`);
  console.log(`金额: ${amount}分 (${amount / 100}元)`);

  // 构建请求参数（使用正式下单接口，商户ID: 20001, 通道ID: 8008）
  const params = {
    mchId: parseInt(MERCHANT_ID),      // 商户ID：20001
    channelId: parseInt(CHANNEL_ID),   // 支付通道ID：8008（使用alipay_mock插件）
    mchOrderNo: outOrderNo,
    amount: amount,
    notifyUrl: 'https://example.com/notify',
    jumpUrl: 'https://example.com/jump',
    extra: '{}',
    compatible: 0,
    test: false, // 不使用测试模式，使用正式下单接口（通过alipay_mock插件模拟，不调用真实支付宝API）
  };

  // 生成签名
  const sign = generateSign(params, MERCHANT_KEY);
  params.sign = sign;

  console.log('\n请求参数:');
  console.log(JSON.stringify(params, null, 2));

  // POST请求体
  const payload = JSON.stringify(params);

  // 发送POST请求（使用正式订单创建接口，通过alipay_mock插件模拟真实下单，不调用真实支付宝API）
  console.log(`\n发送请求到: ${BASE_URL}/api/v1/orders`);
  const startTime = Date.now();
  const response = http.post(
    `${BASE_URL}/api/v1/orders`,
    payload,
    {
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': 'k6-test/1.0',
      },
      tags: { name: 'TestSingleOrder' },
    }
  );

  const duration = Date.now() - startTime;

  console.log(`\n响应状态码: ${response.status}`);
  console.log(`响应时间: ${duration}ms`);
  console.log('\n响应内容:');
  console.log(response.body);

  // 检查响应
  const success = check(response, {
    'status is 200': (r) => r.status === 200,
    'response has data': (r) => {
      try {
        const body = JSON.parse(r.body);
        console.log('\n解析后的响应:');
        console.log(JSON.stringify(body, null, 2));
        return body.code === 200 && body.data !== undefined;
      } catch (e) {
        console.error('解析响应失败:', e);
        return false;
      }
    },
  });

  console.log('\n' + '='.repeat(60));
  if (success) {
    console.log('✅ 测试成功！');
    try {
      const body = JSON.parse(response.body);
      if (body.data && body.data.pay_url) {
        console.log(`支付URL: ${body.data.pay_url}`);
      }
      if (body.data && body.data.order_no) {
        console.log(`订单号: ${body.data.order_no}`);
      }
    } catch (e) {
      // 忽略解析错误
    }
  } else {
    console.log('❌ 测试失败！');
    console.log(`状态码: ${response.status}`);
    console.log(`响应: ${response.body}`);
  }
  console.log('='.repeat(60));
}
