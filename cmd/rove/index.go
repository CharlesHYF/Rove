/*
 * 文件作用：rove index 子命令 -- 初始化 / 状态 / 统计 / 重建索引（规格书 §5.3 alias 管理）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"rove/internal/app"
	"rove/internal/config"
	"rove/pkg/elastic"
	"rove/pkg/index"
)

// indexCmd 实现 rove index <init|status|stats|rebuild>。
var indexCmd = &cobra.Command{
	Use:   "index <init|status|stats|rebuild>",
	Short: "索引生命周期管理",
	Args:  cobra.ExactArgs(1),
	RunE:  runIndex,
}

// runIndex 分发索引子操作。
func runIndex(cmd *cobra.Command, args []string) error {

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	client, err := elastic.New(cfg.ES.URL, cfg.ES.Username, cfg.ES.Password, cfg.Index.Prefix)
	if err != nil {
		return err
	}
	service := app.NewIndex(index.New(client))
	ctx := cmd.Context()

	switch args[0] {
	case "init":
		if err := client.EnsureIndexes(ctx, cfg.Index.EmbeddingDim); err != nil {
			return err
		}
		cmd.Printf("indexes ready: %s, %s\n", client.DocumentsAlias(), client.ChunksAlias())
		return nil
	case "status", "stats":
		status, err := service.Status(ctx)
		if err != nil {
			return err
		}
		cmd.Printf("health: %s\n", status.Health)
		cmd.Printf("documents: %d\n", status.Documents)
		cmd.Printf("chunks: %d\n", status.Chunks)
		cmd.Printf("size_bytes: %d\n", status.SizeBytes)
		return nil
	case "rebuild":
		if err := client.DeleteIndex(ctx, client.DocumentsAlias()); err != nil {
			return err
		}
		if err := client.DeleteIndex(ctx, client.ChunksAlias()); err != nil {
			return err
		}
		if err := client.EnsureIndexes(ctx, cfg.Index.EmbeddingDim); err != nil {
			return err
		}
		cmd.Printf("indexes rebuilt\n")
		return nil
	default:
		return fmt.Errorf("unknown index subcommand: %s (init|status|stats|rebuild)", args[0])
	}
}
