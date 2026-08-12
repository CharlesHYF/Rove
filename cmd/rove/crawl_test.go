/*
 * 文件作用：cmd/rove crawl 命令的集成测试 -- 站点图抓取 -> 索引 -> 搜索（验收 A1 + A3）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/elastic"
)

func TestCrawlCommand(t *testing.T) {

	site := map[string]string{
		"/":     `<html><head><title>Crawl Home</title></head><body><p>crawl home content here</p><a href="/page">page</a></body></html>`,
		"/page": `<html><head><title>Crawl Page</title></head><body><p>crawl page content here</p></body></html>`,
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

	client, err := elastic.New(esURL(), "", "", "rove-crawl-test")
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))

	t.Setenv("ROVE_ES_URL", esURL())
	t.Setenv("ROVE_INDEX_PREFIX", "rove-crawl-test")
	t.Setenv("ROVE_ALLOW_PRIVATE", "true")
	t.Setenv("ROVE_CRAWL_STATE_DB", filepath.Join(t.TempDir(), "crawl.db"))

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"crawl", ts.URL + "/", "--max-pages", "5"})
	require.NoError(t, rootCmd.Execute())
	require.Contains(t, out.String(), "visited")
	require.Contains(t, out.String(), "indexed")

	// 验证已索引可检索（A1 + A3）
	var searchOut bytes.Buffer
	rootCmd.SetOut(&searchOut)
	rootCmd.SetArgs([]string{"search", "crawl home", "--json"})
	require.NoError(t, rootCmd.Execute())
	require.Contains(t, searchOut.String(), "Crawl Home")
}
