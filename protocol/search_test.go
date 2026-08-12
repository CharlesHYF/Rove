/*
 * 文件作用：protocol 检索映射的单元测试 -- 内部 SearchResult 到对外响应的字段契约。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package protocol

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
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
				Scores:   map[string]float64{"lexical": 3.5},
			},
		},
	}

	response := FromSearchResult(result)
	require.Equal(t, "browser", response.Query)
	require.Equal(t, "trace-1", response.TraceID)
	require.Len(t, response.Hits, 1)

	hit := response.Hits[0]
	require.Equal(t, "https://example.com/a", hit.URL)
	require.Equal(t, "browser agent", hit.Text)
	require.Equal(t, "example.com", hit.Source.Domain)
	require.Equal(t, "d1-0", hit.ChunkID)
	require.Equal(t, float64(3.5), hit.Score)
}
