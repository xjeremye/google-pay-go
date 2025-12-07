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

# 调试运行（使用 Delve）
debug:
	@echo "调试运行应用..."
	@dlv debug main.go --headless --listen=:2345 --api-version=2 --accept-multiclient --check-go-version=false

# 调试运行（交互式）
debug-interactive:
	@echo "交互式调试运行应用..."
	@dlv debug main.go --check-go-version=false

# 附加到运行中的进程
debug-attach:
	@echo "附加到进程（需要先运行 make debug）..."
	@dlv attach --headless --listen=:2346 --api-version=2

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
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "检查 Go 版本..."
	@go version | grep -q "go1.2[2-9]\|go1.[3-9]" && \
		(echo "Go 版本 >= 1.22，安装最新 Delve..." && \
		 go install github.com/go-delve/delve/cmd/dlv@latest) || \
		(echo "Go 版本 < 1.22，安装兼容版本 Delve..." && \
		 go install github.com/go-delve/delve/cmd/dlv@v1.21.2)
	@echo "工具安装完成！"

# 安装兼容 Go 1.21 的 Delve
install-delve-compatible:
	@echo "安装兼容 Go 1.21 的 Delve 调试器..."
	@go install github.com/go-delve/delve/cmd/dlv@v1.21.2
	@echo "安装完成！"

# 生成 Swagger 文档
swagger:
	@echo "生成 Swagger 文档..."
	@go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs --parseDependency --parseInternal

# 生成 Swagger 文档并运行
swagger-run: swagger run

