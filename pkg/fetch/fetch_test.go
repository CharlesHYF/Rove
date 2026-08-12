/*
 * 文件作用：pkg/fetch 的单元测试 -- 覆盖 redirect、重试、body 上限、SSRF 策略、charset 转换与 escalation 信号。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package fetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// testClient 返回可访问 httptest（loopback）的 fetcher。
func testClient(t *testing.T, opts Options) *HTTPFetcher {

	t.Helper()
	if !opts.AllowPrivate {
		opts.AllowPrivate = true
	}
	return NewHTTP(opts)
}

func TestFetchFollowsRedirect(t *testing.T) {

	var final string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/end", http.StatusMovedPermanently)
			return
		}
		final = r.URL.Path
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("hello"))
	}))
	defer ts.Close()

	raw, err := testClient(t, Options{}).Fetch(context.Background(), &Request{URL: ts.URL + "/start", Mode: ModeHTTP})
	require.NoError(t, err)
	require.Equal(t, "/end", final, "redirect not followed")
	require.Equal(t, http.StatusOK, raw.StatusCode)
	require.True(t, strings.HasSuffix(raw.URL, "/end"), "raw.URL must be the final URL")
	require.Positive(t, raw.Timings.Total, "Timings.Total must be positive")
}

func TestFetchMaxRedirects(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.Path, http.StatusFound)
	}))
	defer ts.Close()

	_, err := testClient(t, Options{MaxRedirects: 5}).Fetch(context.Background(), &Request{URL: ts.URL + "/loop", Mode: ModeHTTP})
	require.Error(t, err)
	require.Contains(t, err.Error(), "redirect", "redirect loop must error")
}

func TestFetchMaxBody(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(make([]byte, 2048))
	}))
	defer ts.Close()

	_, err := testClient(t, Options{MaxBodyBytes: 1024}).Fetch(context.Background(), &Request{URL: ts.URL, Mode: ModeHTTP})
	require.Error(t, err)
	require.Contains(t, err.Error(), "body", "want body too large error")
}

func TestFetchSSRFBlockLoopback(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	fetcher := NewHTTP(Options{}) // AllowPrivate=false（默认）
	_, err := fetcher.Fetch(context.Background(), &Request{URL: ts.URL, Mode: ModeHTTP})
	require.Error(t, err)
	require.Contains(t, err.Error(), "blocked", "loopback must be blocked by default")
}

func TestFetchCharsetGBK(t *testing.T) {

	// "你好" 的 GBK 编码字节
	gbk := []byte{0xc4, 0xe3, 0xba, 0xc3}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=gbk")
		_, _ = w.Write(gbk)
	}))
	defer ts.Close()

	raw, err := testClient(t, Options{}).Fetch(context.Background(), &Request{URL: ts.URL, Mode: ModeHTTP})
	require.NoError(t, err)
	require.Equal(t, "你好", string(raw.Body), "gbk body must be converted to UTF-8")
}

func TestFetchAutoInsufficient(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><head></head><body><script>window.__APP__={}</script></body></html>"))
	}))
	defer ts.Close()

	_, err := testClient(t, Options{}).Fetch(context.Background(), &Request{URL: ts.URL, Mode: ModeAuto})
	require.ErrorIs(t, err, ErrEscalationRequired)
}

func TestFetchModeBrowserNotImplemented(t *testing.T) {

	_, err := testClient(t, Options{}).Fetch(context.Background(), &Request{URL: "https://example.com", Mode: ModeBrowser})
	require.Error(t, err)
	require.Contains(t, err.Error(), "M5", "browser mode must return M5 not-implemented error")
}

func TestFetchRetriesNetworkError(t *testing.T) {

	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			hijacker, ok := w.(http.Hijacker)
			if !ok {
				return
			}
			conn, _, err := hijacker.Hijack()
			if err != nil {
				return
			}
			_ = conn.Close() // 模拟连接被重置
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	raw, err := testClient(t, Options{Timeout: 5 * time.Second}).Fetch(context.Background(), &Request{URL: ts.URL, Mode: ModeHTTP})
	require.NoError(t, err)
	require.Equal(t, "ok", string(raw.Body))
	require.GreaterOrEqual(t, calls, 3, "expected retries")
}
