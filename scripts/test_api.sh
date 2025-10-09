#!/bin/bash

BASE_URL=${1:-http://localhost:8000}

echo "🧪 Testing Workspace Backend API"
echo "Base URL: $BASE_URL"
echo ""

# Test 1: Health Check
echo "1️⃣  Testing /health endpoint..."
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")
if [ "$HEALTH_RESPONSE" = "ok" ]; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed: $HEALTH_RESPONSE"
    exit 1
fi
echo ""

# Test 2: Join Command (New User)
echo "2️⃣  Testing /api/commands/join (new user)..."
JOIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/commands/join" \
  -H "Content-Type: application/json" \
  -d '{
    "user_name": "test_user_'$(date +%s)'",
    "tier": 1,
    "work_name": "テスト作業"
  }')

echo "Response: $JOIN_RESPONSE"

# Parse session_id and user_id
SESSION_ID=$(echo "$JOIN_RESPONSE" | grep -o '"session_id":[0-9]*' | grep -o '[0-9]*')
USER_ID=$(echo "$JOIN_RESPONSE" | grep -o '"user_id":[0-9]*' | grep -o '[0-9]*')

if [ -n "$SESSION_ID" ] && [ -n "$USER_ID" ]; then
    echo "✅ Join command succeeded"
    echo "   User ID: $USER_ID"
    echo "   Session ID: $SESSION_ID"
else
    echo "❌ Join command failed"
    exit 1
fi
echo ""

# Test 3: Join Command (Duplicate - should return same session)
echo "3️⃣  Testing /api/commands/join (duplicate user)..."
JOIN2_RESPONSE=$(curl -s -X POST "$BASE_URL/api/commands/join" \
  -H "Content-Type: application/json" \
  -d '{
    "user_name": "duplicate_test",
    "tier": 2,
    "work_name": "初回作業"
  }')

SESSION2_ID=$(echo "$JOIN2_RESPONSE" | grep -o '"session_id":[0-9]*' | grep -o '[0-9]*')

# Try joining again
JOIN3_RESPONSE=$(curl -s -X POST "$BASE_URL/api/commands/join" \
  -H "Content-Type: application/json" \
  -d '{
    "user_name": "duplicate_test",
    "tier": 2,
    "work_name": "2回目の試み"
  }')

SESSION3_ID=$(echo "$JOIN3_RESPONSE" | grep -o '"session_id":[0-9]*' | grep -o '[0-9]*')

if [ "$SESSION2_ID" = "$SESSION3_ID" ]; then
    echo "✅ Duplicate join test passed (returned same session)"
    echo "   Session ID: $SESSION2_ID"
else
    echo "❌ Duplicate join test failed (created new session)"
    echo "   First session: $SESSION2_ID"
    echo "   Second session: $SESSION3_ID"
    exit 1
fi
echo ""

echo "🎉 All tests passed!"
