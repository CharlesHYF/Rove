/*
 * 文件作用：端到端验收测试 -- fetch -> pipeline -> index -> search 全链路（规格书 §9 A1 前半段 + A3）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package e2e

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/content"
	"rove/pkg/elastic"
	"rove/pkg/evidence"
	"rove/pkg/fetch"
	"rove/pkg/index"
	"rove/pkg/retrieval"
)

func TestFetchIndexSearch(t *testing.T) {

	// 1. 托管一个 HTML 测试页
	page := `<html lang="en"><head><title>Rove E2E Doc</title></head><body><article><h1>Browser Agents</h1>` +
		`<p>Rove is an open web infrastructure for AI agents, providing crawl browse index retrieve rank evidence.</p>` +
		`<p>Browser agents need reliable web access to the open web.</p></article></body></html>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(page))
	}))
	defer ts.Close()

	// 2. ES 就绪
	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	prefix := "rove-e2e-" + randSuffix()
	client, err := elastic.New(url, "", "", prefix)
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))

	// 3. fetch -> pipeline -> index（全链路，嵌入向量）
	fetcher := fetch.NewHTTP(fetch.Options{AllowPrivate: true})
	raw, err := fetcher.Fetch(context.Background(), &fetch.Request{URL: ts.URL, Mode: fetch.ModeHTTP})
	require.NoError(t, err)

	pipeline := content.NewPipeline(content.DefaultRegistry(), &content.Canonicalizer{}, content.NewDeduper(), content.NewChunker(800, 100))
	doc, dedup, err := pipeline.Process(context.Background(), raw, "web")
	require.NoError(t, err)
	require.False(t, dedup.IsDuplicate)
	require.NotEmpty(t, doc.Chunks)

	indexer := index.New(client, retrieval.NewPseudoEmbedder())
	require.NoError(t, indexer.Index(context.Background(), doc))

	// 4. Hybrid 检索（无 API Key，验收 A4/A7）
	retriever := retrieval.New(client, retrieval.NewPseudoEmbedder())
	result, err := retriever.Search(context.Background(), &retrieval.Query{Text: "browser agents", TopK: 5})
	require.NoError(t, err)
	require.Len(t, result.Hits, 1)
	require.Equal(t, "Rove E2E Doc", result.Hits[0].Document.Title)
	require.Contains(t, result.Hits[0].Chunk.Content, "open web infrastructure")
	require.Contains(t, result.Hits[0].Scores, "vector", "vector leg must contribute (A4)")
	require.Contains(t, result.Hits[0].Scores, "fusion")
	require.Contains(t, result.Timings, "vector_retrieve")

	// 5. Evidence（验收 A5：结果可作为 Evidence 使用）
	items, err := evidence.New().Build(context.Background(), result)
	require.NoError(t, err)
	require.Len(t, items, 1)
	first := items[0]
	require.Equal(t, "Rove E2E Doc", first.Title)
	require.Contains(t, first.Text, "open web infrastructure")
	require.NotEmpty(t, first.URL)
	require.InDelta(t, 1.0, first.Score, 0.0001, "single hit normalized to 1.0")
	require.Equal(t, "web", first.Source.Type)
}

// randSuffix 生成测试索引唯一后缀。
func randSuffix() string {

	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
