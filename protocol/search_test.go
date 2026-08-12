/*
 * 文件作用：protocol 检索映射的单元测试 -- 内部 SearchResult + Evidence 到对外响应的字段契约。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package protocol

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
	"rove/pkg/evidence"
	"rove/pkg/retrieval"
)

func TestFromSearchResult(t *testing.T) {

	now := time.Now().UTC()
	result := &retrieval.SearchResult{
		Query:   "browser",
		TraceID: "trace-1",
		Timings: map[string]time.Duration{"lexical_retrieve": 12 * time.Millisecond},
		Hits: []retrieval.SearchHit{
			{
				Chunk:    document.Chunk{ChunkID: "d1-0", DocumentID: "d1", Content: "browser agent"},
				Document: &document.Document{ID: "d1", URL: "https://example.com/a", Title: "A", Source: document.Source{Domain: "example.com", Type: "web"}, PublishedAt: &now},
				Score:    3.5,
				Scores:   map[string]float64{"fusion": 0.03},
			},
		},
	}
	items := []evidence.Evidence{
		{
			DocumentID: "d1", ChunkID: "d1-0", URL: "https://example.com/a", Title: "A",
			Text: "browser agent", Score: 1.0, RankFeatures: map[string]float64{"fusion": 0.03}, PublishedAt: &now,
			Source: document.Source{Domain: "example.com", Type: "web"},
		},
	}

	response := FromSearchResult(result, items)
	require.Equal(t, "browser", response.Query)
	require.Equal(t, "trace-1", response.TraceID)
	require.Len(t, response.Hits, 1)
	require.Len(t, response.Evidence, 1)
	require.Equal(t, "d1-0", response.Evidence[0].ChunkID)
	require.Equal(t, 1.0, response.Evidence[0].Score)
	require.Contains(t, response.Evidence[0].RankFeatures, "fusion")

	hit := response.Hits[0]
	require.Equal(t, "https://example.com/a", hit.URL)
	require.Equal(t, "browser agent", hit.Text)
	require.Equal(t, "example.com", hit.Source.Domain)
	require.Equal(t, "d1-0", hit.ChunkID)
}
