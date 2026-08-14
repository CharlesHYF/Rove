/*
 * 文件作用：MCP 工具注册 -- rove_search/rove_fetch/rove_browse/rove_crawl 的入参 schema 与处理器。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package mcp

import (
	"context"
	"fmt"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"rove/pkg/fetch"
	"rove/pkg/retrieval"
	"rove/protocol"
)

// 工具参数常量。
const (
	defaultTopK      = 10        // rove_search 默认返回条数
	defaultFetchMode = "auto"    // rove_fetch 默认模式
	modeAuto         = "auto"    // 抓取模式:HTTP 优先,内容不足自动浏览器升级
	modeHTTP         = "http"    // 抓取模式:仅 HTTP
	modeBrowser      = "browser" // 抓取模式:直接浏览器渲染
)

// searchArgs rove_search 入参（omitempty 字段为可选）。
type searchArgs struct {
	Query    string `json:"query" jsonschema:"检索问题，必填"`
	TopK     int    `json:"top_k,omitempty" jsonschema:"返回条数，默认 10"`
	Domain   string `json:"domain,omitempty" jsonschema:"按域名过滤"`
	Language string `json:"language,omitempty" jsonschema:"按语言过滤，如 en/zh"`
	Vertical string `json:"vertical,omitempty" jsonschema:"垂类路由 auto/web/docs/code/academic，默认 auto"`
}

// registerSearchTool 注册 rove_search。
func (s *Server) registerSearchTool(search SearchService) {

	sdkmcp.AddTool(s.sdk, &sdkmcp.Tool{
		Name:        "rove_search",
		Description: "从 Rove 自有索引检索，返回 Evidence（URL/标题/文本/分数/来源/耗时）",
	}, func(ctx context.Context, request *sdkmcp.CallToolRequest, args searchArgs) (*sdkmcp.CallToolResult, any, error) {
		if strings.TrimSpace(args.Query) == "" {
			return nil, nil, fmt.Errorf("query 不能为空")
		}
		vertical, verticalErr := retrieval.ParseVertical(args.Vertical)
		if verticalErr != nil {
			return nil, nil, verticalErr
		}
		topK := args.TopK
		if topK <= 0 {
			topK = defaultTopK
		}
		result, err := search.Search(ctx, &retrieval.Query{
			Text:     args.Query,
			TopK:     topK,
			Filters:  retrieval.Filters{Domain: args.Domain, Language: args.Language},
			Vertical: vertical,
		})
		if err != nil {
			return nil, nil, err
		}
		items, err := search.Evidence(ctx, result)
		if err != nil {
			return nil, nil, err
		}
		return nil, protocol.FromSearchResult(result, items), nil
	})
}

// fetchArgs rove_fetch 入参（omitempty 字段为可选）。
type fetchArgs struct {
	URL  string `json:"url" jsonschema:"要抓取的页面地址，必填"`
	Mode string `json:"mode,omitempty" jsonschema:"抓取模式 auto/http/browser，默认 auto"`
}

// registerFetchTool 注册 rove_fetch。
func (s *Server) registerFetchTool(fetcher FetchService) {

	sdkmcp.AddTool(s.sdk, &sdkmcp.Tool{
		Name:        "rove_fetch",
		Description: "抓取单个页面并输出标准 Document（auto 模式内容不足时自动浏览器升级）",
	}, func(ctx context.Context, request *sdkmcp.CallToolRequest, args fetchArgs) (*sdkmcp.CallToolResult, any, error) {
		if strings.TrimSpace(args.URL) == "" {
			return nil, nil, fmt.Errorf("url 不能为空")
		}
		mode := args.Mode
		if mode == "" {
			mode = defaultFetchMode
		}
		if mode != modeAuto && mode != modeHTTP && mode != modeBrowser {
			return nil, nil, fmt.Errorf("mode 仅支持 auto/http/browser，收到 %q", mode)
		}
		doc, _, err := fetcher.Fetch(ctx, args.URL, fetch.FetchMode(mode))
		if err != nil {
			return nil, nil, err
		}
		return nil, doc, nil
	})
}

// browseArgs rove_browse 入参。
type browseArgs struct {
	URL string `json:"url" jsonschema:"要渲染的页面地址，必填"`
}

// registerBrowseTool 注册 rove_browse。
func (s *Server) registerBrowseTool(browserService BrowseService) {

	sdkmcp.AddTool(s.sdk, &sdkmcp.Tool{
		Name:        "rove_browse",
		Description: "浏览器渲染页面并输出 Page State（文本/链接/交互元素与稳定 id）",
	}, func(ctx context.Context, request *sdkmcp.CallToolRequest, args browseArgs) (*sdkmcp.CallToolResult, any, error) {
		if strings.TrimSpace(args.URL) == "" {
			return nil, nil, fmt.Errorf("url 不能为空")
		}
		state, err := browserService.Browse(ctx, args.URL)
		if err != nil {
			return nil, nil, err
		}
		return nil, state, nil
	})
}

// crawlArgs rove_crawl 入参（omitempty 字段为可选）。
type crawlArgs struct {
	Seed     string `json:"seed" jsonschema:"抓取起始 URL，必填"`
	MaxPages int    `json:"max_pages,omitempty" jsonschema:"抓取页面上限，0 取配置默认"`
	MaxDepth int    `json:"max_depth,omitempty" jsonschema:"抓取深度上限，0 取配置默认"`
}

// registerCrawlTool 注册 rove_crawl。
func (s *Server) registerCrawlTool(crawlerService CrawlService) {

	sdkmcp.AddTool(s.sdk, &sdkmcp.Tool{
		Name:        "rove_crawl",
		Description: "从 seed 抓取站点并建立索引，输出抓取统计（visited/indexed/failed/duplicate）",
	}, func(ctx context.Context, request *sdkmcp.CallToolRequest, args crawlArgs) (*sdkmcp.CallToolResult, any, error) {
		if strings.TrimSpace(args.Seed) == "" {
			return nil, nil, fmt.Errorf("seed 不能为空")
		}
		stats, err := crawlerService.Crawl(ctx, args.Seed, args.MaxPages, args.MaxDepth, 0, "")
		if err != nil {
			return nil, nil, err
		}
		return nil, stats, nil
	})
}
