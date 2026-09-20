#!/bin/bash

# 构建脚本 - 多语言打字学习终端工具
# Build script for Multi-language Typing Learning Terminal Tool

set -e

echo "开始构建 mllt-cli..."
echo "Building mllt-cli..."

# 清理之前的构建文件
if [ -f "mllt-cli" ]; then
    echo "清理之前的构建文件..."
    rm -f mllt-cli
fi

# 构建 macOS 版本
echo "构建 macOS 版本..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o mllt-cli-darwin-arm64 cmd/mllt-cli/main.go
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mllt-cli-darwin-amd64 cmd/mllt-cli/main.go

# 创建通用二进制文件（如果需要）
if command -v lipo &> /dev/null; then
    echo "创建通用二进制文件..."
    lipo -create -output mllt-cli mllt-cli-darwin-arm64 mllt-cli-darwin-amd64
    rm mllt-cli-darwin-arm64 mllt-cli-darwin-amd64
else
    # 如果没有 lipo，使用当前架构的版本
    if [ "$(uname -m)" = "arm64" ]; then
        mv mllt-cli-darwin-arm64 mllt-cli
        rm -f mllt-cli-darwin-amd64
    else
        mv mllt-cli-darwin-amd64 mllt-cli
        rm -f mllt-cli-darwin-arm64
    fi
fi

# 重新签名。lipo 合并后的通用二进制会丢掉顶层签名（单个切片的签名还在，
# 但整体验证不通过），在 Apple Silicon 上会被内核直接 SIGKILL。
if command -v codesign &> /dev/null; then
    echo "重新签名..."
    codesign --force --sign - mllt-cli
    codesign --verify mllt-cli && echo "签名校验通过"
fi

# 设置执行权限
chmod +x mllt-cli

echo "构建完成！"
echo "可执行文件: ./mllt-cli"
echo "文件大小: $(du -h mllt-cli | cut -f1)"
echo "架构信息: $(file mllt-cli)"
echo ""
echo "使用方法:"
echo "  ./mllt-cli          # 直接运行"
echo "  sudo cp mllt-cli /usr/local/bin/  # 安装到系统路径"