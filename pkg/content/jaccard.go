/*
 * 文件作用：token 集合 Jaccard 相似度 -- 确定性近重复检测（PRD FR-CNT-003 允许 SimHash/MinHash 等确定性方法，实现选用 Jaccard 精确版）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"strings"
	"unicode"
)

// tokenSet 将文本转换为小写 token 集合（按非字母数字分隔）。
func tokenSet(text string) map[string]struct{} {

	set := make(map[string]struct{})
	for _, token := range tokenize(text) {
		set[token] = struct{}{}
	}
	return set
}

// jaccard 计算两个 token 集合的 Jaccard 相似度；任一为空返回 0。
func jaccard(a, b map[string]struct{}) float64 {

	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	intersection := 0
	for token := range a {
		if _, ok := b[token]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// tokenize 将文本切成小写 token（按非字母数字分隔）。
func tokenize(text string) []string {

	var tokens []string
	var current strings.Builder
	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return tokens
}
