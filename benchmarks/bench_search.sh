#!/usr/bin/env bash
# 文件作用：检索基准脚本 -- 抓取 seed 后对一组查询运行搜索，输出指标信息。
# 创建日期：2026-08-12
# 修改日期：2026-08-12
set -euo pipefail

SEED="${1:?usage: bench_search.sh <seed-url> [query...]}"
shift || true

./bin/rove index init >/dev/null 2>&1 || true
./bin/rove crawl "$SEED" --max-pages 20 >/dev/null

for query in "$@"; do
  echo "== query: $query"
  ./bin/rove search "$query" --json | head -c 600
  echo
done
