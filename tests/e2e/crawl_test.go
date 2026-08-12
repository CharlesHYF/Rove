/*
 * 文件作用：E2E 抓取验收 -- seed -> crawl -> index -> search（验收 A1，独立索引前缀）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/internal/scheduler"
	"rove/pkg/content"
	"rove/pkg/crawler"
	"rove/pkg/elastic"
	"rove/pkg/fetch"
	"rove/pkg/index"
	"rove/pkg/retrieval"
)

// TestCrawlEndToEnd 验收 A1：seed -> discover -> fetch -> parse -> extract -> dedup -> index -> search。
func TestCrawlEndToEnd(t *testing.T) {

	site := map[string]string{
		"/":  `<html><head><title>E2E Home</title></head><body><p>e2e home content here</p><a href="/a">a</a><a href="/b">b</a></body></html>`,
		"/a": `<html><head><title>E2E Page A</title></head><body><p>page a content here</p></body></html>`,
		"/b": `<html><head><title>E2E Page B</title></head><body><p>page b content here</p></body></html>`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, ok := site[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(body))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	client, err := elastic.New(url, "", "", "rove-e2e-crawl")
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))

	frontier, err := scheduler.NewFrontier(filepath.Join(t.TempDir(), "crawl.db"))
	require.NoError(t, err)
	defer frontier.Close()

	sched := scheduler.NewScheduler(frontier, scheduler.Options{HostConcurrency: 2})
	fetcher := fetch.NewHTTP(fetch.Options{AllowPrivate: true})
	registry := content.DefaultRegistry()
	pipeline := content.NewPipeline(registry, &content.Canonicalizer{}, content.NewDeduper(), content.NewChunker(800, 100))
	indexer := index.New(client, retrieval.NewPseudoEmbedder())

	c := crawler.New(frontier, sched, crawler.NewRobotsClient(fetcher),
		crawler.Policy{SameHostOnly: true, MaxDepth: 3},
		fetcher, registry, pipeline, indexer, 3, 10)

	stats, err := c.Run(context.Background(), []string{ts.URL + "/"})
	require.NoError(t, err)
	require.GreaterOrEqual(t, stats.Visited, 3, "home + a + b")

	result, err := retrieval.New(client, retrieval.NewPseudoEmbedder()).Search(context.Background(), &retrieval.Query{Text: "e2e home", TopK: 5})
	require.NoError(t, err)
	require.NotEmpty(t, result.Hits, "crawled page must be searchable")
	require.Equal(t, "E2E Home", result.Hits[0].Document.Title)
}
