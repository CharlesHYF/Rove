#!/usr/bin/env bash
# 文件作用：测试总入口 -- 检测项目技术栈并执行对应测试;零测试被执行则判失败(除非 ALLOW_NO_TESTS=1 显式放行);解析 go test -json 统计被跳过测试,防止"全跳过假绿"(ROVE_FAIL_ON_SKIP=1 可转失败)。
# 创建日期：2026-08-12
# 修改日期：2026-08-15
set -e

echo "运行测试..."

# 实际被执行的测试套件计数;全程为 0 视为"未真正跑测试",按交付红线判失败
RAN=0

# Go 测试中被跳过的用例数(由 -json 输出统计)
SKIPPED=0

if [ -f go.mod ]; then
  echo "Go 测试..."
  RAN=$((RAN + 1))

  skip_log=$(mktemp)
  go test -json ./... -cover 2>&1 | awk -v skip_log="${skip_log}" '
    /"Action":"skip"/ {
      skips++
      name = $0
      sub(/^.*"Test":"/, "", name)
      sub(/".*$/, "", name)
      print "  [SKIP] " name
    }
    /"Action":"fail"/ && /"Test":"/ {
      name = $0
      sub(/^.*"Test":"/, "", name)
      sub(/".*$/, "", name)
      print "  [FAIL] " name
    }
    END {
      print skips > skip_log
    }
  '
  test_status=${PIPESTATUS[0]}
  SKIPPED=$(cat "${skip_log}")
  rm -f "${skip_log}"

  if [ "${test_status}" -ne 0 ]; then
    echo "[NG] Go 测试存在失败,请修复后再交付。"
    exit 1
  fi
fi

if [ -f pom.xml ]; then
  echo "Java 测试..."
  mvn test
  RAN=$((RAN + 1))
fi

if [ -f pyproject.toml ]; then
  echo "Python 测试..."
  uv run pytest --cov --cov-fail-under=80
  RAN=$((RAN + 1))
fi

# 前端测试:仅当存在 frontend/ 时进入,避免纯后端项目在 cd 处崩溃
if [ -d frontend ] && [ -f frontend/package.json ]; then
  echo "前端单元测试..."
  (cd frontend && npm test --if-present)
  RAN=$((RAN + 1))
fi

# 交付红线:至少跑过一种测试;确无测试的原型/脚本项目须显式 ALLOW_NO_TESTS=1
if [ "${RAN}" -eq 0 ]; then
  if [ "${ALLOW_NO_TESTS:-0}" = "1" ]; then
    echo "[WARN] 未执行任何测试,但 ALLOW_NO_TESTS=1 已显式放行。"
  else
    echo "[NG] 未执行任何测试(未找到 go.mod/pom.xml/pyproject.toml/frontend)。原型或脚本项目请用 ALLOW_NO_TESTS=1 make check 显式放行。"
    exit 1
  fi
fi

# 跳过检测:默认仅告警;ROVE_FAIL_ON_SKIP=1 时任何跳过都判失败,防止 e2e 静默跳过导致假绿
if [ "${SKIPPED}" -gt 0 ]; then
  if [ "${ROVE_FAIL_ON_SKIP:-0}" = "1" ]; then
    echo "[NG] 发现 ${SKIPPED} 个被跳过的测试,且 ROVE_FAIL_ON_SKIP=1,判定失败(请检查 ES/浏览器等依赖是否就绪)。"
    exit 1
  fi
  echo "[WARN] 发现 ${SKIPPED} 个被跳过的测试(常见原因: ES/浏览器未启动)。本地开发允许;交付/CI 建议设 ROVE_FAIL_ON_SKIP=1 强制。"
fi

echo "[OK] 测试执行完毕(实际执行 ${RAN} 类,跳过 ${SKIPPED} 个)。"
