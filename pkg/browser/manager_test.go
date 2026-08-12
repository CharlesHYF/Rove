/*
 * 文件作用：pkg/browser BrowserManager 的集成测试 -- SPA 页面 JS 渲染内容抓取（无浏览器时 Skip）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package browser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
)

// testManager 构造浏览器管理器；无可用可执行文件时跳过。
func testManager(t *testing.T) *Manager {

	t.Helper()
	executable, err := DiscoverExecutable("")
	if err != nil {
		t.Skipf("浏览器不可用，跳过: %v", err)
	}
	manager, err := NewManager(Options{ExecutablePath: executable, Timeout: 30 * time.Second})
	require.NoError(t, err)
	return manager
}

func TestBrowserFetchesRenderedHTML(t *testing.T) {

	// SPA 式页面：内容由 JS 渲染（HTTP 正文不足，浏览器可拿到渲染结果）
	page := `<html><head><title>SPA Page</title></head><body><div id="app"></div>
<script>document.getElementById("app").innerHTML = "<h1>Rendered By JS</h1><p>browser rendered content here</p>";</script></body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(page))
	}))
	defer server.Close()

	manager := testManager(t)
	defer manager.Close()

	raw, err := manager.Fetch(context.Background(), server.URL)
	require.NoError(t, err)
	require.Equal(t, document.FetchMethodBrowser, raw.FetchMethod)
	require.Contains(t, string(raw.Body), "Rendered By JS", "browser must return JS-rendered HTML")
	require.NotEmpty(t, raw.URL)
}
