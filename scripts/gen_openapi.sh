#!/bin/bash
set -e

echo "Generating OpenAPI code..."

oapi-codegen \
  -package dto \
  -generate types,chi-server \
  -o presentation/http/dto/server.gen.go \
  shared/api/openapi.yaml

echo "✅ OpenAPI code generation completed!"
