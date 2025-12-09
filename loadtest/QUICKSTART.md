# 压测快速开始指南

## 5分钟快速上手

### 1. 安装压测工具

**安装 k6:**
```bash
# macOS
brew install k6

# Linux (Ubuntu/Debian)
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**安装 wrk (可选):**
```bash
# macOS
brew install wrk

# Linux
sudo apt-get install wrk
```

### 2. 配置压测参数

```bash
cd loadtest
cp config.env.example config.env
# 编辑 config.env，设置你的服务器地址和商户密钥
```

### 3. 运行压测

**方式一：使用 Makefile（推荐）**
```bash
# 从项目根目录运行
make loadtest-order      # 创建订单压测
make loadtest-health     # 健康检查压测
make loadtest-mixed      # 混合场景压测
```

**方式二：直接运行脚本**
```bash
cd loadtest
./run_k6.sh create_order http://localhost:8080
```

### 4. 查看结果

压测结果会自动保存到 `loadtest/results/` 目录，并显示在终端。

查看详细结果：
```bash
make loadtest-analyze
# 或
cd loadtest && ./analyze_results.sh
```

## 常见场景

### 场景1：快速验证服务器性能
```bash
make loadtest-health
```
这个场景会快速测试健康检查接口，验证服务器基础性能。

### 场景2：测试订单创建接口
```bash
make loadtest-order
```
这个场景会模拟真实的订单创建流程，包括签名生成、参数验证等。

### 场景3：混合场景测试
```bash
make loadtest-mixed
```
这个场景会混合多种操作（70%创建订单，20%查询订单，10%健康检查），更接近真实业务场景。

## 压测参数调整

### 修改并发数

编辑 `loadtest/k6/create_order.js`，修改 `stages` 配置：

```javascript
export const options = {
  stages: [
    { duration: '30s', target: 50 },   // 修改这里的 target 值
    { duration: '1m', target: 100 },
    // ...
  ],
};
```

### 修改持续时间

同样在 `stages` 配置中修改 `duration` 值。

### 修改阈值

修改 `thresholds` 配置：

```javascript
thresholds: {
  http_req_duration: ['p(95)<500', 'p(99)<1000'], // 修改响应时间阈值
  http_req_failed: ['rate<0.01'],                  // 修改错误率阈值
},
```

## 注意事项

1. **首次运行前**：确保已配置 `config.env` 文件
2. **签名密钥**：确保 `MERCHANT_KEY` 与后端配置一致
3. **测试环境**：建议在独立的测试环境进行压测
4. **资源监控**：压测时注意监控服务器CPU、内存、数据库连接数等

## 获取帮助

- 详细文档：查看 `loadtest/README.md`
- 问题反馈：查看压测结果中的错误信息
