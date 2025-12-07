#!/bin/bash
# 安装兼容 Go 1.21 的 Delve 版本

echo "正在安装兼容 Go 1.21 的 Delve 调试器..."

# 安装兼容 Go 1.21 的 Delve 版本
go install github.com/go-delve/delve/cmd/dlv@v1.21.2

echo "安装完成！"
echo "验证安装："
dlv version

