/*
 * 文件作用：cmd/rove 浏览器升级的集成测试 -- browser 模式与 auto 模式自动升级（验收 A2）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/browser"
)

// spaShell 返回 HTTP 正文不足、需 JS 渲染的页面。
func spaShell() string {

	return `<html><head><title>SPA Shell</title></head><body><div id="app"><script>
document.getElementById("app").innerHTML = "<h1>Dynamic Title</h1><p>dynamic body content rendered by javascript</p>";</script></div></body></html>`
}

// spaServer 返回输出 SPA shell 页面的测试服务器。
func spaServer(t *testing.T) *httptest.Server {

	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(spaShell()))
	}))
}

func TestFetchBrowserMode(t *testing.T) {

	if _, err := browser.DiscoverExecutable(""); err != nil {
		t.Skipf("浏览器不可用，跳过: %v", err)
	}
	ts := spaServer(t)
	defer ts.Close()

	t.Setenv("ROVE_ALLOW_PRIVATE", "true")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"fetch", ts.URL, "--mode", "browser", "--json"})
	require.NoError(t, rootCmd.Execute())
	require.Contains(t, out.String(), "Dynamic Title", "browser mode must return rendered content")
}

func TestFetchAutoEscalation(t *testing.T) {

	if _, err := browser.DiscoverExecutable(""); err != nil {
		t.Skipf("浏览器不可用，跳过: %v", err)
	}
	ts := spaServer(t)
	defer ts.Close()

	t.Setenv("ROVE_ALLOW_PRIVATE", "true")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"fetch", ts.URL, "--mode", "auto", "--json"})
	require.NoError(t, rootCmd.Execute())
	require.Contains(t, out.String(), "Dynamic Title", "auto mode must escalate to browser (A2)")
}
