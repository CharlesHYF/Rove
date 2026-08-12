/*
 * 文件作用：检索 JSON schema -- CLI/Agent 消费的稳定输出契约（规格书 §11）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package protocol 定义 Rove 对外接口的 JSON 契约。
package protocol

import (
	"time"

	"rove/pkg/document"
	"rove/pkg/retrieval"
)

// SearchRequest CLI 检索入参。
type SearchRequest struct {
	Query    string
	TopK     int
	Domain   string
	Language string
}

// Hit 单条结果。
type Hit struct {
	URL         string
	Title       string
	Text        string
	Score       float64
	Scores      map[string]float64
	Source      document.Source
	PublishedAt *time.Time
	ChunkID     string
	DocumentID  string
}

// SearchResponse CLI/Agent 检索输出。
type SearchResponse struct {
	Query   string
	Hits    []Hit
	Timings map[string]time.Duration
	TraceID string
}

// FromSearchResult 将 retrieval.SearchResult 映射为对外响应。
func FromSearchResult(result *retrieval.SearchResult) *SearchResponse {

	response := &SearchResponse{
		Query:   result.Query,
		Hits:    make([]Hit, 0, len(result.Hits)),
		Timings: result.Timings,
		TraceID: result.TraceID,
	}
	for _, hit := range result.Hits {
		response.Hits = append(response.Hits, Hit{
			URL:         hit.Document.URL,
			Title:       hit.Document.Title,
			Text:        hit.Chunk.Content,
			Score:       hit.Score,
			Scores:      hit.Scores,
			Source:      hit.Document.Source,
			PublishedAt: hit.Document.PublishedAt,
			ChunkID:     hit.Chunk.ChunkID,
			DocumentID:  hit.Chunk.DocumentID,
		})
	}
	return response
}
