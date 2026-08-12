/*
 * 文件作用：定义 Rove 核心数据模型 -- RawDocument / Document / Chunk / Source / FetchTimings 与幂等 ID 生成函数。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package document 定义 Rove 核心数据模型：RawDocument / Document / Chunk。
package document

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

// FetchMethod 标记内容获取方式。
type FetchMethod string

const (
	FetchMethodHTTP    FetchMethod = "http"
	FetchMethodBrowser FetchMethod = "browser"
)

// FetchTimings 记录抓取各阶段耗时（time.Duration 提供纳秒级精度）。
type FetchTimings struct {
	DNS     time.Duration
	Connect time.Duration
	TLS     time.Duration
	TTFB    time.Duration
	Total   time.Duration
}

// RawDocument 是 Fetcher 的输出，保留响应原始信息，不做业务级正文解释。
type RawDocument struct {
	ID          string
	URL         string // 最终 URL（含 redirect 后）
	StatusCode  int
	Headers     http.Header
	ContentType string
	Body        []byte
	FetchMethod FetchMethod
	Timings     FetchTimings
	FetchedAt   time.Time
}

// Source 描述文档来源。
type Source struct {
	Domain string // hostname，不含端口
	Type   string // web | docs | code | academic | ...（默认 web）
}

// Document 是内容管线输出的标准文档。
type Document struct {
	ID           string // 确定性：sha256(canonical_url) 前 16 字节 hex，幂等重抓按此覆盖
	URL          string
	CanonicalURL string
	Title        string
	Description  string
	Content      string // 提取后的正文文本
	Markdown     string // 可读文本视图（chunk 切分的语义来源）
	Language     string // ISO 639-1，如 "en"、"zh"；未知为 ""
	PublishedAt  *time.Time
	UpdatedAt    *time.Time
	FetchedAt    time.Time
	Source       Source
	Metadata     map[string]any
	ContentHash  string // sha256(Content)
	Chunks       []Chunk
}

// Chunk 是检索与 Evidence 的最小单元。
type Chunk struct {
	ChunkID     string // <DocumentID>-<Position>，作为 ES _id 保证幂等
	DocumentID  string
	Position    int    // 文档内序号
	HeadingPath string // 如 "Installation > Quick Start"
	Content     string
	TokenCount  int    // 近似估算
	Language    string
	Embedding   []float32 // 可选；M3 起由 Embedder 填充
	Metadata    map[string]any
}

// NewDocumentID 生成确定性文档 ID：sha256(canonicalURL) 前 16 字节 hex。
func NewDocumentID(canonicalURL string) string {

	sum := sha256.Sum256([]byte(canonicalURL))
	return hex.EncodeToString(sum[:8])
}

// ChunkID 生成确定性 chunk ID。
func ChunkID(documentID string, position int) string {

	return fmt.Sprintf("%s-%d", documentID, position)
}
