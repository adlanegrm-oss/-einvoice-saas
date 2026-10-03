#!/bin/bash
set -e

echo "Applying core foundation patch..."
mkdir -p pkg/money pkg/tax pkg/validation pkg/lifecycle docs

cp -r pkg/* ./pkg/ 2>/dev/null || true
cp -r docs/* ./docs/ 2>/dev/null || true

echo "Running tests..."
go test ./pkg/money/... -v
echo "Patch applied successfully!"
