#!/bin/bash

echo "🔍 Testing Market Watch API endpoints..."

BASE_URL="http://localhost:8080"

echo ""
echo "1. Testing health check..."
curl -s "$BASE_URL/api/health" | jq '.' || echo "Health check failed"

echo ""
echo "2. Testing data count..."
curl -s "$BASE_URL/api/debug/count" | jq '.' || echo "Data count failed"

echo ""
echo "3. Testing collection status..."
curl -s "$BASE_URL/api/collection/status" | jq '.' || echo "Collection status failed"

echo ""
echo "4. Testing dashboard summary..."
curl -s "$BASE_URL/api/dashboard/summary" | jq '.' || echo "Dashboard summary failed"

echo ""
echo "5. Testing chart data for PLTR..."
curl -s "$BASE_URL/api/volume/PLTR/chart?range=1D" | jq '.' || echo "Chart data failed"

echo ""
echo "6. Force collection..."
curl -s -X POST "$BASE_URL/api/collection/force" | jq '.' || echo "Force collection failed"

echo ""
echo "✅ API tests completed"
