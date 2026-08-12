/*
 * 文件作用：Crawler -- 有界并发工作池：seed 入队 -> scheduler 取号 -> fetch -> pipeline -> index -> 链接回填（规格书 §4 / A1）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package crawler

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"rove/internal/scheduler"
	"rove/pkg/content"
	"rove/pkg/document"
	"rove/pkg/fetch"
)

// Indexer 是 Crawler 依赖的最小索引接口（index.Indexer 满足）。
type Indexer interface {
	Index(ctx context.Context, doc *document.Document) error
}

// errBlockedByRobots 表示页面被 robots.txt 拒绝（计入 duplicate，不视为失败）。
var errBlockedByRobots = errors.New("blocked by robots.txt")

// Stats 一轮抓取统计。
type Stats struct {
	Visited   int
	Indexed   int
	Failed    int
	Duplicate int
}

// Crawler 是抓取编排器。
type Crawler struct {
	frontier *scheduler.Frontier
	sched    *scheduler.Scheduler
	robots   *RobotsClient
	policy   Policy
	fetcher  *fetch.HTTPFetcher
	registry *content.Registry
	pipeline *content.Pipeline
	indexer  Indexer
	workers  int
	maxPages int
}

// New 构造 Crawler。
func New(frontier *scheduler.Frontier, sched *scheduler.Scheduler, robots *RobotsClient, policy Policy,
	fetcher *fetch.HTTPFetcher, registry *content.Registry, pipeline *content.Pipeline,
	indexer Indexer, workers, maxPages int) *Crawler {

	return &Crawler{
		frontier: frontier, sched: sched, robots: robots, policy: policy,
		fetcher: fetcher, registry: registry, pipeline: pipeline,
		indexer: indexer, workers: workers, maxPages: maxPages,
	}
}

// Run 从 seeds 入队开始执行一轮有界并发抓取，直到无可用条目或达到 maxPages。
func (c *Crawler) Run(ctx context.Context, seeds []string) (*Stats, error) {

	now := time.Now().UTC()
	seedEntries := make([]scheduler.FrontierEntry, 0, len(seeds))
	for _, seed := range seeds {
		normalized, err := NormalizeURL(seed)
		if err != nil {
			continue
		}
		seedEntries = append(seedEntries, scheduler.FrontierEntry{
			NormalizedURL: normalized, Host: hostOf(normalized), Priority: 9, Depth: 0,
			DiscoveredAt: now, NextFetchAt: now, State: "queued",
		})
	}
	if _, err := c.frontier.Enqueue(ctx, seedEntries); err != nil {
		return nil, err
	}

	var stats Stats
	var mu sync.Mutex
	var visited atomic.Int64
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for worker := 0; worker < c.workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if visited.Load() >= int64(c.maxPages) {
					cancel()
					return
				}
				entry, err := c.sched.Next(ctx)

				if err != nil {
					cancel()
					return
				}
				if entry == nil {
					return
				}
				visited.Add(1)
				err = c.crawlOne(ctx, entry)

				mu.Lock()
				switch {
				case errors.Is(err, errBlockedByRobots):
					stats.Duplicate++
				case err != nil:
					stats.Failed++
				default:
					stats.Visited++
					stats.Indexed++
				}
				mu.Unlock()
				_ = c.sched.Complete(ctx, entry, err)
			}
		}()
	}
	wg.Wait()
	return &stats, nil
}

// crawlOne 抓取单个条目并回填发现链接。
func (c *Crawler) crawlOne(ctx context.Context, entry *scheduler.FrontierEntry) error {

	allowed, err := c.robots.IsAllowed(ctx, entry.NormalizedURL)
	if err != nil || !allowed {
		return errBlockedByRobots
	}
	delay, _ := c.robots.CrawlDelay(ctx, entry.NormalizedURL)
	if delay > 0 {
		time.Sleep(delay)
	}

	raw, err := c.fetcher.Fetch(ctx, &fetch.Request{URL: entry.NormalizedURL, Mode: fetch.ModeHTTP})
	if err != nil {
		return err
	}

	// 先解析拿 Links（Pipeline 当前不保留 Links，M4 正确性优先，二次解析可接受）
	var parsedLinks []string
	if parser, perr := c.registry.For(raw.ContentType); perr == nil {
		if parsed, aerr := parser.Parse(ctx, raw); aerr == nil {
			parsedLinks = parsed.Links
		}
	}

	doc, dedup, err := c.pipeline.Process(ctx, raw, "web")
	if err != nil {
		return err
	}
	if dedup.IsDuplicate {
		return nil
	}
	if err := c.indexer.Index(ctx, doc); err != nil {
		return err
	}

	// 链接回填（深度内 + 策略放行）
	if entry.Depth < c.policy.MaxDepth {
		links, lerr := c.policy.CandidateLinks(entry.NormalizedURL, &content.ParsedContent{Links: parsedLinks})
		if lerr == nil {
			now := time.Now().UTC()
			children := make([]scheduler.FrontierEntry, 0, len(links))
			for _, link := range links {
				normalized, nerr := NormalizeURL(link)
				if nerr != nil {
					continue
				}
				children = append(children, scheduler.FrontierEntry{
					NormalizedURL: normalized, Host: hostOf(normalized), Priority: 5, Depth: entry.Depth + 1,
					SourceURL: entry.NormalizedURL, DiscoveredAt: now, NextFetchAt: now, State: "queued",
				})
			}
			_, _ = c.frontier.Enqueue(ctx, children)
		}
	}
	return nil
}
