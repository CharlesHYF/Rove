/*
 * 文件作用：提供 HTTP First 抓取能力 -- redirect 跟随、超时重试、charset 转换、body 上限、SSRF 策略与 Browser Escalation 触发信号。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package fetch 提供 HTTP First 抓取能力，并定义 Browser Escalation 的触发信号。
package fetch

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html/charset"

	"rove/pkg/document"
	"rove/pkg/rove"
)

const (
	defaultMaxBodyBytes = 10 << 20 // 10MB
	defaultTimeout      = 30 * time.Second
	defaultMaxRedirects = 10
	defaultUserAgent    = "rove/0.1"
	minBodyBytes        = 500 // auto 模式正文不足阈值
	maxAttempts         = 3
)

// FetchMode 指定抓取方式。
type FetchMode string

const (
	ModeAuto    FetchMode = "auto"
	ModeHTTP    FetchMode = "http"
	ModeBrowser FetchMode = "browser"
)

// ErrEscalationRequired 是内部控制流信号（非用户可见错误）：auto 模式 HTTP 内容不足，需要浏览器升级（M5）。
var ErrEscalationRequired = errors.New("escalation required: http content insufficient")

// errTooManyRedirects 与 errRedirectBlocked 是 CheckRedirect 的内部 sentinel，用于区分网络错误与策略错误。
var (
	errTooManyRedirects = errors.New("too many redirects")
	errRedirectBlocked  = errors.New("redirect blocked by network policy")
)

// Request 描述一次抓取请求。
type Request struct {
	URL  string
	Mode FetchMode
}

// Options 配置 HTTPFetcher 行为。
type Options struct {
	MaxBodyBytes int64
	Timeout      time.Duration
	MaxRedirects int
	UserAgent    string
	AllowPrivate bool
}

// Fetcher 是获取层统一接口。
type Fetcher interface {
	Fetch(ctx context.Context, req *Request) (*document.RawDocument, error)
}

// HTTPFetcher 基于 net/http 的轻量抓取器。
type HTTPFetcher struct {
	opts   Options
	client *http.Client
}

// NewHTTP 构造 HTTPFetcher，未指定字段使用默认值。
func NewHTTP(opts Options) *HTTPFetcher {

	if opts.MaxBodyBytes <= 0 {
		opts.MaxBodyBytes = defaultMaxBodyBytes
	}
	if opts.Timeout <= 0 {
		opts.Timeout = defaultTimeout
	}
	if opts.MaxRedirects <= 0 {
		opts.MaxRedirects = defaultMaxRedirects
	}
	if opts.UserAgent == "" {
		opts.UserAgent = defaultUserAgent
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	fetcher := &HTTPFetcher{opts: opts}
	fetcher.client = &http.Client{
		Transport: transport,
		Timeout:   opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {

			if len(via) >= opts.MaxRedirects {
				return fmt.Errorf("%w: stopped after %d redirects", errTooManyRedirects, opts.MaxRedirects)
			}
			if !allowURL(req.URL, opts.AllowPrivate) {
				return fmt.Errorf("%w: %s", errRedirectBlocked, req.URL.Host)
			}
			return nil
		},
	}
	return fetcher
}

// Fetch 执行一次抓取。auto 模式在 HTTP 内容不足时返回 ErrEscalationRequired（%w 包装以便 errors.Is）。
func (f *HTTPFetcher) Fetch(ctx context.Context, req *Request) (*document.RawDocument, error) {

	if req.Mode == ModeBrowser {
		return nil, rove.NewError("browser.not_implemented", rove.CategoryBrowser, false, "browser mode 由 M5 里程碑实现")
	}

	u, err := url.Parse(req.URL)
	if err != nil || u.Host == "" {
		return nil, rove.NewError("fetch.invalid_url", rove.CategoryContent, false, "invalid url: %q", req.URL)
	}
	if !allowURL(u, f.opts.AllowPrivate) {
		return nil, rove.NewError("fetch.ssrf_blocked", rove.CategoryPolicy, false, "blocked url (loopback/private): %s", u.Host)
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		raw, err := f.fetchOnce(ctx, u)
		if err == nil {
			if req.Mode == ModeAuto && !f.sufficient(raw) {
				return nil, fmt.Errorf("%w: %s", ErrEscalationRequired, raw.URL)
			}
			return raw, nil
		}
		lastErr = err
		if !isRetryable(err) || attempt == maxAttempts {
			break
		}
		select {
		case <-time.After(backoff(attempt)):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}

// fetchOnce 执行单次 HTTP 请求并组装 RawDocument。
func (f *HTTPFetcher) fetchOnce(ctx context.Context, u *url.URL) (*document.RawDocument, error) {

	var timings document.FetchTimings
	start := time.Now()
	var timeDNS, timeConnect, timeTLS time.Time
	trace := &httptrace.ClientTrace{
		DNSStart:             func(httptrace.DNSStartInfo) { timeDNS = time.Now() },
		DNSDone:              func(httptrace.DNSDoneInfo) { timings.DNS = time.Since(timeDNS) },
		ConnectStart:         func(_, _ string) { timeConnect = time.Now() },
		ConnectDone:          func(_, _ string, _ error) { timings.Connect = time.Since(timeConnect) },
		TLSHandshakeStart:    func() { timeTLS = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { timings.TLS = time.Since(timeTLS) },
		GotFirstResponseByte: func() { timings.TTFB = time.Since(start) },
	}

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, rove.NewError("fetch.request", rove.CategoryContent, false, "build request: %v", err)
	}
	req.Header.Set("User-Agent", f.opts.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/json,text/plain,text/markdown;q=0.9,*/*;q=0.8")

	resp, err := f.client.Do(req)
	if err != nil {
		if errors.Is(err, errTooManyRedirects) {
			return nil, rove.NewError("fetch.redirect_loop", rove.CategoryPolicy, false, "too many redirects (max %d)", f.opts.MaxRedirects)
		}
		if errors.Is(err, errRedirectBlocked) {
			return nil, rove.NewError("fetch.ssrf_blocked", rove.CategoryPolicy, false, "blocked redirect (loopback/private): %s", u.Host)
		}
		return nil, rove.NewError("fetch.network", rove.CategoryNetwork, true, "request %s: %v", u, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, f.opts.MaxBodyBytes+1))
	if err != nil {
		return nil, rove.NewError("fetch.network", rove.CategoryNetwork, true, "read body %s: %v", u, err)
	}
	if int64(len(body)) > f.opts.MaxBodyBytes {
		return nil, rove.NewError("fetch.body_too_large", rove.CategoryContent, false, "body exceeds max %d bytes", f.opts.MaxBodyBytes)
	}

	contentType := resp.Header.Get("Content-Type")
	// charset 探测失败时保留原始字节（降级，不阻断抓取）
	if charsetReader, cerr := charset.NewReader(bytes.NewReader(body), contentType); cerr == nil {
		if decoded, derr := io.ReadAll(charsetReader); derr == nil {
			body = decoded
		}
	}

	timings.Total = time.Since(start)
	return &document.RawDocument{
		ID:          randID(),
		URL:         resp.Request.URL.String(),
		StatusCode:  resp.StatusCode,
		Headers:     resp.Header.Clone(),
		ContentType: contentType,
		Body:        body,
		FetchMethod: document.FetchMethodHTTP,
		Timings:     timings,
		FetchedAt:   time.Now().UTC(),
	}, nil
}

