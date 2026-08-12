/*
 * 文件作用：BrowserManager -- Chromium 进程生命周期与页面抓取（HTTP 内容不足时的升级路径，规格书 §5.5）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package browser 基于 chromedp 提供浏览器渲染与页面状态能力。
package browser

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/chromedp/chromedp"

	"rove/pkg/document"
	"rove/pkg/rove"
)

// Options 浏览器配置。
type Options struct {
	ExecutablePath string
	Timeout        time.Duration
}

// Manager 管理 Chromium 进程与会话。
type Manager struct {
	opts        Options
	allocCtx    context.Context
	cancelAlloc context.CancelFunc
}

// NewManager 发现可执行文件并启动 Chromium 分配器。
func NewManager(opts Options) (*Manager, error) {

	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}
	executable, err := DiscoverExecutable(opts.ExecutablePath)
	if err != nil {
		return nil, rove.NewError("browser.not_found", rove.CategoryBrowser, false, "browser executable: %v", err)
	}
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(executable),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	return &Manager{opts: opts, allocCtx: allocCtx, cancelAlloc: cancelAlloc}, nil
}

// Close 释放 Chromium 进程。
func (m *Manager) Close() error {

	m.cancelAlloc()
	return nil
}

// Fetch 用浏览器渲染页面并返回 RawDocument（FetchMethod=browser，进入统一 Pipeline）。
func (m *Manager) Fetch(ctx context.Context, urlStr string) (*document.RawDocument, error) {

	// 每次抓取新建独立 context（隔离 Cookie/Storage，进程复用）；外层叠加超时
	browserCtx, cancel := chromedp.NewContext(m.allocCtx)
	defer cancel()

	pageCtx, cancelTimeout := context.WithTimeout(browserCtx, m.opts.Timeout)
	defer cancelTimeout()

	var html, finalURL string
	actions := []chromedp.Action{
		chromedp.Navigate(urlStr),
		chromedp.WaitReady("html", chromedp.ByQuery),
		chromedp.Evaluate("document.documentElement.outerHTML", &html),
		chromedp.Location(&finalURL),
	}
	if err := chromedp.Run(pageCtx, actions...); err != nil {
		return nil, rove.NewError("browser.navigation", rove.CategoryBrowser, true, "render %s: %v", urlStr, err)
	}
	if html == "" {
		return nil, rove.NewError("browser.navigation", rove.CategoryBrowser, true, "empty rendered html for %s", urlStr)
	}
	return &document.RawDocument{
		ID:          "browser-" + randHex(),
		URL:         finalURL,
		StatusCode:  200, // 浏览器渲染无法可靠取得原始状态码（决策 3）
		ContentType: "text/html; charset=utf-8",
		Body:        []byte(html),
		FetchMethod: document.FetchMethodBrowser,
		FetchedAt:   time.Now().UTC(),
	}, nil
}

// Browse 渲染页面并构建 PageState（交互元素带稳定 DOM 路径 id）。
func (m *Manager) Browse(ctx context.Context, urlStr string) (*PageState, error) {

	browserCtx, cancel := chromedp.NewContext(m.allocCtx)
	defer cancel()

	pageCtx, cancelTimeout := context.WithTimeout(browserCtx, m.opts.Timeout)
	defer cancelTimeout()

	var state PageState
	var raw struct {
		Text        string   `json:"text"`
		Links       []string `json:"links"`
		Interactive []struct {
			ID   string `json:"id"`
			Tag  string `json:"tag"`
			Text string `json:"text"`
			Type string `json:"type"`
			Href string `json:"href"`
		} `json:"interactive"`
	}
	actions := []chromedp.Action{
		chromedp.Navigate(urlStr),
		chromedp.WaitReady("html", chromedp.ByQuery),
		chromedp.Evaluate(pageStateJS, &raw),
		chromedp.Location(&state.URL),
		chromedp.Title(&state.Title),
	}
	if err := chromedp.Run(pageCtx, actions...); err != nil {
		return nil, rove.NewError("browser.navigation", rove.CategoryBrowser, true, "browse %s: %v", urlStr, err)
	}
	state.Text = raw.Text
	state.Links = raw.Links
	for _, element := range raw.Interactive {
		state.Interactive = append(state.Interactive, InteractiveElement{
			ID: element.ID, Tag: element.Tag, Text: element.Text, Type: element.Type, Href: element.Href,
		})
	}
	return &state, nil
}

// DiscoverExecutable 按优先级发现浏览器可执行文件：显式路径 -> 常见路径 -> PATH 搜索。
func DiscoverExecutable(explicit string) (string, error) {

	if explicit != "" {
		if _, err := os.Stat(explicit); err == nil {
			return explicit, nil
		}
		return "", fmt.Errorf("explicit path %q not found", explicit)
	}
	for _, candidate := range commonPaths() {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	for _, name := range []string{"google-chrome", "chromium", "chromium-browser", "msedge", "chrome"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no chrome/edge/chromium executable found (set ROVE_BROWSER_EXECUTABLE)")
}

// commonPaths 返回常见浏览器路径（Windows Edge/Chrome，Linux）。
func commonPaths() []string {

	if runtime.GOOS == "windows" {
		return []string{
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		}
	}
	return []string{"/usr/bin/google-chrome", "/usr/bin/chromium", "/usr/bin/chromium-browser"}
}

// randHex 生成随机 hex。
func randHex() string {

	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
