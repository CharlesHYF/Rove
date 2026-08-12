/*
 * 文件作用：internal/app 服务的组装冒烟测试 -- 索引 -> 检索全链路（ES 不可达时跳过）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package app

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/internal/runtime"
	"rove/pkg/document"
	"rove/pkg/elastic"
	"rove/pkg/index"
	"rove/pkg/retrieval"
)

func TestServicesSmoke(t *testing.T) {

	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	client, err := elastic.New(url, "", "", "rove-test-app")
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))

	ctx := context.Background()
	indexService := NewIndex(index.New(client))
	searchService := NewSearch(retrieval.New(client))

	now := time.Now().UTC()
	doc := &document.Document{
		ID: "app-1", URL: "https://x.com/1", CanonicalURL: "https://x.com/1", Title: "App Smoke",
		Content: "app service smoke test", Markdown: "# App Smoke", Language: "en", FetchedAt: now,
		Source: document.Source{Domain: "x.com", Type: "web"}, ContentHash: "h",
		Chunks: []document.Chunk{{ChunkID: "app-1-0", DocumentID: "app-1", Position: 0, Content: "app service smoke test", TokenCount: 4, Language: "en"}},
	}
	require.NoError(t, indexService.Index(ctx, doc))

	result, err := searchService.Search(ctx, &retrieval.Query{Text: "smoke", TopK: 5})
	require.NoError(t, err)
	require.Len(t, result.Hits, 1)
	require.Equal(t, "App Smoke", result.Hits[0].Document.Title)
}

func TestServicesPublishEvents(t *testing.T) {

	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	client, err := elastic.New(url, "", "", "rove-test-app-events")
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))

	ctx := context.Background()
	bus := runtime.NewBus(20)
	defer bus.Close()

	indexService := NewIndex(index.New(client), bus)
	searchService := NewSearch(retrieval.New(client), bus)

	now := time.Now().UTC()
	doc := &document.Document{
		ID: "ev-1", URL: "https://x.com/ev", CanonicalURL: "https://x.com/ev", Title: "Event Doc",
		Content: "event smoke test", Markdown: "# Event", Language: "en", FetchedAt: now,
		Source: document.Source{Domain: "x.com", Type: "web"}, ContentHash: "he",
		Chunks: []document.Chunk{{ChunkID: "ev-1-0", DocumentID: "ev-1", Position: 0, Content: "event smoke test", TokenCount: 3, Language: "en"}},
	}
	require.NoError(t, indexService.Index(ctx, doc))
	_, err = searchService.Search(ctx, &retrieval.Query{Text: "event", TopK: 5})
	require.NoError(t, err)

	events := bus.Recent(10)
	require.Len(t, events, 2)
	require.Equal(t, runtime.Index, events[0].Type)
	require.Equal(t, runtime.Search, events[1].Type)
	require.Contains(t, events[1].Fields, "hits")
}

func TestInspector(t *testing.T) {

	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	client, err := elastic.New(url, "", "", "rove-test-inspector")
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))

	bus := runtime.NewBus(10)
	defer bus.Close()
	inspector := NewInspector(bus, NewIndex(index.New(client), bus))

	status, err := inspector.Status(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, status.Health)

	bus.Publish(runtime.Event{Type: runtime.Crawl, Message: "crawl round done"})
	require.Len(t, inspector.Events(10), 1)
	require.NotNil(t, inspector.Subscribe())
}
