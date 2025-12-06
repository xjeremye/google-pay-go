# API 使用说明

## 创建订单接口

创建订单接口支持三种请求方式：

### 1. POST JSON（推荐）

**请求方式：** `POST /api/v1/orders`

**Content-Type：** `application/json`

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "mchId": 1,
    "channelId": 1,
    "mchOrderNo": "ORD20240101001",
    "amount": 10000,
    "notifyUrl": "https://example.com/notify",
    "jumpUrl": "https://example.com/jump",
    "extra": "{}",
    "compatible": 0,
    "test": false,
    "sign": "ABC123..."
  }'
```

### 2. POST Form

**请求方式：** `POST /api/v1/orders`

**Content-Type：** `application/x-www-form-urlencoded` 或 `multipart/form-data`

**请求示例：**

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "mchId=1" \
  -d "channelId=1" \
  -d "mchOrderNo=ORD20240101001" \
  -d "amount=10000" \
  -d "notifyUrl=https://example.com/notify" \
  -d "jumpUrl=https://example.com/jump" \
  -d "extra={}" \
  -d "compatible=0" \
  -d "test=false" \
  -d "sign=ABC123..."
```

### 3. GET Query String

**请求方式：** `GET /api/v1/orders`

**请求示例：**

```bash
curl -X GET "http://localhost:8080/api/v1/orders?mchId=1&channelId=1&mchOrderNo=ORD20240101001&amount=10000&notifyUrl=https://example.com/notify&jumpUrl=https://example.com/jump&extra={}&compatible=0&test=false&sign=ABC123..."
```

**浏览器访问示例：**

```
http://localhost:8080/api/v1/orders?mchId=1&channelId=1&mchOrderNo=ORD20240101001&amount=10000&notifyUrl=https://example.com/notify&sign=ABC123...
```

## 请求参数说明

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| mchId | int | 是 | 商户ID |
| channelId | int | 是 | 渠道ID |
| mchOrderNo | string | 是 | 商户订单号 |
| amount | int | 是 | 金额（单位：分） |
| notifyUrl | string | 是 | 支付结果通知地址 |
| jumpUrl | string | 否 | 支付完成后跳转地址 |
| extra | string | 否 | 额外参数（JSON 字符串） |
| compatible | int | 否 | 兼容模式（0=标准模式，1=兼容模式） |
| test | bool | 否 | 测试模式 |
| sign | string | 是 | 签名 |

## 响应格式

### 标准模式（compatible=0）

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "mchOrderNo": "ORD20240101001",
    "payOrderId": "PAY20240101120000001",
    "payUrl": "https://pay.example.com/pay?order_no=PAY20240101120000001",
    "sign": "ABC123..."
  }
}
```

### 兼容模式（compatible=1）

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "trade_no": "PAY20240101120000001",
    "payurl": "https://pay.example.com/pay?order_no=PAY20240101120000001",
    "msg": "订单创建成功",
    "code": 1
  }
}
```

## 错误响应

```json
{
  "code": 7301,
  "message": "商户不存在",
  "data": null
}
```

## 注意事项

1. **签名验证**：所有请求都必须包含有效的签名（sign 参数）
2. **参数顺序**：GET 请求时，参数顺序不影响签名验证
3. **URL 编码**：GET 请求时，URL 中的特殊字符需要编码
4. **Content-Type**：POST 请求时，请确保设置正确的 Content-Type
5. **幂等性**：相同的商户订单号（mchOrderNo）只能创建一次订单

## 签名生成

签名生成方法请参考 `docs/order/sign.go` 中的实现。

### 标准模式签名

1. 按 key 排序所有参数（排除 sign）
2. 拼接成 `key=value&key=value` 格式
3. 最后加上 `&key={商户密钥}`
4. MD5 加密并转大写

### 兼容模式签名

1. 过滤掉 sign、sign_type 和空值
2. 按 key 排序
3. 拼接成 `key=value&key=value` 格式
4. 最后加上商户密钥
5. MD5 加密并转大写

