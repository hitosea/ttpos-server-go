#!/usr/bin/env bash
#
# Orchestrator: setup → K6 test → SQL verify → cleanup
#
# Usage:  bash run.sh [--keep]
#   --keep    Skip database cleanup (for manual inspection)
#
set -euo pipefail
cd "$(dirname "$0")"

K6="${K6:-$(command -v k6 || echo ~/bin/k6)}"
KEEP_DB=false

for arg in "$@"; do
  case "$arg" in
    --keep) KEEP_DB=true ;;
  esac
done

echo "══════════════════════════════════════════════════"
echo " Order-First-Pay-Later K6 Test Suite"
echo "══════════════════════════════════════════════════"

# ── Phase 1: Setup ─────────────────────────────────────────────────
echo ""
echo "▶ Phase 1: Setup test tenant..."
node setup.mjs
echo ""

# Read DB name for cleanup
DB_NAME=$(node -e "console.log(JSON.parse(require('fs').readFileSync('env.json','utf8')).DB_NAME)")

cleanup() {
  if [ "$KEEP_DB" = true ]; then
    echo "[cleanup] --keep flag set, skipping DB drop. Database: $DB_NAME"
    return
  fi
  echo "[cleanup] Dropping test database: $DB_NAME"
  node -e "
    const mysql = require('mysql2/promise');
    const env = JSON.parse(require('fs').readFileSync('env.json','utf8'));
    (async () => {
      const conn = await mysql.createConnection({
        host: env.DB_HOST, port: parseInt(env.DB_PORT),
        user: 'root', password: '${DB_ROOT_PASSWORD:-8e5ed3c0a7368953}',
      });
      await conn.execute('DROP DATABASE IF EXISTS \`${DB_NAME}\`');
      await conn.end();
      console.log('[cleanup] Done');
    })().catch(e => console.error('[cleanup] Warning:', e.message));
  "
}

# Register cleanup on exit (unless --keep)
trap cleanup EXIT

# ── Phase 2: K6 Test ──────────────────────────────────────────────
echo "▶ Phase 2: Running K6 tests..."
echo ""
"$K6" run order_first_test.js --no-color 2>&1 || K6_EXIT=$?
K6_EXIT=${K6_EXIT:-0}
echo ""

# ── Phase 3: SQL Verify ──────────────────────────────────────────
echo "▶ Phase 3: SQL verification..."
node verify.mjs || VERIFY_EXIT=$?
VERIFY_EXIT=${VERIFY_EXIT:-0}
echo ""

# ── Summary ──────────────────────────────────────────────────────
echo "══════════════════════════════════════════════════"
if [ "$K6_EXIT" -eq 0 ] && [ "$VERIFY_EXIT" -eq 0 ]; then
  echo " RESULT: ALL PASSED"
else
  echo " RESULT: FAILED (k6=$K6_EXIT, verify=$VERIFY_EXIT)"
fi
echo "══════════════════════════════════════════════════"

exit $((K6_EXIT + VERIFY_EXIT))
