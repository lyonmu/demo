#!/bin/bash

# WebSocket Demo 启动脚本（使用 Envoy 代理）

echo "🚀 启动 WebSocket Demo（使用 Envoy 代理）"
echo ""

# 检查后端服务是否运行
if ! lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null ; then
    echo "📦 启动后端服务..."
    go run main.go &
    BACKEND_PID=$!
    echo "✅ 后端服务已启动 (PID: $BACKEND_PID)"
    sleep 2
else
    echo "✅ 后端服务已在运行"
fi

# 检查 Envoy 是否运行
if ! docker ps | grep -q envoy-proxy-ws-demo; then
    echo "📦 启动 Envoy 代理..."
    docker-compose -f envoy-compose.yml up -d
    sleep 2
    echo "✅ Envoy 代理已启动"
else
    echo "✅ Envoy 代理已在运行"
fi

echo ""
echo "✨ 服务已就绪！"
echo ""
echo "📝 访问地址："
echo "   - 通过 Envoy 代理: http://localhost:19894"
echo "   - 直接访问后端: http://localhost:8080"
echo "   - Envoy 管理界面: http://localhost:19901"
echo ""
echo "按 Ctrl+C 停止服务"

# 等待用户中断
trap "echo ''; echo '🛑 正在停止服务...'; docker-compose -f envoy-compose.yml down; kill $BACKEND_PID 2>/dev/null; exit" INT
wait

