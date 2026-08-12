/*
 * 文件作用：cmd/rove fetch 命令的端到端测试 -- 验证 CLI 装配 Fetcher/Pipeline 后的 JSON 与人类可读输出。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// htmlServer 返回输出固定 HTML 页面的测试服务器（正文 > 500 字节，避免触发 auto 模式 escalation）。
func htmlServer(t *testing.T) *httptest.Server {

	t.Helper()
	body := `<article><h1>CLI</h1><p>cli body content for testing</p>` + strings.Repeat("<p>additional paragraph content to keep the page body large enough for auto mode</p>", 20) + `</article>`
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html lang="en"><head><title>CLI Test Page</title></head><body>` + body + `</body></html>`))
	}))
}

// runCLI 以给定参数执行 rootCmd 并捕获输出。
func runCLI(t *testing.T, args ...string) string {

	t.Helper()
	t.Setenv("ROVE_ALLOW_PRIVATE", "true") // httptest 在 loopback 上
	fetchFlags.mode = "auto"
	fetchFlags.json = false

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	require.NoError(t, rootCmd.Execute())
	return out.String()
}

func TestFetchCommandJSON(t *testing.T) {

	ts := htmlServer(t)
	defer ts.Close()

	out := runCLI(t, "fetch", ts.URL, "--json")
	require.Contains(t, out, `"Title": "CLI Test Page"`, "JSON output must contain title")
	require.Contains(t, out, `"Chunks":`, "JSON output must contain chunks")
}

func TestFetchCommandHuman(t *testing.T) {

	ts := htmlServer(t)
	defer ts.Close()

	out := runCLI(t, "fetch", ts.URL)
	require.Contains(t, out, "title: CLI Test Page", "human output must contain title")
	require.Contains(t, out, "chunks: ", "human output must contain chunk count")
}
