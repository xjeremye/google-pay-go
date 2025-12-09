-- wrk 压测脚本 - 健康检查
-- 使用方法: wrk -t4 -c100 -d30s -s health_check.lua http://localhost:8080

request = function()
  return wrk.format("GET", "/health")
end

response = function(status, headers, body)
  if status ~= 200 then
    print("Health check failed: " .. status)
  end
end

done = function(summary, latency, requests)
  io.write("健康检查压测完成\n")
  io.write(string.format("总请求数: %d\n", summary.requests.total))
  io.write(string.format("成功请求数: %d\n", summary.requests.completed))
  io.write(string.format("QPS: %.2f\n", summary.requests.completed / (summary.duration / 1000000)))
  io.write(string.format("平均延迟: %.2fms\n", latency.mean / 1000))
  io.write(string.format("P95延迟: %.2fms\n", latency:percentile(95) / 1000))
  io.write(string.format("P99延迟: %.2fms\n", latency:percentile(99) / 1000))
end
