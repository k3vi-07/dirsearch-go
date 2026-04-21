#!/bin/bash

# 简单测试脚本

echo "=== dirsearch-go 测试 ==="

# 1. 测试编译
echo "测试编译..."
make build
if [ $? -eq 0 ]; then
    echo "✓ 编译成功"
else
    echo "✗ 编译失败"
    exit 1
fi

# 2. 测试版本
echo ""
echo "测试版本信息..."
./build/dirsearch --version

# 3. 测试帮助
echo ""
echo "测试帮助信息..."
./build/dirsearch --help | head -5

# 4. 测试字典
echo ""
echo "测试字典文件..."
cat > db/test.txt << 'EOF'
admin
login
api
test
EOF

wc -l db/test.txt

echo ""
echo "=== 基础测试完成 ==="
