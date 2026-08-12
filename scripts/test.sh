#!/bin/bash
# 文件作用：测试总入口 -- 检测项目技术栈并执行对应测试;零测试被执行则判失败(除非 ALLOW_NO_TESTS=1 显式放行)。
# 创建日期：2026-08-12
# 修改日期：2026-08-12
set -e

echo "运行测试..."

# 实际被执行的测试套件计数;全程为 0 视为"未真正跑测试",按交付红线判失败
RAN=0

if [ -f go.mod ]; then
  echo "Go 测试..."
  go test ./... -cover
  RAN=$((RAN + 1))
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

echo "[OK] 测试执行完毕(实际执行 ${RAN} 类)。"
