/*
 * 文件作用：pkg/crawler robots 的单元测试 -- 放行/拒绝、crawl-delay 与缺失文件 fail-open。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/fetch"
)

func TestRobotsIsAllowed(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("User-agent: *\nDisallow: /private/\nCrawl-delay: 2\n"))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := NewRobotsClient(fetch.NewHTTP(fetch.Options{AllowPrivate: true}))
	ctx := context.Background()

	allowed, err := client.IsAllowed(ctx, ts.URL+"/public/page")
	require.NoError(t, err)
	require.True(t, allowed, "public path must be allowed")

	allowed, err = client.IsAllowed(ctx, ts.URL+"/private/secret")
	require.NoError(t, err)
	require.False(t, allowed, "disallowed path must be blocked")

	delay, err := client.CrawlDelay(ctx, hostOf(ts.URL))
	require.NoError(t, err)
	require.Equal(t, 2*time.Second, delay)
}

func TestRobotsMissingFile(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := NewRobotsClient(fetch.NewHTTP(fetch.Options{AllowPrivate: true}))
	allowed, err := client.IsAllowed(context.Background(), ts.URL+"/x")
	require.NoError(t, err)
	require.True(t, allowed, "missing robots.txt must fail-open")
}

// hostOf 提取 URL 的 hostname。
func hostOf(urlStr string) string {

	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}
