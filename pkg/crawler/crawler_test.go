/*
 * 文件作用：pkg/crawler 编排的集成测试 -- 站点图抓取（httptest 3 页 + 链接回填）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-13
 */
package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/internal/scheduler"
	"rove/pkg/content"
	"rove/pkg/document"
	"rove/pkg/fetch"
)

// fakeIndexer 是 crawler.Indexer 的测试替身。
type fakeIndexer struct {
	count atomic.Int64
}

// Index 记录一次索引调用。
func (f *fakeIndexer) Index(ctx context.Context, doc *document.Document) error {

	f.count.Add(1)
	return nil
}

// timeNow 返回当前 UTC 时间。
func timeNow() time.Time {

	return time.Now().UTC()
}

func TestCrawlerVisitsSite(t *testing.T) {

	// 站点图：/ -> /a -> /b（同 host）
	site := map[string]string{
		"/":  `<html><head><title>Home</title></head><body><p>home page content here</p><a href="/a">a</a></body></html>`,
		"/a": `<html><head><title>Page A</title></head><body><p>page a content here</p><a href="/b">b</a></body></html>`,
		"/b": `<html><head><title>Page B</title></head><body><p>page b content here</p></body></html>`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, ok := site[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(body))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	frontier, _ := scheduler.NewFrontier(":memory:")
	defer frontier.Close()
	ctx := context.Background()
	_, _ = frontier.Enqueue(ctx, []scheduler.FrontierEntry{{
		NormalizedURL: ts.URL + "/", Host: hostOf(ts.URL), Priority: 9, Depth: 0,
		DiscoveredAt: timeNow(), NextFetchAt: timeNow(), State: "queued",
	}})

	sched := scheduler.NewScheduler(frontier, scheduler.Options{HostConcurrency: 2})
	fetcher := fetch.NewHTTP(fetch.Options{AllowPrivate: true})
	pipeline := content.NewPipeline(content.DefaultRegistry(), &content.Canonicalizer{}, content.NewDeduper(), content.NewChunker(800, 100))
	indexer := &fakeIndexer{}
	crawler := New(frontier, sched, NewRobotsClient(fetcher), Policy{SameHostOnly: true, MaxDepth: 3},
		fetcher, content.DefaultRegistry(), pipeline, indexer, 2, 10)

	stats, err := crawler.Run(ctx, []string{ts.URL + "/"})
	require.NoError(t, err)
	require.GreaterOrEqual(t, stats.Visited, 3, "should visit home, a, b")
	require.GreaterOrEqual(t, indexer.count.Load(), int64(3), "all pages indexed")
}

// TestCrawlerReCrawlSeedAfterDone 回归：首轮抓取完成后同 seed 重跑，须重新抓取而非静默忽略。
func TestCrawlerReCrawlSeedAfterDone(t *testing.T) {

	site := map[string]string{
		"/": `<html><head><title>Home</title></head><body><p>home page content here</p></body></html>`,
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, ok := site[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(body))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	ctx := context.Background()
	frontier, _ := scheduler.NewFrontier(":memory:")
	defer frontier.Close()
	sched := scheduler.NewScheduler(frontier, scheduler.Options{HostConcurrency: 2})
	fetcher := fetch.NewHTTP(fetch.Options{AllowPrivate: true})
	pipeline := content.NewPipeline(content.DefaultRegistry(), &content.Canonicalizer{}, content.NewDeduper(), content.NewChunker(800, 100))
	newCrawler := func() *Crawler {

		return New(frontier, sched, NewRobotsClient(fetcher), Policy{SameHostOnly: true, MaxDepth: 3},
			fetcher, content.DefaultRegistry(), pipeline, &fakeIndexer{}, 2, 10)
	}

	// 第一轮：seed 抓取完成，状态置 done
	firstStats, err := newCrawler().Run(ctx, []string{ts.URL + "/"})
	require.NoError(t, err)
	require.Equal(t, 1, firstStats.Visited)

	// 第二轮：同 seed 重跑，必须重新抓取（旧实现 Enqueue 直接忽略，visited=0）
	secondStats, err := newCrawler().Run(ctx, []string{ts.URL + "/"})
	require.NoError(t, err)
	require.Equal(t, 1, secondStats.Visited, "done 后重抓同 seed 必须生效")
	require.Equal(t, 1, secondStats.Indexed)
}
