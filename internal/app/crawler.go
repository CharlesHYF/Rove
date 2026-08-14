/*
 * 文件作用：Application Service -- Crawl 抓取服务，seed -> frontier -> scheduler -> fetch -> pipeline -> index（验收 A1），供 CLI 与 MCP 复用。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package app

import (
	"context"

	"rove/internal/config"
	"rove/internal/scheduler"
	"rove/pkg/content"
	"rove/pkg/crawler"
	"rove/pkg/elastic"
	"rove/pkg/fetch"
	"rove/pkg/index"
	"rove/pkg/retrieval"
)

// CrawlerService 站点抓取服务。
type CrawlerService struct {
	cfg *config.Config
}

// NewCrawler 按配置构造抓取服务。
func NewCrawler(cfg *config.Config) *CrawlerService {

	return &CrawlerService{cfg: cfg}
}

// Crawl 从 seed 执行一轮抓取并返回统计；maxPages/maxDepth/workers 为 0 时取配置默认；stateDB 非空时覆盖 frontier 路径。
func (s *CrawlerService) Crawl(ctx context.Context, seed string, maxPages, maxDepth, workers int, stateDB string) (*crawler.Stats, error) {

	client, err := elastic.New(s.cfg.ES.URL, s.cfg.ES.Username, s.cfg.ES.Password, s.cfg.Index.Prefix)
	if err != nil {
		return nil, err
	}
	if err := client.EnsureIndexes(ctx, s.cfg.Index.EmbeddingDim); err != nil {
		return nil, err
	}

	dbPath := stateDB
	if dbPath == "" {
		dbPath = s.cfg.Crawl.StateDB
	}
	frontier, err := scheduler.NewFrontier(dbPath)
	if err != nil {
		return nil, err
	}
	defer frontier.Close()

	sched := scheduler.NewScheduler(frontier, scheduler.Options{
		HostConcurrency: 1,
		MaxRetries:      5,
	})
	fetcher := fetch.NewHTTP(fetch.Options{AllowPrivate: s.cfg.Fetch.AllowPrivate})
	registry := content.DefaultRegistry()
	pipeline := content.NewPipeline(registry, &content.Canonicalizer{}, content.NewDeduper(), content.NewChunker(s.cfg.Chunk.MaxTokens, s.cfg.Chunk.Overlap))
	indexer := index.New(client, retrieval.NewPseudoEmbedder(s.cfg.Index.EmbeddingDim))

	if workers <= 0 {
		workers = s.cfg.Crawl.Workers
	}
	if maxPages <= 0 {
		maxPages = s.cfg.Crawl.MaxPages
	}
	if maxDepth <= 0 {
		maxDepth = s.cfg.Crawl.MaxDepth
	}

	c := crawler.New(frontier, sched, crawler.NewRobotsClient(fetcher), crawler.Policy{
		MaxDepth: maxDepth, DomainAllowlist: s.cfg.Crawl.DomainAllowlist, SameHostOnly: true,
	}, fetcher, registry, pipeline, indexer, workers, maxPages)

	return c.Run(ctx, []string{seed})
}
