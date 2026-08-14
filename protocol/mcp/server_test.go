/*
 * 文件作用：MCP Server 测试 -- 用官方 SDK 内存传输走真实协议流，验证工具注册、调用与错误转义。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"rove/pkg/browser"
	"rove/pkg/content"
	"rove/pkg/crawler"
	"rove/pkg/document"
	"rove/pkg/evidence"
	"rove/pkg/fetch"
	"rove/pkg/retrieval"
)

// fakeSearch 检索服务测试替身。
type fakeSearch struct {
	result *retrieval.SearchResult
	items  []evidence.Evidence
	err    error
}

func (f *fakeSearch) Search(ctx context.Context, query *retrieval.Query) (*retrieval.SearchResult, error) {

	return f.result, f.err
}

func (f *fakeSearch) Evidence(ctx context.Context, result *retrieval.SearchResult) ([]evidence.Evidence, error) {

	return f.items, f.err
}

// fakeFetch 抓取服务测试替身。
type fakeFetch struct {
	doc *document.Document
	err error
}

func (f *fakeFetch) Fetch(ctx context.Context, url string, mode fetch.FetchMode) (*document.Document, *content.DedupResult, error) {

	if f.err != nil {
		return nil, nil, f.err
	}
	return f.doc, &content.DedupResult{}, nil
}

// fakeBrowse 浏览服务测试替身。
type fakeBrowse struct {
	state *browser.PageState
	err   error
}

func (f *fakeBrowse) Browse(ctx context.Context, url string) (*browser.PageState, error) {

	return f.state, f.err
}

// fakeCrawl 抓取服务测试替身。
type fakeCrawl struct {
	stats *crawler.Stats
	err   error
}

func (f *fakeCrawl) Crawl(ctx context.Context, seed string, maxPages, maxDepth, workers int, stateDB string) (*crawler.Stats, error) {

	return f.stats, f.err
}

// connectTestServer 用内存传输连接测试服务端并返回客户端会话。
func connectTestServer(t *testing.T, deps Deps) *sdkmcp.ClientSession {

	t.Helper()
	server := NewServer("rove", "0.1.0", deps)
	serverTransport, clientTransport := sdkmcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := server.sdk.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("连接服务端失败: %v", err)
	}
	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = session.Close()
	})
	return session
}

// testDeps 构造四个非空替身。
func testDeps() Deps {

	return Deps{
		Search: &fakeSearch{},
		Fetch:  &fakeFetch{},
		Browse: &fakeBrowse{},
		Crawl:  &fakeCrawl{},
	}
}

func TestServerListsFourTools(t *testing.T) {

	session := connectTestServer(t, testDeps())

	toolNames := map[string]bool{}
	for tool, err := range session.Tools(context.Background(), nil) {
		require.NoError(t, err)
		toolNames[tool.Name] = true
	}
	require.Len(t, toolNames, 4)
	require.True(t, toolNames["rove_search"])
	require.True(t, toolNames["rove_fetch"])
	require.True(t, toolNames["rove_browse"])
	require.True(t, toolNames["rove_crawl"])
}

func TestCallSearchTool(t *testing.T) {

	result := &retrieval.SearchResult{
		Query:   "browser agents",
		TraceID: "trace-1",
		Hits: []retrieval.SearchHit{
			{
				Document: &document.Document{ID: "d1", URL: "https://example.com/a", Title: "Browser Doc"},
				Chunk:    document.Chunk{ChunkID: "d1-0", DocumentID: "d1", Content: "browser agent"},
				Score:    0.9,
			},
		},
		Timings: map[string]time.Duration{"lexical_retrieve": 18 * time.Millisecond},
	}
	items := []evidence.Evidence{
		{DocumentID: "d1", ChunkID: "d1-0", URL: "https://example.com/a", Title: "Browser Doc", Text: "browser agent", Score: 0.9},
	}
	deps := testDeps()
	deps.Search = &fakeSearch{result: result, items: items}
	session := connectTestServer(t, deps)

	callResult, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "rove_search",
		Arguments: map[string]any{"query": "browser agents", "top_k": 5},
	})
	require.NoError(t, err)
	require.False(t, callResult.IsError)
	require.NotEmpty(t, callResult.Content)
	text := callResult.Content[0].(*sdkmcp.TextContent).Text

	var response struct {
		Query string `json:"query"`
	}
	require.NoError(t, json.Unmarshal([]byte(text), &response))
	require.Equal(t, "browser agents", response.Query)
	require.Contains(t, text, "Browser Doc")
	require.Contains(t, text, "trace-1")
}

func TestCallSearchToolEmptyQueryIsError(t *testing.T) {

	session := connectTestServer(t, testDeps())

	callResult, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "rove_search",
		Arguments: map[string]any{"query": "   "},
	})
	require.NoError(t, err)
	require.True(t, callResult.IsError, "空 query 应作为工具错误返回")
}

func TestCallSearchToolBackendErrorIsError(t *testing.T) {

	deps := testDeps()
	deps.Search = &fakeSearch{err: context.DeadlineExceeded}
	session := connectTestServer(t, deps)

	callResult, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "rove_search",
		Arguments: map[string]any{"query": "browser"},
	})
	require.NoError(t, err)
	require.True(t, callResult.IsError, "后端错误应作为工具错误返回")
}

func TestCallFetchTool(t *testing.T) {

	deps := testDeps()
	deps.Fetch = &fakeFetch{doc: &document.Document{Title: "Example Page", URL: "https://example.com", Content: "hello world"}}
	session := connectTestServer(t, deps)

	callResult, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "rove_fetch",
		Arguments: map[string]any{"url": "https://example.com"},
	})
	require.NoError(t, err)
	require.False(t, callResult.IsError)
	text := callResult.Content[0].(*sdkmcp.TextContent).Text
	require.Contains(t, text, "Example Page")
	require.Contains(t, text, "hello world")
}

func TestCallFetchToolInvalidModeIsError(t *testing.T) {

	session := connectTestServer(t, testDeps())

	callResult, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "rove_fetch",
		Arguments: map[string]any{"url": "https://example.com", "mode": "ftp"},
	})
	require.NoError(t, err)
	require.True(t, callResult.IsError, "非法模式应作为工具错误返回")
}

func TestCallBrowseTool(t *testing.T) {

	deps := testDeps()
	deps.Browse = &fakeBrowse{state: &browser.PageState{URL: "https://example.com", Title: "SPA 页面", Text: "rendered"}}
	session := connectTestServer(t, deps)

	callResult, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "rove_browse",
		Arguments: map[string]any{"url": "https://example.com"},
	})
	require.NoError(t, err)
	require.False(t, callResult.IsError)
	text := callResult.Content[0].(*sdkmcp.TextContent).Text
	require.Contains(t, text, "SPA 页面")
	require.Contains(t, text, "rendered")
}

func TestCallCrawlTool(t *testing.T) {

	deps := testDeps()
	deps.Crawl = &fakeCrawl{stats: &crawler.Stats{Visited: 3, Indexed: 2, Failed: 1, Duplicate: 1}}
	session := connectTestServer(t, deps)

	callResult, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "rove_crawl",
		Arguments: map[string]any{"seed": "https://example.com", "max_pages": 5, "max_depth": 1},
	})
	require.NoError(t, err)
	require.False(t, callResult.IsError)
	text := callResult.Content[0].(*sdkmcp.TextContent).Text
	require.Contains(t, text, `"Visited":3`)
	require.Contains(t, text, `"Indexed":2`)
}

func TestNewServerRejectsNilDeps(t *testing.T) {

	require.Panics(t, func() {
		NewServer("rove", "0.1.0", Deps{})
	})
}
