/*
 * 文件作用：Evidence Engine 的单元测试 -- 字段映射与分数归一化（规格书 §3 / §09）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package evidence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
	"rove/pkg/retrieval"
)

func TestBuild(t *testing.T) {

	now := time.Now().UTC()
	result := &retrieval.SearchResult{
		Query: "browser",
		Hits: []retrieval.SearchHit{
			{
				Chunk:    document.Chunk{ChunkID: "d1-0", DocumentID: "d1", Content: "browser agent infrastructure"},
				Document: &document.Document{ID: "d1", URL: "https://example.com/a", Title: "A", Source: document.Source{Domain: "example.com", Type: "web"}, PublishedAt: &now},
				Score:    8.0,
				Scores:   map[string]float64{"fusion": 0.03, "rank": 8.0, "freshness": 0.9},
			},
			{
				Chunk:    document.Chunk{ChunkID: "d2-0", DocumentID: "d2", Content: "cooking pasta"},
				Document: &document.Document{ID: "d2", URL: "https://example.com/b", Title: "B", Source: document.Source{Domain: "example.com", Type: "web"}},
				Score:    4.0,
				Scores:   map[string]float64{"fusion": 0.01, "rank": 4.0, "freshness": 0.5},
			},
		},
	}

	items, err := New().Build(context.Background(), result)
	require.NoError(t, err)
	require.Len(t, items, 2)

	first := items[0]
	require.Equal(t, "d1", first.DocumentID)
	require.Equal(t, "d1-0", first.ChunkID)
	require.Equal(t, "https://example.com/a", first.URL)
	require.Equal(t, "browser agent infrastructure", first.Text)
	require.Equal(t, 1.0, first.Score, "top hit normalized to 1.0")
	require.NotNil(t, first.PublishedAt)
	require.Equal(t, "example.com", first.Source.Domain)
	require.Contains(t, first.RankFeatures, "freshness")

	second := items[1]
	require.InDelta(t, 0.5, second.Score, 0.0001, "second hit normalized by max")
}

func TestBuildEmpty(t *testing.T) {

	items, err := New().Build(context.Background(), &retrieval.SearchResult{Query: "x"})
	require.NoError(t, err)
	require.Empty(t, items)
}
