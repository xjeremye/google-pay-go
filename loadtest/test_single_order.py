#!/usr/bin/env python3
"""
单次订单创建测试脚本
使用方法: python3 test_single_order.py [base_url] [merchant_id] [channel_id] [merchant_key]
"""

import sys
import json
import hashlib
import time
import random
import requests
from urllib.parse import urlencode

# 配置
BASE_URL = sys.argv[1] if len(sys.argv) > 1 else 'http://localhost:8888'
MERCHANT_ID = sys.argv[2] if len(sys.argv) > 2 else '20001'
CHANNEL_ID = sys.argv[3] if len(sys.argv) > 3 else '8008'
MERCHANT_KEY = sys.argv[4] if len(sys.argv) > 4 else 'your_merchant_key'

def format_value(v):
    """格式化值（与 Go 的 fmt.Sprintf("%v", v) 一致）"""
    if v is None:
        return ''
    if isinstance(v, bool):
        return 'false' if not v else 'true'  # Go 的 %v 对 bool 输出 false/true
    if isinstance(v, (int, float)):
        return str(v)
    return str(v)

def generate_sign(params, key):
    """生成签名（与后端一致的签名算法）"""
    # 1. 移除 sign 字段，过滤掉 null 值
    filtered = {k: v for k, v in params.items() if k != 'sign' and v is not None}
    
    # 2. 按 key 排序
    sorted_keys = sorted(filtered.keys())
    
    # 3. 拼接参数：key=value&key=value&key={merchant_key}
    # 使用 format_value 确保值与 Go 的 %v 格式化一致
    sign_parts = []
    for k in sorted_keys:
        sign_parts.append(f'{k}={format_value(filtered[k])}')
    sign_parts.append(f'key={key}')
    sign_str = '&'.join(sign_parts)
    
    # 调试：打印签名字符串
    # 4. MD5 加密并转大写
    hash_obj = hashlib.md5(sign_str.encode('utf-8'))
    sign_result = hash_obj.hexdigest().upper()
    
    return sign_result

def generate_out_order_no():
    """生成商户订单号"""
    timestamp = int(time.time() * 1000)
    random_num = random.randint(0, 9999)
    return f'TEST_{timestamp}_{random_num}'

def main():
    print('=' * 60)
    print('开始测试订单创建接口')
    print('=' * 60)
    print(f'BASE_URL: {BASE_URL}')
    print(f'MERCHANT_ID: {MERCHANT_ID}')
    print(f'CHANNEL_ID: {CHANNEL_ID}')
    print('=' * 60)
    
    # 生成测试数据
    out_order_no = generate_out_order_no()
    amount = 10000  # 100元（10000分）
    print(f'\n订单号: {out_order_no}')
    print(f'金额: {amount}分 ({amount / 100}元)')
    
    # 构建请求参数
    params = {
        'mchId': int(MERCHANT_ID),
        'channelId': int(CHANNEL_ID),
        'mchOrderNo': out_order_no,
        'amount': amount,
        'notifyUrl': 'https://example.com/notify',
        'jumpUrl': 'https://example.com/jump',
        'extra': '{}',
        'compatible': 0,
        'test': False,
    }
    
    # 生成签名
    sign = generate_sign(params, MERCHANT_KEY)
    params['sign'] = sign
    
    print('\n请求参数:')
    print(json.dumps(params, indent=2, ensure_ascii=False))
    
    # 打印签名字符串用于调试
    temp_params = {k: v for k, v in params.items() if k != 'sign'}
    sorted_keys = sorted(temp_params.keys())
    sign_parts = [f'{k}={format_value(temp_params[k])}' for k in sorted_keys]
    sign_parts.append(f'key={MERCHANT_KEY}')
    sign_str = '&'.join(sign_parts)
    print(f'\n签名字符串: {sign_str}')
    print(f'生成的签名: {params["sign"]}')
    if MERCHANT_KEY == 'your_merchant_key':
        print(f'\n⚠️  警告: 使用的是默认密钥 "{MERCHANT_KEY}"')
        print('   签名验证会失败，请提供正确的商户密钥！')
        print('   使用方法: python3 test_single_order.py [base_url] [merchant_id] [channel_id] [merchant_key]')
        print('   或查询数据库获取商户密钥: SELECT sign_key FROM dvadmin_merchant WHERE id = 20001;')
    
    # 发送POST请求
    url = f'{BASE_URL}/api/v1/orders'
    print(f'\n发送请求到: {url}')
    
    start_time = time.time()
    try:
        response = requests.post(
            url,
            json=params,
            headers={
                'Content-Type': 'application/json',
                'User-Agent': 'python-test/1.0',
            },
            timeout=10
        )
        duration = (time.time() - start_time) * 1000
        
        print(f'\n响应状态码: {response.status_code}')
        print(f'响应时间: {duration:.2f}ms')
        print('\n响应内容:')
        print(response.text)
        
        # 解析响应
        try:
            body = response.json()
            print('\n解析后的响应:')
            print(json.dumps(body, indent=2, ensure_ascii=False))
            
            print('\n' + '=' * 60)
            if response.status_code == 200 and body.get('code') == 200:
                print('✅ 测试成功！')
                if 'data' in body:
                    data = body['data']
                    if 'pay_url' in data:
                        print(f'支付URL: {data["pay_url"]}')
                    if 'order_no' in data:
                        print(f'订单号: {data["order_no"]}')
            else:
                print('❌ 测试失败！')
                print(f'错误码: {body.get("code")}')
                print(f'错误信息: {body.get("message", "未知错误")}')
        except json.JSONDecodeError:
            print('\n' + '=' * 60)
            print('❌ 响应不是有效的JSON格式')
            print('=' * 60)
            
    except requests.exceptions.RequestException as e:
        print('\n' + '=' * 60)
        print(f'❌ 请求失败: {e}')
        print('=' * 60)
        sys.exit(1)

if __name__ == '__main__':
    main()
