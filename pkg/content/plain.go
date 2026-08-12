/*
 * 文件作用：纯文本 Parser -- 处理 text/plain，统一换行并裁剪空白。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"context"
	"strings"

	"rove/pkg/document"
)

// TextParser 处理 text/plain。
type TextParser struct{}

// NewTextParser 构造 TextParser。
func NewTextParser() *TextParser {

	return &TextParser{}
}

// Parse 直接规范化文本。
func (p *TextParser) Parse(ctx context.Context, raw *document.RawDocument) (*ParsedContent, error) {

	text := normalizeText(string(raw.Body))
	return &ParsedContent{
		Content:  text,
		Markdown: text,
		Metadata: map[string]any{},
	}, nil
}

// normalizeText 统一换行并裁剪首尾空白。
func normalizeText(text string) string {

	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.TrimSpace(text)
}
