/*
 * 文件作用：pkg/retrieval 的集成测试 -- BM25 召回、document 聚合、分数分解与过滤（规格书 §4.2）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package retrieval

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
	"rove/pkg/elastic"
	"rove/pkg/index"
)

// testSetup 返回隔离的 Retriever 与 ES 客户端；ES 不可达时跳过。
func testSetup(t *testing.T) (*Retriever, *elastic.Client) {

	t.Helper()
	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	prefix := "rove-test-" + randSuffix()
	client, err := elastic.New(url, "", "", prefix)
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过集成测试: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))
	return New(client), client
}

// randSuffix 生成测试索引唯一后缀。
func randSuffix() string {

	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func TestSearchBM25(t *testing.T) {

	ctx := context.Background()
	retriever, client := testSetup(t)
	indexer := index.New(client)

	now := time.Now().UTC()
	docs := []*document.Document{
		{
			ID: "d1", URL: "https://example.com/browser-agent", CanonicalURL: "https://example.com/browser-agent",
			Title: "Browser Agent Guide", Content: "browser agent infrastructure for AI", Markdown: "# Browser Agent Guide",
			Language: "en", FetchedAt: now, Source: document.Source{Domain: "example.com", Type: "web"}, ContentHash: "h1",
			Chunks: []document.Chunk{{ChunkID: "d1-0", DocumentID: "d1", Position: 0, Content: "browser agent infrastructure for AI", TokenCount: 6, Language: "en"}},
		},
		{
			ID: "d2", URL: "https://other.com/cooking", CanonicalURL: "https://other.com/cooking",
			Title: "Cooking Recipes", Content: "how to cook pasta", Markdown: "# Cooking",
			Language: "en", FetchedAt: now, Source: document.Source{Domain: "other.com", Type: "web"}, ContentHash: "h2",
			Chunks: []document.Chunk{{ChunkID: "d2-0", DocumentID: "d2", Position: 0, Content: "how to cook pasta", TokenCount: 4, Language: "en"}},
		},
	}
	require.NoError(t, indexer.IndexBulk(ctx, docs))

	result, err := retriever.Search(ctx, &Query{Text: "browser agent", TopK: 5})
	require.NoError(t, err)
	require.Len(t, result.Hits, 1, "only d1 matches browser agent")
	require.Equal(t, "d1", result.Hits[0].Document.ID)
	require.Equal(t, "Browser Agent Guide", result.Hits[0].Document.Title)
	require.Equal(t, "browser agent infrastructure for AI", result.Hits[0].Chunk.Content)
	require.Greater(t, result.Hits[0].Score, float64(0))
	require.Contains(t, result.Hits[0].Scores, "lexical")
	require.Contains(t, result.Hits[0].Scores, "freshness")
	require.NotEmpty(t, result.TraceID)
}

func TestSearchDomainFilter(t *testing.T) {

	ctx := context.Background()
	retriever, client := testSetup(t)
	indexer := index.New(client)

	now := time.Now().UTC()
	makeDoc := func(id, domain, content string) *document.Document {
		return &document.Document{
			ID: id, URL: "https://" + domain + "/p", CanonicalURL: "https://" + domain + "/p",
			Title: id, Content: content, Markdown: id,
			Language: "en", FetchedAt: now, Source: document.Source{Domain: domain, Type: "web"}, ContentHash: "h" + id,
			Chunks: []document.Chunk{{ChunkID: id + "-0", DocumentID: id, Position: 0, Content: content, TokenCount: 3, Language: "en"}},
		}
	}
	require.NoError(t, indexer.IndexBulk(ctx, []*document.Document{
		makeDoc("da", "a.com", "shared keyword alpha"),
		makeDoc("db", "b.com", "shared keyword beta"),
	}))

	result, err := retriever.Search(ctx, &Query{Text: "shared", TopK: 5, Filters: Filters{Domain: "a.com"}})
	require.NoError(t, err)
	require.Len(t, result.Hits, 1)
	require.Equal(t, "da", result.Hits[0].Document.ID)
}
