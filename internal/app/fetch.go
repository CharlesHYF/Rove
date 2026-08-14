/*
 * 文件作用：Application Service -- Fetch 抓取服务，供 CLI 与 MCP 复用（HTTP First + Browser Escalation，验收 A2）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package app

import (
	"context"
	"errors"
	"fmt"

	"rove/internal/config"
	"rove/pkg/browser"
	"rove/pkg/content"
	"rove/pkg/document"
	"rove/pkg/fetch"
	"rove/pkg/rove"
)

// FetchService 抓取服务：HTTP First，auto 模式内容不足时自动浏览器升级。
type FetchService struct {
	cfg *config.Config
}

// NewFetch 按配置构造抓取服务。
func NewFetch(cfg *config.Config) *FetchService {

	return &FetchService{cfg: cfg}
}

// Fetch 抓取 url 并经内容管线输出标准 Document（含去重判定）。
func (s *FetchService) Fetch(ctx context.Context, url string, mode fetch.FetchMode) (*document.Document, *content.DedupResult, error) {

	raw, err := s.fetchRaw(ctx, url, mode)
	if err != nil {
		return nil, nil, err
	}
	return newPipeline(s.cfg).Process(ctx, raw, "web")
}

// fetchRaw 获取原始内容：browser 模式直接渲染；auto 模式 HTTP 内容不足时升级浏览器。
func (s *FetchService) fetchRaw(ctx context.Context, url string, mode fetch.FetchMode) (*document.RawDocument, error) {

	if mode == fetch.ModeBrowser {
		manager, managerErr := newBrowserManager(s.cfg)
		if managerErr != nil {
			return nil, managerErr
		}
		defer manager.Close()
		return manager.Fetch(ctx, url)
	}
	raw, err := newHTTPFetcher(s.cfg).Fetch(ctx, &fetch.Request{URL: url, Mode: mode})
	if err != nil && errors.Is(err, fetch.ErrEscalationRequired) && mode == fetch.ModeAuto {
		manager, browserErr := newBrowserManager(s.cfg)
		if browserErr != nil {
			return nil, rove.NewError("fetch.escalation_required", rove.CategoryBrowser, false,
				"页面需要浏览器渲染:%s(%v)", url, browserErr)
		}
		defer manager.Close()
		return manager.Fetch(ctx, url)
	}
	return raw, err
}

// newHTTPFetcher 按配置构造 HTTP Fetcher。
func newHTTPFetcher(cfg *config.Config) *fetch.HTTPFetcher {

	return fetch.NewHTTP(fetch.Options{
		MaxBodyBytes: cfg.Fetch.MaxBodyBytes,
		Timeout:      cfg.Fetch.Timeout.Duration,
		MaxRedirects: cfg.Fetch.MaxRedirects,
		UserAgent:    cfg.Fetch.UserAgent,
		AllowPrivate: cfg.Fetch.AllowPrivate,
	})
}

// newPipeline 按配置组装内容管线。
func newPipeline(cfg *config.Config) *content.Pipeline {

	return content.NewPipeline(
		content.DefaultRegistry(),
		&content.Canonicalizer{},
		content.NewDeduper(),
		content.NewChunker(cfg.Chunk.MaxTokens, cfg.Chunk.Overlap),
	)
}

// newBrowserManager 按配置构造浏览器管理器（禁用或不可用时返回错误）。
func newBrowserManager(cfg *config.Config) (*browser.Manager, error) {

	if !cfg.Browser.Enabled {
		return nil, fmt.Errorf("browser disabled by config")
	}
	return browser.NewManager(browser.Options{
		ExecutablePath: cfg.Browser.ExecutablePath,
		Timeout:        cfg.Browser.Timeout.Duration,
	})
}
