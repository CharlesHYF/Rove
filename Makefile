# 作用：Rove 工程 Makefile -- 提供规范校验(lint)、测试(check)与交付总闸门(verify)。
# 创建日期：2026-08-12
# 修改日期：2026-08-12

.PHONY: lint check verify build

lint: ## 规范校验（禁用字符/文件头/必需文件/命名，会 fail）
	bash scripts/check.sh

check: ## 运行所有测试
	bash scripts/test.sh

verify: ## 交付前总闸门：先规范校验，全绿后再跑测试
	bash scripts/check.sh
	bash scripts/test.sh

build: ## 编译 CLI
	go build -o bin/rove ./cmd/rove
