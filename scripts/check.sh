#!/usr/bin/env bash
# 文件作用：规范校验器 -- 把 charles-coding 里的确定性规则(禁用字符/文件头/必需文件/命名)变成会 fail 的检查，交付前由 make lint 强制执行，不依赖 Agent 自觉。
# 创建日期：2026-08-12
# 修改日期：2026-08-12

# 说明：故意不用 set -e。grep/perl 无匹配时返回非 0 属正常，需手动累计错误而非中断。
set -uo pipefail

# 违规计数：任意一项 > 0 则最终退出码非 0，供 CI / make lint 拦截
VIOLATIONS=0

# 源码扩展名白名单：只有这些文件强制校验文件头注释块
SOURCE_EXT_REGEX='\.(java|kt|kts|go|py|js|jsx|ts|tsx|vue|sql|sh|html|css|scss)$'

# 文本扫描排除的二进制/资源扩展名(禁用字符检查跳过这些)
BINARY_EXT_REGEX='\.(png|jpe?g|gif|webp|ico|svg|pdf|zip|gz|tar|jar|class|woff2?|ttf|eot|mp[34]|mov|lock)$'

# 必需的项目级文件清单
REQUIRED_FILES=(
	"AGENTS.md"
	"README.md"
	".gitignore"
	".editorconfig"
	".gitattributes"
)

# 打印一条违规信息并累加计数
report() {
	local message="$1"

	echo "  [FAIL] ${message}"
	VIOLATIONS=$((VIOLATIONS + 1))
}

# 列出待检查文件：优先用 git 跟踪清单，非 git 目录退回 find
list_files() {

	if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
		git ls-files
	else
		find . -type f -not -path './.git/*' | sed 's|^\./||'
	fi
}

# 检查一：禁用字符(引号 / 破折号 / Emoji;含各类 Unicode 变体)
check_forbidden_chars() {
	echo "[1/4] 检查禁用字符(引号 / 破折号 / Emoji)..."

	local file
	while IFS= read -r file; do

		if [[ "${file}" =~ ${BINARY_EXT_REGEX} ]]; then
			continue
		fi

		if [ ! -f "${file}" ]; then
			continue
		fi

		# 引号类:弯引号 U+2018/2019/201C/201D、CJK 角引号 U+300C-300F、全角引号 U+FF02/FF07
		local hits
		hits=$(perl -CSD -ne 'print "L$.: $_" if /[\x{2018}\x{2019}\x{201c}\x{201d}\x{300c}-\x{300f}\x{ff02}\x{ff07}]/' "${file}")

		if [ -n "${hits}" ]; then
			report "${file} 含非法引号(应改用半角 \" 或 '):"
			echo "${hits}" | sed 's/^/         /'
		fi

		# 破折号/横线类:U+2010-2015、U+2212 减号、U+FF0D 全角连字符、U+2E3A/2E3B 双三 em dash
		local dash_hits
		dash_hits=$(perl -CSD -ne 'print "L$.: $_" if /[\x{2010}-\x{2015}\x{2212}\x{2e3a}\x{2e3b}\x{ff0d}]/' "${file}")

		if [ -n "${dash_hits}" ]; then
			report "${file} 含 Unicode 破折号/横线(应改用半角 --):"
			echo "${dash_hits}" | sed 's/^/         /'
		fi

		# Emoji：仅匹配主要 Emoji 区段，规避 → 等合法箭头符号
		local emoji_hits
		emoji_hits=$(perl -CSD -ne 'print "L$.: $_" if /[\x{1F300}-\x{1FAFF}\x{1F1E6}-\x{1F1FF}\x{2600}-\x{27BF}\x{FE0F}]/' "${file}")

		if [ -n "${emoji_hits}" ]; then
			report "${file} 含 Emoji(应改用 SVG 或 icon 字体):"
			echo "${emoji_hits}" | sed 's/^/         /'
		fi

	done < <(list_files)
}

# 检查二：源码文件头注释块(须含"作用"与"创建日期"两个标记)
check_file_header() {
	echo "[2/4] 检查源码文件头注释块..."

	local file
	while IFS= read -r file; do

		if [[ ! "${file}" =~ ${SOURCE_EXT_REGEX} ]]; then
			continue
		fi

		if [ ! -f "${file}" ]; then
			continue
		fi

		# 扫整个文件找"作用"与"创建日期"两个标记(不限行数)
		if ! grep -q "作用" "${file}"; then
			report "${file} 缺少\"文件作用\"说明(源码注释块)"
			continue
		fi

		if ! grep -q "创建日期" "${file}"; then
			report "${file} 缺少\"创建日期\"(源码注释块)"
		fi

	done < <(list_files)
}

