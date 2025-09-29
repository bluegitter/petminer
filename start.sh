#!/bin/bash

# MiningPet 启动脚本
echo "🐾 欢迎来到 MiningPet!"
echo "================================"

# 检查是否安装了Go
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未检测到Go，请先安装Go 1.21+"
    exit 1
fi

# 检查是否安装了Node.js
if ! command -v node &> /dev/null; then
    echo "❌ 错误: 未检测到Node.js，请先安装Node.js 16+"
    exit 1
fi

# 启动后端服务器
echo "🚀 启动后端服务器..."
cd backend
go mod tidy
go run cmd/server/main.go &
BACKEND_PID=$!
cd ..

# 等待后端启动
echo "⏳ 等待后端服务器启动..."
sleep 10 

# 检查后端是否运行
if ! curl -s http://localhost:8081/api/v1/pets > /dev/null; then
    echo "❌ 后端服务器启动失败"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

echo "✅ 后端服务器启动成功 (PID: $BACKEND_PID)"

# 启动前端开发服务器
echo "🎨 启动前端开发服务器..."
cd frontend

# 安装依赖（如果需要）
if [ ! -d "node_modules" ]; then
    echo "📦 安装前端依赖..."
    npm install
fi

npm start &
FRONTEND_PID=$!
cd ..

echo "✅ 前端开发服务器启动成功 (PID: $FRONTEND_PID)"
echo ""
echo "🎮 游戏已启动!"
echo "前端地址: http://localhost:3000"
echo "后端API: http://localhost:8081"
echo ""
echo "按 Ctrl+C 停止所有服务"

# 设置信号处理
trap 'echo ""; echo "🛑 正在停止服务..."; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit 0' INT

# 等待进程
wait
