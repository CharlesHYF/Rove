/*
 * 文件作用：rove mcp 子命令 -- 以 stdio 启动 MCP Server，向 Agent 客户端暴露 search/fetch/browse/crawl 工具（PRD FR-MCP-001）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package main

import (
	"github.com/spf13/cobra"

	"rove/internal/app"
	"rove/internal/config"
	"rove/pkg/elastic"
	"rove/pkg/retrieval"
	"rove/protocol/mcp"
)

// mcpCmd 实现 rove mcp。
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "以 stdio 启动 MCP Server（供 Claude Desktop / Cursor 等 Agent 接入）",
	Args:  cobra.NoArgs,
	RunE:  runMCP,
}

// runMCP 装配服务并启动 MCP Server。
func runMCP(cmd *cobra.Command, args []string) error {

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	client, err := elastic.New(cfg.ES.URL, cfg.ES.Username, cfg.ES.Password, cfg.Index.Prefix)
	if err != nil {
		return err
	}
	server := mcp.NewServer("rove", version, mcp.Deps{
		Search: app.NewSearch(retrieval.New(client, retrieval.NewPseudoEmbedder(cfg.Index.EmbeddingDim))),
		Fetch:  app.NewFetch(cfg),
		Browse: app.NewBrowser(cfg),
		Crawl:  app.NewCrawler(cfg),
	})
	return server.Serve(cmd.Context())
}
