/*
 * 文件作用：Markdown Parser -- 处理 text/markdown，规范化文本并从首个一级标题提取 Title。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"context"
	"strings"

	"rove/pkg/document"
)

// MarkdownParser 处理 text/markdown。
type MarkdownParser struct{}

// NewMarkdownParser 构造 MarkdownParser。
func NewMarkdownParser() *MarkdownParser {

	return &MarkdownParser{}
}

// Parse 规范化 Markdown，并从首个 "# " 标题提取 Title。
func (p *MarkdownParser) Parse(ctx context.Context, raw *document.RawDocument) (*ParsedContent, error) {

	markdown := normalizeText(string(raw.Body))
	pc := &ParsedContent{Content: markdown, Markdown: markdown, Metadata: map[string]any{}}
	for _, line := range strings.Split(markdown, "\n") {
		if strings.HasPrefix(line, "# ") {
			pc.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			break
		}
	}
	return pc, nil
}
