/*
 * 文件作用：内容管线编排 -- 将 RawDocument 转换为标准 Document：解析 -> 规范化 -> 去重 -> 分块。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/url"

	"rove/pkg/document"
)

// DedupResult 描述去重判定结果。
type DedupResult struct {
	IsDuplicate bool
	Reason      string
}

// Pipeline 将 RawDocument 转换为标准 Document。
type Pipeline struct {
	registry      *Registry
	canonicalizer *Canonicalizer
	deduper       *Deduper
	chunker       *Chunker
}

// NewPipeline 组装内容管线。
func NewPipeline(registry *Registry, canonicalizer *Canonicalizer, deduper *Deduper, chunker *Chunker) *Pipeline {

	return &Pipeline{
		registry:      registry,
		canonicalizer: canonicalizer,
		deduper:       deduper,
		chunker:       chunker,
	}
}

// Process 执行完整内容管线。去重命中时 Document.Chunks 为空，由调用方决定跳过索引（M2/M4 使用）。
func (p *Pipeline) Process(ctx context.Context, raw *document.RawDocument, sourceType string) (*document.Document, *DedupResult, error) {

	parser, err := p.registry.For(raw.ContentType)
	if err != nil {
		return nil, nil, err
	}

	parsed, err := parser.Parse(ctx, raw)
	if err != nil {
		return nil, nil, err
	}

	canonical, err := p.canonicalizer.Canonicalize(ctx, raw.URL, parsed.CanonicalURL)
	if err != nil {
		return nil, nil, err
	}

	content := parsed.Content
	if content == "" {
		content = parsed.Markdown
	}

	doc := &document.Document{
		ID:           document.NewDocumentID(canonical),
		URL:          raw.URL,
		CanonicalURL: canonical,
		Title:        parsed.Title,
		Description:  parsed.Description,
		Content:      content,
		Markdown:     parsed.Markdown,
		Language:     parsed.Language,
		PublishedAt:  parsed.PublishedAt,
		UpdatedAt:    parsed.UpdatedAt,
		FetchedAt:    raw.FetchedAt,
		Source:       document.Source{Domain: Hostname(raw.URL), Type: sourceType},
		Metadata:     parsed.Metadata,
		ContentHash:  HashContent(content),
	}

	dup, reason, err := p.deduper.IsDuplicate(ctx, doc)
	if err != nil {
		return nil, nil, err
	}
	if dup {
		return doc, &DedupResult{IsDuplicate: true, Reason: reason}, nil
	}

	chunks, err := p.chunker.Chunk(ctx, doc)
	if err != nil {
		return nil, nil, err
	}
	doc.Chunks = chunks
	return doc, &DedupResult{}, nil
}

// HashContent 计算内容 sha256 hex。
func HashContent(text string) string {

	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

// Hostname 提取 URL 的 hostname，失败时返回 "unknown"。
func Hostname(rawURL string) string {

	u, err := url.Parse(rawURL)
	if err != nil {
		return "unknown"
	}
	if u.Hostname() == "" {
		return "unknown"
	}
	return u.Hostname()
}
