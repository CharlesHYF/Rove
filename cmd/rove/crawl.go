/*
 * 文件作用：rove crawl 子命令 -- seed -> frontier -> scheduler -> fetch -> pipeline -> index（验收 A1）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package main

import (
	"github.com/spf13/cobra"

	"rove/internal/config"
	"rove/internal/scheduler"
	"rove/pkg/content"
	"rove/pkg/crawler"
	"rove/pkg/elastic"
	"rove/pkg/fetch"
	"rove/pkg/index"
	"rove/pkg/retrieval"
)

var crawlFlags struct {
	maxPages int
	maxDepth int
	workers  int
	db       string
}

// crawlCmd 实现 rove crawl <seed>。
var crawlCmd = &cobra.Command{
	Use:   "crawl <seed>",
	Short: "从 seed 开始抓取并索引",
	Args:  cobra.ExactArgs(1),
	RunE:  runCrawl,
}

// init 注册 crawl 命令参数。
func init() {

	crawlCmd.Flags().IntVar(&crawlFlags.maxPages, "max-pages", 0, "抓取页面上限（默认配置 crawl.max_pages）")
	crawlCmd.Flags().IntVar(&crawlFlags.maxDepth, "max-depth", 0, "抓取深度上限（默认配置 crawl.max_depth）")
	crawlCmd.Flags().IntVar(&crawlFlags.workers, "workers", 0, "并发 worker 数（默认配置 crawl.workers）")
	crawlCmd.Flags().StringVar(&crawlFlags.db, "db", "", "frontier SQLite 路径（默认配置 crawl.state_db）")
}

// runCrawl 执行一轮抓取。
func runCrawl(cmd *cobra.Command, args []string) error {

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	client, err := elastic.New(cfg.ES.URL, cfg.ES.Username, cfg.ES.Password, cfg.Index.Prefix)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := client.EnsureIndexes(ctx, cfg.Index.EmbeddingDim); err != nil {
		return err
	}

	dbPath := crawlFlags.db
	if dbPath == "" {
		dbPath = cfg.Crawl.StateDB
	}
	frontier, err := scheduler.NewFrontier(dbPath)
	if err != nil {
		return err
	}
	defer frontier.Close()

	sched := scheduler.NewScheduler(frontier, scheduler.Options{
		HostConcurrency: 1,
		MaxRetries:      5,
	})
	fetcher := fetch.NewHTTP(fetch.Options{AllowPrivate: cfg.Fetch.AllowPrivate})
	registry := content.DefaultRegistry()
	pipeline := content.NewPipeline(registry, &content.Canonicalizer{}, content.NewDeduper(), content.NewChunker(cfg.Chunk.MaxTokens, cfg.Chunk.Overlap))
	indexer := index.New(client, retrieval.NewPseudoEmbedder())

	workers := crawlFlags.workers
	if workers <= 0 {
		workers = cfg.Crawl.Workers
	}
	maxPages := crawlFlags.maxPages
	if maxPages <= 0 {
		maxPages = cfg.Crawl.MaxPages
	}
	maxDepth := crawlFlags.maxDepth
	if maxDepth <= 0 {
		maxDepth = cfg.Crawl.MaxDepth
	}

	c := crawler.New(frontier, sched, crawler.NewRobotsClient(fetcher), crawler.Policy{
		MaxDepth: maxDepth, DomainAllowlist: cfg.Crawl.DomainAllowlist, SameHostOnly: true,
	}, fetcher, registry, pipeline, indexer, workers, maxPages)

	stats, err := c.Run(ctx, []string{args[0]})
	if err != nil {
		return err
	}
	cmd.Printf("crawl done: visited=%d indexed=%d failed=%d duplicate=%d\n",
		stats.Visited, stats.Indexed, stats.Failed, stats.Duplicate)
	return nil
}
