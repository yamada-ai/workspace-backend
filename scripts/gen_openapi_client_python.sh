#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OPENAPI_SPEC="$PROJECT_ROOT/shared/api/openapi.yaml"
OUTPUT_DIR="$PROJECT_ROOT/twitch-bot/app/api/generated"

echo "🐍 Generating Python client from OpenAPI spec..."

# openapi-python-client が必要
if ! command -v openapi-python-client &> /dev/null; then
    echo "❌ openapi-python-client not found."
    echo "Install it with: pip install openapi-python-client"
    exit 1
fi

# 出力ディレクトリをクリーンアップ
rm -rf "$OUTPUT_DIR"
mkdir -p "$(dirname "$OUTPUT_DIR")"

# クライアント生成
openapi-python-client generate \
  --path "$OPENAPI_SPEC" \
  --output-path "$OUTPUT_DIR" \
  --overwrite

echo "✅ Python client generation completed!"
echo "📁 Generated files in: $OUTPUT_DIR"
