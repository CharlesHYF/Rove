/*
 * 文件作用：Rove MCP Server -- 基于官方 go-sdk 组装 stdio 服务并注册四个 Agent 工具（PRD FR-MCP-001）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
// Package mcp 提供 Rove 的 MCP Server（stdio 传输），供 Claude Desktop / Cursor 等 Agent 客户端接入。
package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"rove/pkg/browser"
	"rove/pkg/content"
	"rove/pkg/crawler"
	"rove/pkg/document"
	"rove/pkg/evidence"
	"rove/pkg/fetch"
	"rove/pkg/retrieval"
)

// SearchService 检索服务接口（消费方定义，由 internal/app 实现）。
type SearchService interface {
	Search(ctx context.Context, query *retrieval.Query) (*retrieval.SearchResult, error)
	Evidence(ctx context.Context, result *retrieval.SearchResult) ([]evidence.Evidence, error)
}

// FetchService 抓取服务接口（消费方定义，由 internal/app 实现）。
type FetchService interface {
	Fetch(ctx context.Context, url string, mode fetch.FetchMode) (*document.Document, *content.DedupResult, error)
}

// BrowseService 浏览服务接口（消费方定义，由 internal/app 实现）。
type BrowseService interface {
	Browse(ctx context.Context, url string) (*browser.PageState, error)
}

// CrawlService 站点抓取服务接口（消费方定义，由 internal/app 实现）。
type CrawlService interface {
	Crawl(ctx context.Context, seed string, maxPages, maxDepth, workers int, stateDB string) (*crawler.Stats, error)
}

// Deps 是四个工具的依赖集合，由 cmd/rove/mcp.go 注入 internal/app 服务。
type Deps struct {
	Search SearchService
	Fetch  FetchService
	Browse BrowseService
	Crawl  CrawlService
}

// Server 包装官方 SDK Server。
type Server struct {
	sdk *sdkmcp.Server
}

// NewServer 构造 MCP Server 并注册四个工具；Deps 四接口均不能为 nil。
func NewServer(name, version string, deps Deps) *Server {

	if deps.Search == nil || deps.Fetch == nil || deps.Browse == nil || deps.Crawl == nil {
		panic("mcp.NewServer: Deps 四接口均不能为 nil")
	}
	server := &Server{sdk: sdkmcp.NewServer(&sdkmcp.Implementation{Name: name, Version: version}, nil)}
	server.registerSearchTool(deps.Search)
	server.registerFetchTool(deps.Fetch)
	server.registerBrowseTool(deps.Browse)
	server.registerCrawlTool(deps.Crawl)
	return server
}

// Serve 以 stdio 传输运行服务，阻塞至 stdin 关闭。
func (s *Server) Serve(ctx context.Context) error {

	return s.sdk.Run(ctx, &sdkmcp.StdioTransport{})
}
