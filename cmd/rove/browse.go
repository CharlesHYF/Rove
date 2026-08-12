/*
 * 文件作用：rove browse 子命令与浏览器管理器装配 -- 输出 Page State（PRD FR-BRW-002）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"rove/internal/config"
	"rove/pkg/browser"
)

var browseFlags struct {
	json bool
}

// browseCmd 实现 rove browse <url>。
var browseCmd = &cobra.Command{
	Use:   "browse <url>",
	Short: "浏览器渲染页面并输出 Page State",
	Args:  cobra.ExactArgs(1),
	RunE:  runBrowse,
}

// init 注册 browse 命令参数。
func init() {

	browseCmd.Flags().BoolVar(&browseFlags.json, "json", false, "以 JSON 输出 Page State")
}

// runBrowse 渲染页面并输出状态。
func runBrowse(cmd *cobra.Command, args []string) error {

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	manager, err := newBrowserManager(cfg)
	if err != nil {
		return err
	}
	defer manager.Close()

	state, err := manager.Browse(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	if browseFlags.json {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(state)
	}

	cmd.Printf("title: %s\n", state.Title)
	cmd.Printf("url: %s\n", state.URL)
	cmd.Printf("links: %d\n", len(state.Links))
	cmd.Printf("interactive: %d\n", len(state.Interactive))
	for _, element := range state.Interactive {
		cmd.Printf("  [%s] %s %q\n", element.ID, element.Tag, element.Text)
	}
	return nil
}

// newBrowserManager 按配置构造浏览器管理器（禁用或不可用时返回错误）。
func newBrowserManager(cfg *config.Config) (*browser.Manager, error) {

	if !cfg.Browser.Enabled {
		return nil, fmt.Errorf("browser disabled by config")
	}
	return browser.NewManager(browser.Options{
		ExecutablePath: cfg.Browser.ExecutablePath,
		Timeout:        cfg.Browser.Timeout.Duration,
	})
}
