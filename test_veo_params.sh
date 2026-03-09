#!/bin/bash

# Veo 3.1 参数适配测试脚本

BASE_URL="http://localhost:3000"
API_KEY="test_key"

echo "========================================="
echo "测试 1: 基础文本到视频生成"
echo "========================================="
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -s \
  "$BASE_URL/v1/video/generations" \
  -d '{
  "model": "veo-3.1-generate-preview",
  "prompt": "Test text-to-video generation",
  "duration": 8,
  "metadata": {
    "generateAudio": true
  }
}' | jq -r 'if .id then "✅ 成功: \(.id)" else "❌ 失败: \(.)" end'

echo ""
echo "========================================="
echo "测试 2: 使用 n 参数生成多个视频"
echo "========================================="
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -s \
  "$BASE_URL/v1/video/generations" \
  -d '{
  "model": "veo-3.1-fast-generate-preview",
  "prompt": "Test multi-video generation",
  "duration": 6,
  "n": 3,
  "metadata": {
    "generateAudio": true,
    "seed": 12345
  }
}' | jq -r 'if .id then "✅ 成功: \(.id)" else "❌ 失败: \(.)" end'

echo ""
echo "========================================="
echo "测试 3: 使用 width/height 设置分辨率"
echo "========================================="
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -s \
  "$BASE_URL/v1/video/generations" \
  -d '{
  "model": "veo-3.1-generate-preview",
  "prompt": "Test resolution with width/height",
  "width": 1920,
  "height": 1080,
  "duration": 8,
  "metadata": {
    "generateAudio": true
  }
}' | jq -r 'if .id then "✅ 成功: \(.id)" else "❌ 失败: \(.)" end'

echo ""
echo "========================================="
echo "测试 4: 使用 metadata.sampleCount 覆盖"
echo "========================================="
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -s \
  "$BASE_URL/v1/video/generations" \
  -d '{
  "model": "veo-3.1-generate-preview",
  "prompt": "Test metadata sampleCount override",
  "duration": 4,
  "metadata": {
    "generateAudio": true,
    "sampleCount": 2,
    "resolution": "1080p"
  }
}' | jq -r 'if .id then "✅ 成功: \(.id)" else "❌ 失败: \(.)" end'

echo ""
echo "========================================="
echo "测试 5: 错误的 duration (应该返回错误)"
echo "========================================="
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -s \
  "$BASE_URL/v1/video/generations" \
  -d '{
  "model": "veo-3.1-generate-preview",
  "prompt": "Test invalid duration",
  "duration": 5,
  "metadata": {
    "generateAudio": true
  }
}' | jq -r 'if .error then "✅ 预期错误: \(.error.message)" else "❌ 应该返回错误但成功: \(.)" end'

echo ""
echo "========================================="
echo "测试 6: 不提供 generateAudio (应自动设置为true)"
echo "========================================="
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -s \
  "$BASE_URL/v1/video/generations" \
  -d '{
  "model": "veo-3.1-generate-preview",
  "prompt": "Test auto generateAudio",
  "duration": 8
  }' | jq -r 'if .id then "✅ 成功 (自动设置generateAudio): \(.id)" else "❌ 失败: \(.)" end'

echo ""
echo "========================================="
echo "测试 7: 边界值 - n=4 (最大值)"
echo "========================================="
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -s \
  "$BASE_URL/v1/video/generations" \
  -d '{
  "model": "veo-3.1-generate-preview",
  "prompt": "Test max sampleCount",
  "duration": 8,
  "n": 4,
  "metadata": {
    "generateAudio": true,
    "seed": 999999
  }
}' | jq -r 'if .id then "✅ 成功 (n=4): \(.id)" else "❌ 失败: \(.)" end'

echo ""
echo "========================================="
echo "测试 8: 边界值 - duration=4 (最小值)"
echo "========================================="
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -s \
  "$BASE_URL/v1/video/generations" \
  -d '{
  "model": "veo-3.1-generate-preview",
  "prompt": "Test min duration",
  "duration": 4,
  "metadata": {
    "generateAudio": true
  }
}' | jq -r 'if .id then "✅ 成功 (duration=4): \(.id)" else "❌ 失败: \(.)" end'

echo ""
echo "========================================="
echo "测试 9: 使用 4K 分辨率"
echo "========================================="
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -s \
  "$BASE_URL/v1/video/generations" \
  -d '{
  "model": "veo-3.1-generate-preview",
  "prompt": "Test 4K resolution",
  "duration": 8,
  "metadata": {
    "generateAudio": true,
    "resolution": "4k",
    "aspectRatio": "16:9"
  }
}' | jq -r 'if .id then "✅ 成功 (4K): \(.id)" else "❌ 失败: \(.)" end'

echo ""
echo "========================================="
echo "测试完成！"
echo "========================================="
