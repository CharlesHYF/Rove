/*
 * 文件作用：rove map 子命令 -- 调用 MapService 发现页面链接（仅 URL Discovery，不建索引）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package main

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"rove/internal/app"
	"rove/internal/config"
)

var mapFlags struct {
	json bool
}

// mapCmd 实现 rove map <url>。
var mapCmd = &cobra.Command{
	Use:   "map <url>",
	Short: "发现页面链接（站内/站外，不建索引）",
	Args:  cobra.ExactArgs(1),
	RunE:  runMap,
}

// init 注册 map 命令参数。
func init() {

	mapCmd.Flags().BoolVar(&mapFlags.json, "json", false, "以 JSON 输出链接发现结果")
}

// runMap 调用 MapService 发现链接。
func runMap(cmd *cobra.Command, args []string) error {

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	result, err := app.NewMap(cfg).Map(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	if mapFlags.json {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	cmd.Printf("url: %s\n", result.URL)
	cmd.Printf("canonical: %s\n", result.Canonical)
	cmd.Printf("in-site links: %d\n", len(result.InSite))
	for _, link := range result.InSite {
		cmd.Printf("  %s\n", link)
	}
	cmd.Printf("out-site links: %d\n", len(result.OutSite))
	for _, link := range result.OutSite {
		cmd.Printf("  %s\n", link)
	}
	return nil
}
