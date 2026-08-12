/*
 * 文件作用：定义 ParsedContent 类型与 Parser 注册表 -- 按 MIME 类型路由解析器。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package content 实现内容管线：解析、提取、规范化、去重与分块。
package content

import (
	"context"
	"strings"
	"time"

	"rove/pkg/document"
	"rove/pkg/rove"
)

// ParsedContent 是 Parser 的输出，供管线后续步骤消费。
type ParsedContent struct {
	Title        string
	Description  string
	Content      string
	Markdown     string
	CanonicalURL string // <link rel="canonical">，可为空
	Links        []string
	Language     string
	PublishedAt  *time.Time
	UpdatedAt    *time.Time
	Metadata     map[string]any
}

// Parser 按 Content-Type 将 RawDocument 解析为 ParsedContent。
type Parser interface {
	Parse(ctx context.Context, raw *document.RawDocument) (*ParsedContent, error)
}

// Registry 按 MIME 类型路由 Parser。
type Registry struct {
	parsers map[string]Parser
}

// NewRegistry 创建空注册表。
func NewRegistry() *Registry {

	return &Registry{parsers: map[string]Parser{}}
}

// Register 注册 MIME 类型到 Parser。
func (r *Registry) Register(mimeType string, parser Parser) {

	r.parsers[strings.ToLower(strings.TrimSpace(mimeType))] = parser
}

// For 返回匹配的 Parser，忽略参数部分（如 charset）。
func (r *Registry) For(mimeType string) (Parser, error) {

	normalized := strings.ToLower(strings.TrimSpace(mimeType))
	if index := strings.Index(normalized, ";"); index >= 0 {
		normalized = strings.TrimSpace(normalized[:index])
	}
	if parser, ok := r.parsers[normalized]; ok {
		return parser, nil
	}
	return nil, rove.NewError("content.unsupported_type", rove.CategoryContent, false, "unsupported content type: %s", mimeType)
}

// DefaultRegistry 返回默认 Parser 注册表（pdf 在 M4 注册）。
func DefaultRegistry() *Registry {

	registry := NewRegistry()
	registry.Register("text/html", NewHTMLParser())
	registry.Register("text/plain", NewTextParser())
	registry.Register("application/json", NewJSONParser())
	registry.Register("text/markdown", NewMarkdownParser())
	registry.Register("application/pdf", NewPDFParser())
	return registry
}
