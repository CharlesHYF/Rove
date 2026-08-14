/*
 * 文件作用：Application Service -- Browser 浏览服务，输出页面状态（PRD FR-BRW-002），供 CLI 与 MCP 复用。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package app

import (
	"context"

	"rove/internal/config"
	"rove/pkg/browser"
)

// BrowserService 浏览器渲染服务。
type BrowserService struct {
	cfg *config.Config
}

// NewBrowser 按配置构造浏览服务。
func NewBrowser(cfg *config.Config) *BrowserService {

	return &BrowserService{cfg: cfg}
}

// Browse 渲染页面并返回 Page State。
func (s *BrowserService) Browse(ctx context.Context, url string) (*browser.PageState, error) {

	manager, err := newBrowserManager(s.cfg)
	if err != nil {
		return nil, err
	}
	defer manager.Close()
	return manager.Browse(ctx, url)
}
