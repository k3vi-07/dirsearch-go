#!/bin/bash

# dirsearch-go 集成测试脚本

set -e

echo "=== dirsearch-go 集成测试 ==="
echo ""

# 1. 编译项目
echo "[1/5] 编译项目..."
make build
echo "✓ 编译完成"
echo ""

# 2. 创建测试数据
echo "[2/5] 创建测试数据..."
mkdir -p db wordlists
cat > db/common.txt << 'EOF'
admin
administrator
login
panel
dashboard
api
test
config
backup
upload
files
images
css
js
assets
user
users
EOF
echo "✓ 测试字典创建完成"
echo ""

# 3. 启动测试服务器
echo "[3/5] 启动测试服务器..."
go run test_server.go > /dev/null 2>&1 &
SERVER_PID=$!
sleep 2
echo "✓ 测试服务器启动 (PID: $SERVER_PID)"
echo ""

# 4. 运行扫描测试
echo "[4/5] 运行扫描测试..."
./build/dirsearch \
	-u http://localhost:8080 \
	-w db/common.txt \
	-t 10 \
	--include-status 200 \
	--log-level info
echo "✓ 扫描完成"
echo ""

# 5. 清理
echo "[5/5] 清理..."
kill $SERVER_PID 2>/dev/null || true
rm -f test_server test_server.go
echo "✓ 清理完成"
echo ""

echo "=== 测试完成 ==="
