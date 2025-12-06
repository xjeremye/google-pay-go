.PHONY: build run test clean deps

# 应用名称
APP_NAME=golang-pay-core

# 构建应用
build:
	@echo "构建应用..."
	@go build -o bin/$(APP_NAME) main.go

# 运行应用
run:
	@echo "运行应用..."
	@go run main.go

# 运行测试
test:
	@echo "运行测试..."
	@go test -v ./...

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf bin/
	@go clean

# 下载依赖
deps:
	@echo "下载依赖..."
	@go mod download
	@go mod tidy

# 格式化代码
fmt:
	@echo "格式化代码..."
	@go fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	@golangci-lint run || echo "请安装 golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# 安装依赖工具
install-tools:
	@echo "安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

