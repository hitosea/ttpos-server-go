#!/bin/bash
# Fix Go coverage paths for SonarQube compatibility.
# BMP module name is "ttpos-bmp" but SonarQube expects "ttpos-bmp/" filesystem paths.
# Reuses the same transformation logic as main/tests/fix-coverage-paths.sh.
#
# Usage: ./fix-coverage-paths.sh <coverage-file> [go-mod-path]

set -euo pipefail

COVERAGE_FILE="${1:?Usage: fix-coverage-paths.sh <coverage-file> [go-mod-path]}"
GO_MOD="${2:-ttpos-bmp/go.mod}"

if [ ! -f "$COVERAGE_FILE" ]; then
  echo "Coverage file not found: $COVERAGE_FILE"
  exit 0
fi

if [ ! -f "$GO_MOD" ]; then
  echo "go.mod not found: $GO_MOD, using hardcoded module name"
  MODULE="ttpos-bmp"
else
  MODULE=$(head -1 "$GO_MOD" | awk '{print $2}')
fi

TARGET_DIR=$(dirname "$GO_MOD")

echo "Fixing coverage paths: ${MODULE}/ -> ${TARGET_DIR}/"
sed -i "s|${MODULE}/|${TARGET_DIR}/|g" "$COVERAGE_FILE"

echo "Done: $COVERAGE_FILE"
