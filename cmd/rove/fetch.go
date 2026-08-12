/*
 * 文件作用：rove fetch 子命令 -- 抓取 URL 并经内容管线输出标准 Document（JSON 或人类可读摘要）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package main

import (
	"encoding/json"
	"errors"

	"github.com/spf13/cobra"

	"rove/internal/config"
	"rove/pkg/content"
	"rove/pkg/document"
	"rove/pkg/fetch"
	"rove/pkg/rove"
)

var fetchFlags struct {
	mode string
	json bool
}

// fetchCmd 实现 rove fetch <url>。
var fetchCmd = &cobra.Command{
	Use:   "fetch <url>",
	Short: "抓取 URL 并输出标准 Document",
	Args:  cobra.ExactArgs(1),
	RunE:  runFetch,
}

// init 注册 fetch 命令参数。
func init() {

	fetchCmd.Flags().StringVar(&fetchFlags.mode, "mode", "auto", "抓取模式: auto|http|browser")
	fetchCmd.Flags().BoolVar(&fetchFlags.json, "json", false, "以 JSON 输出 Document")
}

// runFetch 执行抓取与内容管线。
func runFetch(cmd *cobra.Command, args []string) error {

	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	var raw *document.RawDocument
	if fetchFlags.mode == "browser" {
		manager, browserErr := newBrowserManager(cfg)
		if browserErr != nil {
			return browserErr
		}
		defer manager.Close()
		raw, err = manager.Fetch(cmd.Context(), args[0])
	} else {
		raw, err = newFetcher(cfg).Fetch(cmd.Context(), &fetch.Request{URL: args[0], Mode: fetch.FetchMode(fetchFlags.mode)})
		if err != nil && errors.Is(err, fetch.ErrEscalationRequired) && fetchFlags.mode == "auto" {
			// HTTP 内容不足 -> 浏览器升级（验收 A2，进入同一 Document Pipeline）
			manager, browserErr := newBrowserManager(cfg)
			if browserErr != nil {
				return rove.NewError("fetch.escalation_required", rove.CategoryBrowser, false,
					"页面需要浏览器渲染：%s（%v）", args[0], browserErr)
			}
			defer manager.Close()
			raw, err = manager.Fetch(cmd.Context(), args[0])
		}
	}
	if err != nil {
		return err
	}

	doc, dedupResult, err := newPipeline(cfg).Process(cmd.Context(), raw, "web")
	if err != nil {
		return err
	}

	if fetchFlags.json {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(doc)
	}

	cmd.Printf("title: %s\n", doc.Title)
	cmd.Printf("url: %s\n", doc.URL)
	cmd.Printf("canonical: %s\n", doc.CanonicalURL)
	cmd.Printf("language: %s\n", doc.Language)
	cmd.Printf("source: %s (%s)\n", doc.Source.Domain, doc.Source.Type)
	cmd.Printf("chunks: %d\n", len(doc.Chunks))
	if dedupResult.IsDuplicate {
		cmd.Printf("warning: duplicate (%s)\n", dedupResult.Reason)
	}
	return nil
}

// newFetcher 按配置构造 HTTP Fetcher。
func newFetcher(cfg *config.Config) *fetch.HTTPFetcher {

	return fetch.NewHTTP(fetch.Options{
		MaxBodyBytes: cfg.Fetch.MaxBodyBytes,
		Timeout:      cfg.Fetch.Timeout.Duration,
		MaxRedirects: cfg.Fetch.MaxRedirects,
		UserAgent:    cfg.Fetch.UserAgent,
		AllowPrivate: cfg.Fetch.AllowPrivate,
	})
}

// newPipeline 按配置组装内容管线。
func newPipeline(cfg *config.Config) *content.Pipeline {

	return content.NewPipeline(
		content.DefaultRegistry(),
		&content.Canonicalizer{},
		content.NewDeduper(),
		content.NewChunker(cfg.Chunk.MaxTokens, cfg.Chunk.Overlap),
	)
}
