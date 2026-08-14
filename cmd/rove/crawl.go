/*
 * 文件作用：rove crawl 子命令 -- 调用 CrawlerService 执行 seed 抓取与索引（验收 A1）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-15
 */
package main

import (
	"github.com/spf13/cobra"

	"rove/internal/app"
	"rove/internal/config"
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

// runCrawl 调用 CrawlerService 执行一轮抓取。
func runCrawl(cmd *cobra.Command, args []string) error {

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	stats, err := app.NewCrawler(cfg).Crawl(cmd.Context(), args[0], crawlFlags.maxPages, crawlFlags.maxDepth, crawlFlags.workers, crawlFlags.db)
	if err != nil {
		return err
	}
	cmd.Printf("crawl done: visited=%d indexed=%d failed=%d duplicate=%d\n",
		stats.Visited, stats.Indexed, stats.Failed, stats.Duplicate)
	return nil
}
