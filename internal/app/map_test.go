/*
 * 文件作用：Map 链接发现服务测试 -- 站内/站外划分、去重与 canonical。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/internal/config"
)

// newMapTestServer 托管一个含站内/站外链接的页面。
func newMapTestServer() *httptest.Server {

	page := `<html><head><title>Map Test</title><link rel="canonical" href="https://example.com/"></head><body>
		<a href="/a">站内一</a>
		<a href="/a">站内一重复</a>
		<a href="/b/c?x=1">站内二</a>
		<a href="https://other.com/page">站外</a>
	</body></html>`
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(page))
	}))
}

func TestMapDiscoversLinks(t *testing.T) {

	ts := newMapTestServer()
	defer ts.Close()

	cfg := config.Default()
	cfg.Fetch.AllowPrivate = true
	result, err := NewMap(cfg).Map(context.Background(), ts.URL)

	require.NoError(t, err)
	require.Contains(t, result.URL, ts.URL, "抓取地址应保留")
	require.Equal(t, "https://example.com/", result.Canonical, "canonical 来自 rel=canonical")
	require.Contains(t, result.InSite, ts.URL+"/a")
	require.Contains(t, result.InSite, ts.URL+"/b/c?x=1")
	require.Contains(t, result.OutSite, "https://other.com/page")
	require.Len(t, result.InSite, 2, "重复链接应去重")
}

func TestMapUnsupportedContent(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("binary"))
	}))
	defer ts.Close()

	cfg := config.Default()
	cfg.Fetch.AllowPrivate = true
	_, err := NewMap(cfg).Map(context.Background(), ts.URL)

	require.Error(t, err, "不支持的内容类型应返回错误")
}
