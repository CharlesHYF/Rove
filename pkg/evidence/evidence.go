/*
 * 文件作用：Evidence Engine -- 将排序后的检索结果转为 Agent 可消费的可追溯证据（规格书 §3 / §09）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package evidence 提供可追溯的证据输出，不生成业务结论。
package evidence

import (
	"context"
	"time"

	"rove/pkg/document"
	"rove/pkg/retrieval"
)

// Evidence 是 Agent 消费的最终检索单元。
type Evidence struct {
	DocumentID   string
	ChunkID      string
	URL          string
	Title        string
	Text         string
	Score        float64 // 归一化 0-1
	RankFeatures map[string]float64
	PublishedAt  *time.Time
	Source       document.Source
	Metadata     map[string]any
}

// Engine 构建 Evidence。
type Engine struct{}

// New 构造 Evidence Engine。
func New() *Engine {

	return &Engine{}
}

// Build 将检索结果转换为 Evidence，分数按最高分归一化到 0-1。
func (e *Engine) Build(ctx context.Context, result *retrieval.SearchResult) ([]Evidence, error) {

	if len(result.Hits) == 0 {
		return []Evidence{}, nil
	}

	maxScore := result.Hits[0].Score
	for _, hit := range result.Hits {
		if hit.Score > maxScore {
			maxScore = hit.Score
		}
	}
	if maxScore <= 0 {
		maxScore = 1
	}

	items := make([]Evidence, 0, len(result.Hits))
	for _, hit := range result.Hits {
		items = append(items, Evidence{
			DocumentID:   hit.Chunk.DocumentID,
			ChunkID:      hit.Chunk.ChunkID,
			URL:          hit.Document.URL,
			Title:        hit.Document.Title,
			Text:         hit.Chunk.Content,
			Score:        hit.Score / maxScore,
			RankFeatures: hit.Scores,
			PublishedAt:  hit.Document.PublishedAt,
			Source:       hit.Document.Source,
			Metadata:     hit.Document.Metadata,
		})
	}
	return items, nil
}
