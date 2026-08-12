/*
 * 文件作用：internal/app 服务的组装冒烟测试 -- 索引 -> 检索全链路（ES 不可达时跳过）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package app

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
	"rove/pkg/elastic"
	"rove/pkg/index"
	"rove/pkg/retrieval"
)

func TestServicesSmoke(t *testing.T) {

	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	client, err := elastic.New(url, "", "", "rove-test-app")
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))

	ctx := context.Background()
	indexService := NewIndex(index.New(client))
	searchService := NewSearch(retrieval.New(client))

	now := time.Now().UTC()
	doc := &document.Document{
		ID: "app-1", URL: "https://x.com/1", CanonicalURL: "https://x.com/1", Title: "App Smoke",
		Content: "app service smoke test", Markdown: "# App Smoke", Language: "en", FetchedAt: now,
		Source: document.Source{Domain: "x.com", Type: "web"}, ContentHash: "h",
		Chunks: []document.Chunk{{ChunkID: "app-1-0", DocumentID: "app-1", Position: 0, Content: "app service smoke test", TokenCount: 4, Language: "en"}},
	}
	require.NoError(t, indexService.Index(ctx, doc))

	result, err := searchService.Search(ctx, &retrieval.Query{Text: "smoke", TopK: 5})
	require.NoError(t, err)
	require.Len(t, result.Hits, 1)
	require.Equal(t, "App Smoke", result.Hits[0].Document.Title)
}
