#!/bin/bash

echo "=== PetMiner API 测试脚本 ==="

# 基础URL
BASE_URL="http://localhost:8081"

echo "1. 测试健康检查端点..."
curl -s -w "\n状态码: %{http_code}\n时间: %{time_total}s\n" \
     -X GET "$BASE_URL/health"

echo -e "\n2. 测试版本信息端点..."
curl -s -w "\n状态码: %{http_code}\n时间: %{time_total}s\n" \
     -X GET "$BASE_URL/version"

echo -e "\n3. 测试获取所有宠物 (GET /api/v1/pets)..."
curl -s -w "\n状态码: %{http_code}\n时间: %{time_total}s\n" \
     -X GET "$BASE_URL/api/v1/pets" \
     -H "Content-Type: application/json"

echo -e "\n4. 测试创建宠物 (POST /api/v1/pets)..."
curl -s -w "\n状态码: %{http_code}\n时间: %{time_total}s\n" \
     -X POST "$BASE_URL/api/v1/pets" \
     -H "Content-Type: application/json" \
     -d '{"owner_name": "test_owner"}'

echo -e "\n5. 再次测试获取所有宠物..."
curl -s -w "\n状态码: %{http_code}\n时间: %{time_total}s\n" \
     -X GET "$BASE_URL/api/v1/pets" \
     -H "Content-Type: application/json"

echo -e "\n6. 测试获取事件 (GET /api/v1/events)..."
curl -s -w "\n状态码: %{http_code}\n时间: %{time_total}s\n" \
     -X GET "$BASE_URL/api/v1/events?limit=5" \
     -H "Content-Type: application/json"

echo -e "\n=== 测试完成 ==="