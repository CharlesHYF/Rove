/*
 * 文件作用：rove fetch 子命令 -- 调用 FetchService 抓取 URL 并输出标准 Document（JSON 或人类可读摘要）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-15
 */
package main

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"rove/internal/app"
	"rove/internal/config"
	"rove/pkg/fetch"
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

// runFetch 调用 FetchService 执行抓取与内容管线。
func runFetch(cmd *cobra.Command, args []string) error {

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	doc, dedupResult, err := app.NewFetch(cfg).Fetch(cmd.Context(), args[0], fetch.FetchMode(fetchFlags.mode))
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
