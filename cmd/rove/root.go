/*
 * 文件作用：Rove CLI 根命令 -- 定义命令树入口与版本信息。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package main

import (
	"github.com/spf13/cobra"
)

// version 在发布时通过 ldflags 注入。
var version = "0.1.0"

// rootCmd 是全部子命令的根；无子命令时进入 TUI。
var rootCmd = &cobra.Command{
	Use:     "rove",
	Short:   "Rove -- open web infrastructure for AI agents",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
}

// init 注册全部子命令。
func init() {

	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(crawlCmd)
	rootCmd.AddCommand(browseCmd)
	rootCmd.AddCommand(mcpCmd)
}
