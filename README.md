# Golang 支付系统核心 API

基于 Golang 开发的高并发支付系统核心 API 应用。

## 项目特性

- ✅ **高并发支持**: 使用 Gin 框架，支持高并发请求处理
- ✅ **工程化架构**: 清晰的分层架构，易于维护和扩展
- ✅ **数据库支持**: 集成 GORM，支持 MySQL 数据库
- ✅ **缓存支持**: 集成 Redis，提升性能
- ✅ **日志系统**: 使用 Zap 高性能日志库
- ✅ **配置管理**: 使用 Viper 管理配置，支持多环境
- ✅ **优雅关闭**: 支持优雅关闭服务器

## 项目结构

```
.
├── config/                 # 配置文件
│   ├── config.yaml        # 开发环境配置
│   ├── config.prod.yaml   # 生产环境配置（需从模板复制）
│   ├── config.prod.yaml.example  # 生产环境配置模板
│   ├── config.test.yaml   # 测试环境配置
│   └── config.go          # 配置结构定义
├── internal/              # 内部代码
│   ├── controller/        # 控制器层
│   │   └── order_controller.go
│   ├── database/          # 数据库连接
│   │   ├── mysql.go
│   │   └── redis.go
│   ├── logger/            # 日志模块
│   │   └── logger.go
│   ├── middleware/        # 中间件
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logger.go
│   │   └── recovery.go
│   ├── models/            # 数据模型
│   │   ├── merchant.go
│   │   ├── order.go
│   │   └── pay_channel.go
│   ├── router/            # 路由配置
│   │   └── router.go
│   ├── service/           # 业务逻辑层
│   │   ├── merchant_service.go
│   │   ├── order_service.go
│   │   └── pay_channel_service.go
│   ├── utils/             # 工具函数
│   │   ├── cache.go
│   │   └── id.go
│   └── response/          # 响应处理
│       └── response.go
├── sql/                   # SQL 文件
│   └── django-vue3-admin.sql
├── docs/                  # 文档
│   └── core.md
├── go.mod                 # Go 模块文件
├── go.sum                 # Go 依赖校验
├── main.go                # 应用入口
└── README.md              # 项目说明

```

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 5.7+
- Redis 5.0+ (可选)

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置数据库

#### 开发环境

编辑 `config/config.yaml` 文件，配置数据库连接信息：

```yaml
database:
  host: localhost
  port: 3306
  user: root
  password: your_password
  dbname: django_vue3_admin
```

#### 生产环境

1. 复制配置模板：

```bash
cp config/config.prod.yaml.example config/config.prod.yaml
```

2. 编辑 `config/config.prod.yaml`，配置生产环境参数（数据库、Redis 等）

3. 注意：`config.prod.yaml` 已在 `.gitignore` 中，不会被提交到版本控制

#### 测试环境

测试环境配置文件为 `config/config.test.yaml`，默认使用测试数据库和不同的端口。

### 4. 运行应用

#### 方式一：使用默认配置（开发环境）

```bash
go run main.go
```

#### 方式二：通过环境变量指定环境

```bash
# 开发环境（默认）
APP_ENV=dev go run main.go

# 测试环境
APP_ENV=test go run main.go

# 生产环境
APP_ENV=prod go run main.go
```

#### 方式三：通过命令行参数指定环境

```bash
# 开发环境
go run main.go dev

# 测试环境
go run main.go test

# 生产环境
go run main.go prod

# 或指定配置文件路径
go run main.go config/config.prod.yaml
```

#### 方式四：编译后运行

```bash
# 编译
go build -o bin/golang-pay-core main.go

# 运行（使用环境变量）
APP_ENV=prod ./bin/golang-pay-core

# 或使用命令行参数
./bin/golang-pay-core prod
```

**配置优先级：** 命令行参数 > 环境变量 APP_ENV > 默认（dev）

应用将在配置的端口启动（开发环境默认 8080，测试环境 8081）。

## API 文档

### 健康检查

```http
GET /health
```

### 创建订单

```http
POST /api/v1/orders
Content-Type: application/json

{
  "out_order_no": "ORD20240101001",
  "money": 10000,
  "tax": 100,
  "notify_url": "https://example.com/notify",
  "jump_url": "https://example.com/jump",
  "notify_money": 10000,
  "pay_channel_id": 1
}
```

### 查询订单

```http
GET /api/v1/orders/{order_no}
```

### 根据商户订单号查询

```http
GET /api/v1/orders/query?out_order_no=ORD20240101001
Authorization: Bearer {token}
```

## 架构说明

### 分层架构

1. **Controller 层**: 处理 HTTP 请求，参数验证，调用 Service
2. **Service 层**: 业务逻辑处理，事务管理
3. **Model 层**: 数据模型定义，数据库映射
4. **Database 层**: 数据库连接管理
5. **Middleware 层**: 中间件（认证、日志、跨域等）
6. **Utils 层**: 工具函数（缓存、ID 生成等）

### 高并发优化

1. **连接池**: 数据库和 Redis 使用连接池
2. **缓存**: 使用 Redis 缓存热点数据
3. **异步处理**: 支持异步任务处理（可扩展）
4. **优雅关闭**: 支持优雅关闭，避免请求丢失

### 工程化实践

1. **配置管理**: 使用 Viper 统一管理配置
2. **日志系统**: 结构化日志，支持文件输出和日志轮转
3. **错误处理**: 统一的错误处理和响应格式
4. **代码规范**: 清晰的分层和命名规范

## 开发指南

### 添加新的 API

1. 在 `internal/models/` 中定义数据模型
2. 在 `internal/service/` 中实现业务逻辑
3. 在 `internal/controller/` 中实现控制器
4. 在 `internal/router/router.go` 中注册路由

### 添加新的中间件

在 `internal/middleware/` 中创建新的中间件文件，然后在路由中注册。

## 配置说明

详细配置说明请参考 `config/config.yaml` 文件中的注释。

## 许可证

MIT License
