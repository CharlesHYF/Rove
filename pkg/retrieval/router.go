/*
 * 文件作用：Query Router -- 确定性查询垂类推断（web/docs/code/academic，无 LLM）与参数校验（PRD FR-QRY-001）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package retrieval

import (
	"fmt"
	"strings"
)

// Vertical 检索垂类。
type Vertical string

// 垂类取值。
const (
	VerticalAuto     Vertical = ""         // 自动推断
	VerticalWeb      Vertical = "web"      // 通用 Web（默认）
	VerticalDocs     Vertical = "docs"     // 文档
	VerticalCode     Vertical = "code"     // 代码
	VerticalAcademic Vertical = "academic" // 学术
)

// 垂类标记词表（按代码 -> 学术 -> 文档优先级匹配）。
var (
	codeMarkers = []string{
		"func ", "def ", "class ", "import ", "()", "github.com/",
		"npm ", "pip install", "git clone", "snippet", "syntax",
	}
	academicMarkers = []string{
		"arxiv", "doi:", "doi.org", "citation", "scholar", "et al",
		"paper", "conference", "thesis",
	}
	docsMarkers = []string{
		"how to", "how do i", "tutorial", "documentation", "install",
		"configure", "guide", "usage", "reference",
	}
)

// RouteVertical 确定性推断查询垂类（无 LLM）：code > academic > docs > web。
func RouteVertical(text string) Vertical {

	lower := strings.ToLower(text)
	if containsAnyMarker(lower, codeMarkers) {
		return VerticalCode
	}
	if containsAnyMarker(lower, academicMarkers) {
		return VerticalAcademic
	}
	if containsAnyMarker(lower, docsMarkers) {
		return VerticalDocs
	}
	return VerticalWeb
}

// ParseVertical 解析并校验垂类参数。
func ParseVertical(value string) (Vertical, error) {

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "auto":
		return VerticalAuto, nil
	case "web":
		return VerticalWeb, nil
	case "docs":
		return VerticalDocs, nil
	case "code":
		return VerticalCode, nil
	case "academic":
		return VerticalAcademic, nil
	default:
		return "", fmt.Errorf("vertical 仅支持 auto/web/docs/code/academic，收到 %q", value)
	}
}

// containsAnyMarker 判断文本是否包含任一标记。
func containsAnyMarker(text string, markers []string) bool {

	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}