// sufficient 判断 HTTP 正文是否足以进入解析管线（规格书 §5.5）。
func (f *HTTPFetcher) sufficient(raw *document.RawDocument) bool {

	if raw.StatusCode == http.StatusUnauthorized || raw.StatusCode == http.StatusForbidden {
		return false
	}
	if int64(len(raw.Body)) < minBodyBytes {
		return false
	}
	if strings.Contains(raw.ContentType, "html") && len(raw.Body) > 0 {
		scripts := bytes.Count(bytes.ToLower(raw.Body), []byte("<script"))
		if scripts*100/len(raw.Body) > 30 {
			return false
		}
	}
	return true
}

// allowURL 执行 SSRF 策略：默认阻止 loopback/私网/link-local/多播/云 metadata（规格书 §7）。
func allowURL(u *url.URL, allowPrivate bool) bool {

	host := u.Hostname()
	if host == "" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return allowPrivate || !isBlockedIP(ip)
	}
	addrs, err := net.LookupIP(host)
	if err != nil {
		return false // DNS 解析失败时 fail-closed
	}
	for _, addr := range addrs {
		if !allowPrivate && isBlockedIP(addr) {
			return false
		}
	}
	return true
}

// isBlockedIP 判定 IP 是否属于默认阻止范围。
func isBlockedIP(ip net.IP) bool {

	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast() {
		return true
	}
	return ip.Equal(net.ParseIP("169.254.169.254")) // 云 metadata endpoint
}

// isRetryable 判定错误是否值得重试（规格书 §5.6 错误模型）。
func isRetryable(err error) bool {

	var re *rove.Error
	if errors.As(err, &re) {
		return re.Retryable
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout()
	}
	return false
}

// backoff 返回第 attempt 次重试前的退避时长（1s / 4s / 9s）。
func backoff(attempt int) time.Duration {

	return time.Duration(attempt*attempt) * time.Second
}

// randID 生成 16 字节随机 hex 作为 RawDocument 临时 ID。
func randID() string {

	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
