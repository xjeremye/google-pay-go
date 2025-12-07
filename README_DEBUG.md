# 调试指南

本文档介绍如何调试运行 Go 支付系统应用。

## 方式一：使用 VS Code 调试

### 1. 安装 Go 扩展

在 VS Code 中安装官方 Go 扩展：
- 扩展 ID: `golang.go`

### 2. 安装 Delve 调试器

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

### 3. 开始调试

1. 打开 VS Code
2. 按 `F5` 或点击左侧调试面板
3. 选择 "Debug Go Application" 配置
4. 设置断点（点击行号左侧）
5. 开始调试

### 4. 调试配置说明

- **Debug Go Application**: 使用开发环境配置调试
- **Debug Go Application (Prod)**: 使用生产环境配置调试
- **Attach to Process**: 附加到已运行的进程

## 方式二：使用命令行调试（Delve）

### 1. 安装 Delve

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

### 2. 交互式调试

```bash
make debug-interactive
```

或者直接使用：

```bash
dlv debug main.go
```

### 3. 常用调试命令

在 Delve 交互式界面中：

```
(dlv) break main.main          # 在 main 函数设置断点
(dlv) break internal/service/order_create.go:180  # 在指定文件行设置断点
(dlv) continue                 # 继续执行
(dlv) next                     # 下一行
(dlv) step                     # 进入函数
(dlv) print variable           # 打印变量
(dlv) locals                   # 显示局部变量
(dlv) args                     # 显示函数参数
(dlv) stack                    # 显示调用栈
(dlv) goroutines               # 显示所有 goroutines
(dlv) exit                     # 退出调试
```

### 4. 无头模式调试（用于远程调试）

```bash
make debug
```

然后在另一个终端附加调试器：

```bash
dlv connect localhost:2345
```

## 方式三：使用 GoLand/IntelliJ IDEA

### 1. 创建运行配置

1. 打开 "Run" -> "Edit Configurations..."
2. 点击 "+" -> "Go Build"
3. 配置：
   - **Name**: Debug Application
   - **Run kind**: Package
   - **Package path**: `github.com/golang-pay-core`
   - **Working directory**: 项目根目录
   - **Environment variables**: `APP_ENV=dev`

### 2. 开始调试

1. 设置断点
2. 点击调试按钮（或按 `Shift+F9`）

## 方式四：使用日志调试

### 1. 启用详细日志

在 `config/config.yaml` 中设置：

```yaml
logger:
  level: debug  # 或 trace
```

### 2. 添加调试日志

在代码中添加：

```go
logger.Logger.Debug("调试信息",
    zap.String("key", "value"),
    zap.Int("number", 123),
)
```

## 调试技巧

### 1. 条件断点

在 VS Code 中：
- 右键断点 -> "Edit Breakpoint"
- 设置条件，例如：`orderCtx.Money > 10000`

### 2. 日志断点

在 VS Code 中：
- 右键断点 -> "Edit Breakpoint"
- 选择 "Logpoint"
- 输入日志表达式，例如：`订单金额: {orderCtx.Money}`

### 3. 查看变量

- **VS Code**: 在调试面板的 "Variables" 区域查看
- **Delve**: 使用 `print variable` 命令
- **GoLand**: 在 "Variables" 面板查看

### 4. 调用栈

- **VS Code**: 在调试面板的 "Call Stack" 区域查看
- **Delve**: 使用 `stack` 命令
- **GoLand**: 在 "Frames" 面板查看

### 5. 调试 HTTP 请求

可以使用以下工具测试 API：

```bash
# 使用 curl
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{"mchOrderNo":"test123","mchId":1,"channelId":1,"amount":10000}'

# 使用 httpie
http POST localhost:8080/api/v1/orders \
  mchOrderNo=test123 mchId=1 channelId=1 amount=10000
```

## 常见问题

### 1. Go 版本不兼容

**错误信息**：`Go version go1.21.x is too old for this version of Delve (minimum supported version 1.22)`

**解决方案**：

#### 方案一：升级 Go 版本（推荐）

```bash
# 使用 Homebrew 升级（macOS）
brew upgrade go

# 或下载最新版本
# https://go.dev/dl/
```

#### 方案二：安装兼容的 Delve 版本

```bash
# 安装兼容 Go 1.21 的 Delve 版本
go install github.com/go-delve/delve/cmd/dlv@v1.21.0
```

#### 方案三：临时禁用版本检查（不推荐）

已在配置文件中添加 `--check-go-version=false` 标志，但建议升级 Go 版本。

### 2. Delve 安装失败

```bash
# 确保 Go 版本 >= 1.16
go version

# 如果仍然失败，尝试：
go get -u github.com/go-delve/delve/cmd/dlv
```

### 2. 无法附加到进程

确保进程是以调试模式启动的：

```bash
make debug
```

### 3. 断点不生效

- 确保代码已编译（不是缓存的旧版本）
- 检查断点是否在可执行代码行（不是注释或空行）
- 在 VS Code 中，确保使用的是 "Debug Go Application" 配置

### 4. 调试时性能慢

- 减少断点数量
- 使用条件断点而不是多个普通断点
- 在发布模式下编译（`go build -ldflags="-s -w"`）会移除调试信息

## 性能分析

### 使用 pprof

```bash
# 在代码中导入
import _ "net/http/pprof"

# 访问性能分析端点
go tool pprof http://localhost:8080/debug/pprof/profile
```

## 参考资料

- [Delve 官方文档](https://github.com/go-delve/delve)
- [VS Code Go 调试指南](https://github.com/golang/vscode-go/wiki/debugging)
- [GoLand 调试指南](https://www.jetbrains.com/help/go/debugging-code.html)