# 检查三：项目级必需文件是否齐全
check_required_files() {
	echo "[3/4] 检查项目必需文件..."

	local required
	for required in "${REQUIRED_FILES[@]}"; do

		if [ ! -f "${required}" ]; then
			report "缺少必需文件：${required}"
		fi
	done

	# docs/modules/ 须存在,且模块文档必须在系统模块子目录下(两级:docs/modules/<系统模块>/<模块>.md)
	if [ ! -d "docs/modules" ]; then
		report "缺少模块文档目录：docs/modules/"
	else

		# 一级违规:直接躺在 docs/modules/ 根的 .md(README.md 除外)
		local module_flat
		module_flat=$(find docs/modules -maxdepth 1 -type f -name '*.md' ! -name 'README.md')

		if [ -n "${module_flat}" ]; then
			report "docs/modules/ 下有一级模块文档(须放进系统模块子目录 docs/modules/<系统模块>/x.md):"
			echo "${module_flat}" | sed 's/^/         /'
		fi

		# 至少一个合规的两级模块文档
		local module_doc_count
		module_doc_count=$(find docs/modules -mindepth 2 -type f -name '*.md' ! -name 'README.md' | wc -l | tr -d ' ')

		if [ "${module_doc_count}" -eq 0 ]; then
			report "docs/modules/ 下无模块文档(编码前必须先写 docs/modules/<系统模块>/<模块>.md)"
		fi
	fi

	# test_cases/ 须存在,且用例必须在两级子目录下(test_cases/<系统模块>/<测试类型>/x.md)
	if [ ! -d "test_cases" ]; then
		report "缺少测试用例目录：test_cases/(交付必须附带测试用例)"
	else

		# 浅层违规:躺在 test_cases/ 根或仅一层子目录里的 .md(README.md 除外)
		local case_flat
		case_flat=$(find test_cases -maxdepth 2 -type f -name '*.md' ! -name 'README.md')

		if [ -n "${case_flat}" ]; then
			report "test_cases/ 下有浅层用例文档(须放进 test_cases/<系统模块>/<类型>/x.md):"
			echo "${case_flat}" | sed 's/^/         /'
		fi

		# 至少一个合规的用例文档(位于 <系统模块>/<类型>/ 下)
		local case_doc_count
		case_doc_count=$(find test_cases -mindepth 3 -type f -name '*.md' ! -name 'README.md' | wc -l | tr -d ' ')

		if [ "${case_doc_count}" -eq 0 ]; then
			report "test_cases/ 下无用例文档(交付必须附带 test_cases/<系统模块>/<类型>/x.md)"
		fi
	fi
}

# 检查四：数据传输命名(Java / Kotlin / TS 的响应/请求类)
check_naming() {
	echo "[4/4] 检查数据传输命名(XxxReqVO / XxxRespVO)..."

	local file
	while IFS= read -r file; do

		if [[ ! "${file}" =~ \.(java|kt|ts|tsx|vue)$ ]]; then
			continue
		fi

		if [ ! -f "${file}" ]; then
			continue
		fi

		local bad_vo
		bad_vo=$(grep -nE '\b(class|interface|type)[[:space:]]+[A-Z][A-Za-z0-9]*VO\b' "${file}" \
			| grep -vE '((Req|Resp)VO|(Base|Abstract|Page)[A-Za-z0-9]*VO)\b' || true)

		if [ -n "${bad_vo}" ]; then
			report "${file} 存在裸 VO 命名(请求应为 XxxReqVO、响应应为 XxxRespVO):"
			echo "${bad_vo}" | sed 's/^/         /'
		fi

	done < <(list_files)
}

echo "=== charles-coding 规范校验 ==="

check_forbidden_chars
check_file_header
check_required_files
check_naming

echo "==============================="

if [ "${VIOLATIONS}" -gt 0 ]; then
	echo "[NG] 发现 ${VIOLATIONS} 处规范违规，请修正后再交付。"
	exit 1
fi

echo "[OK] 规范校验通过。"
exit 0
