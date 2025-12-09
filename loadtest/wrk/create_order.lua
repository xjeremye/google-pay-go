-- wrk 压测脚本 - 创建订单（GET方式）
-- 使用方法: wrk -t4 -c100 -d30s -s create_order.lua http://localhost:8080

-- 初始化
init = function(args)
  -- 从环境变量或参数获取配置
  local merchant_id = os.getenv("MERCHANT_ID") or "1"
  local channel_id = os.getenv("CHANNEL_ID") or "1"
  local merchant_key = os.getenv("MERCHANT_KEY") or "your_merchant_key"
  
  -- 存储配置到全局变量
  wrk.merchant_id = merchant_id
  wrk.channel_id = channel_id
  wrk.merchant_key = merchant_key
  
  -- 计数器
  wrk.request_count = 0
end

-- 生成商户订单号
function generate_out_order_no()
  local timestamp = os.time() * 1000
  local random = math.random(1000, 9999)
  return string.format("LOADTEST_%d_%d", timestamp, random)
end

-- 生成签名（简化版，需要根据实际签名算法调整）
function generate_sign(params, key)
  -- 这里需要实现实际的签名算法
  -- 示例：简单的MD5签名
  local sorted_keys = {}
  for k in pairs(params) do
    table.insert(sorted_keys, k)
  end
  table.sort(sorted_keys)
  
  local sign_str = ""
  for i, k in ipairs(sorted_keys) do
    if i > 1 then
      sign_str = sign_str .. "&"
    end
    sign_str = sign_str .. k .. "=" .. tostring(params[k])
  end
  sign_str = sign_str .. "&key=" .. key
  
  -- 这里应该调用MD5，但wrk没有内置MD5，所以简化处理
  -- 实际使用时需要实现正确的签名算法
  return "sign_" .. sign_str
end

-- 请求生成函数
request = function()
  wrk.request_count = wrk.request_count + 1
  
  -- 生成测试数据
  local out_order_no = generate_out_order_no()
  local amount = math.random(1000, 100000)
  
  -- 构建参数
  local params = {
    mchId = wrk.merchant_id,
    channelId = wrk.channel_id,
    mchOrderNo = out_order_no,
    amount = amount,
    notifyUrl = "https://example.com/notify",
    jumpUrl = "https://example.com/jump",
    extra = "{}",
    compatible = 0,
    test = true
  }
  
  -- 生成签名
  local sign = generate_sign(params, wrk.merchant_key)
  params.sign = sign
  
  -- 构建查询字符串
  local query_string = ""
  local first = true
  for k, v in pairs(params) do
    if not first then
      query_string = query_string .. "&"
    end
    query_string = query_string .. k .. "=" .. tostring(v)
    first = false
  end
  
  -- 构建完整URL
  local url = "/api/v1/orders?" .. query_string
  
  return wrk.format("GET", url)
end

-- 响应处理函数
response = function(status, headers, body)
  -- 可以在这里记录响应信息
  if status ~= 200 then
    -- 记录错误
    print("Error: " .. status .. " - " .. body:sub(1, 100))
  end
end

-- 完成函数
done = function(summary, latency, requests)
  io.write("压测完成\n")
  io.write(string.format("总请求数: %d\n", summary.requests.total))
  io.write(string.format("成功请求数: %d\n", summary.requests.completed))
  io.write(string.format("失败请求数: %d\n", summary.requests.completed - summary.requests.completed))
  io.write(string.format("QPS: %.2f\n", summary.requests.completed / (summary.duration / 1000000)))
  io.write(string.format("平均延迟: %.2fms\n", latency.mean / 1000))
  io.write(string.format("P99延迟: %.2fms\n", latency:percentile(99) / 1000))
end
