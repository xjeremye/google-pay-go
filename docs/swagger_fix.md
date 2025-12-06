# Swagger 文档生成和使用

## 问题解决

如果遇到 "Fetch error Internal Server Error doc.json" 错误，通常是因为 Swagger 文档还没有生成。

## 解决方法

### 1. 生成 Swagger 文档

```bash
# 方式一：使用 Makefile（推荐）
make swagger

# 方式二：直接使用 go run
go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs --parseDependency --parseInternal
```

### 2. 验证生成的文件

生成成功后，`docs` 目录下应该有以下文件：
- `docs.go` - 生成的文档代码
- `swagger.json` - JSON 格式的文档
- `swagger.yaml` - YAML 格式的文档

### 3. 重新编译和运行

```bash
# 重新编译
go build -o bin/golang-pay-core main.go

# 运行应用
./bin/golang-pay-core
```

### 4. 访问 Swagger UI

启动应用后，访问：
```
http://localhost:8080/swagger/index.html
```

## 常见问题

### Q: 仍然显示 "Fetch error"？

A: 检查以下几点：
1. 确保已运行 `make swagger` 生成文档
2. 确保 `docs/swagger.json` 文件存在
3. 重新编译并启动应用
4. 检查浏览器控制台是否有其他错误

### Q: Swagger UI 显示空白？

A: 
1. 检查 `docs/swagger.json` 文件内容是否完整
2. 确保所有控制器方法都有正确的 Swagger 注释
3. 检查 `main.go` 中是否正确导入了 `docs` 包

### Q: 如何更新 Swagger 文档？

A: 
1. 修改代码中的 Swagger 注释
2. 重新运行 `make swagger`
3. 重新编译并启动应用

## 注意事项

- Swagger 文档文件（`swagger.json`, `swagger.yaml`, `docs.go`）是自动生成的，不要手动编辑
- 每次修改 Swagger 注释后，都需要重新生成文档
- 建议将 `docs/swagger.json` 和 `docs/swagger.yaml` 添加到 `.gitignore`（可选）

