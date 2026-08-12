/*
 * 文件作用：E2E 浏览器升级验收 -- SPA 渲染进入同一 Document Pipeline 并可检索（验收 A2）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/browser"
	"rove/pkg/content"
	"rove/pkg/elastic"
	"rove/pkg/index"
	"rove/pkg/retrieval"
)

// TestBrowserEscalationPipeline 验收 A2：浏览器渲染结果进入同一 Document Pipeline 并可检索。
func TestBrowserEscalationPipeline(t *testing.T) {

	executable, err := browser.DiscoverExecutable("")
	if err != nil {
		t.Skipf("浏览器不可用，跳过: %v", err)
	}
	page := `<html><head><title>SPA E2E</title></head><body><div id="app"><script>
document.getElementById("app").innerHTML = "<h1>SPA Content</h1><p>rendered by javascript for the e2e test</p>";</script></div></body></html>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(page))
	}))
	defer ts.Close()

	manager, err := browser.NewManager(browser.Options{ExecutablePath: executable, Timeout: 30 * time.Second})
	require.NoError(t, err)
	defer manager.Close()

	raw, err := manager.Fetch(context.Background(), ts.URL)
	require.NoError(t, err)

	// 同一 Document Pipeline
	pipeline := content.NewPipeline(content.DefaultRegistry(), &content.Canonicalizer{}, content.NewDeduper(), content.NewChunker(800, 100))
	doc, dedup, err := pipeline.Process(context.Background(), raw, "web")
	require.NoError(t, err)
	require.False(t, dedup.IsDuplicate)
	require.Contains(t, doc.Title, "SPA E2E")
	require.Contains(t, doc.Content, "rendered by javascript")

	// 索引 + 检索
	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	client, err := elastic.New(url, "", "", "rove-e2e-browser")
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过索引段: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))
	require.NoError(t, index.New(client, retrieval.NewPseudoEmbedder()).Index(context.Background(), doc))

	result, err := retrieval.New(client, retrieval.NewPseudoEmbedder()).Search(context.Background(), &retrieval.Query{Text: "javascript e2e", TopK: 5})
	require.NoError(t, err)
	require.NotEmpty(t, result.Hits, "browser-rendered content must be searchable")
	require.Equal(t, "SPA E2E", result.Hits[0].Document.Title)
}
