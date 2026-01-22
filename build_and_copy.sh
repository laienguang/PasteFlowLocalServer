#!/bin/bash

# 设置错误时退出
set -e

# 添加 Go 到 PATH
export PATH=$PATH:/usr/local/go/bin

# 获取脚本所在目录的绝对路径
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "🚀 开始构建流程..."

# 1. 编译 Go 项目
echo "📦 正在编译 local-server..."
# 显式指定输出为当前目录下的 local-server
go build -o local-server .

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi
echo "✅ 编译成功"

# 2. 定义目标路径
# 相对路径 ../PasteFlowMac/PasteFlowAgent
TARGET_DIR="../PasteFlowMac/PasteFlowAgent"
TARGET_FILE="$TARGET_DIR/local-server"

# 确保目标目录存在
if [ ! -d "$TARGET_DIR" ]; then
    echo "📂 创建目标目录: $TARGET_DIR"
    mkdir -p "$TARGET_DIR"
fi

# 3. 拷贝文件
echo "🚚 正在拷贝到 $TARGET_DIR ..."
cp local-server "$TARGET_DIR/"

if [ $? -eq 0 ]; then
    echo "✅ 拷贝成功"
    
    # 4. 设置权限
    echo "🔑 设置执行权限..."
    chmod +x "$TARGET_FILE"
    
    echo "🎉 构建并部署完成！"
    echo "📍 文件位置: $TARGET_FILE"
else
    echo "❌ 拷贝失败"
    exit 1
fi
