# 多环境配置说明

## 环境配置概述

项目支持多环境配置，包括：
- **开发环境** (dev): `config/config.yaml`
- **测试环境** (test): `config/config.test.yaml`
- **生产环境** (prod): `config/config.prod.yaml`

## 配置文件说明

### 开发环境配置 (`config/config.yaml`)

- **模式**: debug
- **端口**: 8080
- **日志级别**: info
- **日志格式**: text
- **SQL 日志**: 开启
- **数据库连接池**: 较小（适合开发）

### 测试环境配置 (`config/config.test.yaml`)

- **模式**: test
- **端口**: 8081（避免与开发环境冲突）
- **日志级别**: debug
- **日志格式**: text
- **SQL 日志**: 开启
- **数据库**: 使用测试数据库 `django_vue3_admin_test`
- **Redis DB**: 使用 DB 1（避免与开发环境冲突）
- **数据库连接池**: 中等

### 生产环境配置 (`config/config.prod.yaml`)

- **模式**: release
- **端口**: 8080
- **日志级别**: info
- **日志格式**: json（便于日志分析）
- **日志输出**: 文件
- **SQL 日志**: 关闭（提升性能）
- **数据库连接池**: 大（支持高并发）
- **Redis 连接池**: 大（支持高并发）

## 使用方法

### 1. 通过环境变量

```bash
# 开发环境（默认）
APP_ENV=dev go run main.go

# 测试环境
APP_ENV=test go run main.go

# 生产环境
APP_ENV=prod go run main.go
```

### 2. 通过命令行参数

```bash
# 开发环境
go run main.go dev

# 测试环境
go run main.go test

# 生产环境
go run main.go prod

# 指定配置文件路径
go run main.go config/config.prod.yaml
```

### 3. 编译后运行

```bash
# 编译
go build -o bin/golang-pay-core main.go

# 使用环境变量
APP_ENV=prod ./bin/golang-pay-core

# 使用命令行参数
./bin/golang-pay-core prod
```

## 配置优先级

配置加载优先级（从高到低）：
1. **命令行参数** - 直接指定配置文件路径或环境名称
2. **环境变量 APP_ENV** - 自动选择对应环境的配置文件
3. **默认值** - 使用开发环境配置 (`config/config.yaml`)

## 环境变量覆盖

配置系统支持通过环境变量覆盖配置值，环境变量命名规则：

- 使用 `APP_` 前缀
- 配置项使用下划线分隔，全大写
- 例如：`APP_DATABASE_HOST`、`APP_DATABASE_PORT`

示例：
```bash
# 覆盖数据库主机
APP_DATABASE_HOST=192.168.1.100 go run main.go

# 覆盖端口
APP_APP_PORT=9090 go run main.go
```

## 生产环境配置注意事项

### 1. 安全配置

- **数据库密码**: 建议使用环境变量或密钥管理服务
- **Redis 密码**: 生产环境必须设置密码

### 2. 性能优化

- **数据库连接池**: 根据实际负载调整 `max_open_conns`
- **Redis 连接池**: 根据实际负载调整 `pool_size`
- **日志级别**: 生产环境使用 `info`，避免过多日志影响性能

### 3. 配置文件管理

- `config.prod.yaml` 已在 `.gitignore` 中，不会被提交到版本控制
- 生产环境配置应通过配置管理工具或环境变量管理
- 建议使用 `config.prod.yaml.example` 作为模板

## Docker 部署示例

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o golang-pay-core main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/golang-pay-core .
COPY --from=builder /app/config ./config
CMD ["./golang-pay-core", "prod"]
```

```bash
# 使用环境变量
docker run -e APP_ENV=prod golang-pay-core

# 挂载配置文件
docker run -v $(pwd)/config:/app/config golang-pay-core prod
```

## Kubernetes 部署示例

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: golang-pay-core
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: golang-pay-core
        image: golang-pay-core:latest
        env:
        - name: APP_ENV
          value: "prod"
        - name: APP_DATABASE_HOST
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: host
        - name: APP_DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: password
        command: ["./golang-pay-core", "prod"]
```

## 配置验证

启动应用时，日志会显示当前使用的配置文件：

```
服务器启动 address=:8080 mode=release
```

可以通过健康检查接口验证配置：

```bash
curl http://localhost:8080/health
```

## 常见问题

### Q: 如何知道当前使用的是哪个配置文件？

A: 查看启动日志，会显示配置的 mode 和 port，或者查看健康检查接口返回的信息。

### Q: 生产环境配置文件应该放在哪里？

A: 
1. 使用配置管理工具（如 Consul、Vault）
2. 使用 Kubernetes ConfigMap/Secret
3. 使用环境变量
4. 如果使用文件，确保文件权限正确，不要提交到版本控制

### Q: 如何在不同环境使用不同的数据库？

A: 在对应的配置文件中修改 `database.dbname` 字段即可。

### Q: 测试环境和开发环境如何避免冲突？

A: 测试环境使用不同的端口（8081）和 Redis DB（1），确保不会相互影响。

