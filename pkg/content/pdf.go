/*
 * 文件作用：PDF Parser -- 提取全部页面文本，进入统一 Document 管线（PRD FR-CNT-001）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"bytes"
	"context"
	"strings"

	"github.com/ledongthuc/pdf"

	"rove/pkg/document"
	"rove/pkg/rove"
)

// PDFParser 处理 application/pdf。
type PDFParser struct{}

// NewPDFParser 构造 PDFParser。
func NewPDFParser() *PDFParser {

	return &PDFParser{}
}

// Parse 提取 PDF 全部页面文本。
func (p *PDFParser) Parse(ctx context.Context, raw *document.RawDocument) (*ParsedContent, error) {

	reader, err := pdf.NewReader(bytes.NewReader(raw.Body), int64(len(raw.Body)))
	if err != nil {
		return nil, rove.NewError("content.pdf_open", rove.CategoryContent, false, "pdf open: %v", err)
	}

	var sb strings.Builder
	for page := 1; page <= reader.NumPage(); page++ {
		text, err := reader.Page(page).GetPlainText(nil)
		if err != nil {
			return nil, rove.NewError("content.pdf_extract", rove.CategoryContent, false, "pdf page %d: %v", page, err)
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	text := normalizeText(sb.String())
	return &ParsedContent{Content: text, Markdown: text, Metadata: map[string]any{}}, nil
}
