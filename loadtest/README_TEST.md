# 单次订单创建测试

## 快速测试

### 使用 Python 脚本（推荐）

```bash
# 使用默认配置（需要提供正确的商户密钥）
python3 loadtest/test_single_order.py

# 指定参数
python3 loadtest/test_single_order.py http://localhost:8888 20001 8008 your_merchant_key
```

### 使用 k6 脚本

```bash
# 设置环境变量
export BASE_URL=http://localhost:8888
export MERCHANT_ID=20001
export CHANNEL_ID=8008
export MERCHANT_KEY=your_merchant_key

# 运行测试
k6 run loadtest/k6/test_single_order.js
```

## 注意事项

1. **商户密钥**: 必须提供正确的商户密钥，否则签名验证会失败
2. **商户ID**: 默认使用 20001
3. **通道ID**: 默认使用 8008（需要配置为使用 alipay_mock 插件）
4. **应用模式**: 如果应用在 debug 模式，后端会输出正确的签名用于调试

## 获取商户密钥

商户密钥存储在数据库中，需要查询 `dvadmin_merchant` 表的 `sign_key` 字段。
