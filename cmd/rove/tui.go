/*
 * 文件作用：rove 无子命令时进入 TUI -- 组装 services 并启动 Bubble Tea 程序（对话首页 + 五调试视图）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-15
 */
package main

import (
	"github.com/charmbracelet/bubbletea"

	"rove/internal/app"
	"rove/internal/config"
	"rove/internal/runtime"
	"rove/pkg/elastic"
	"rove/pkg/index"
	"rove/pkg/retrieval"
	"rove/tui"
)

// runTUI 启动 TUI（对话查询首页 + 五调试视图）。
func runTUI() error {

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	client, err := elastic.New(cfg.ES.URL, cfg.ES.Username, cfg.ES.Password, cfg.Index.Prefix)
	if err != nil {
		return err
	}
	bus := runtime.NewBus(200)
	defer bus.Close()

	indexService := app.NewIndex(index.New(client, retrieval.NewPseudoEmbedder(cfg.Index.EmbeddingDim)), bus)
	searchService := app.NewSearch(retrieval.New(client, retrieval.NewPseudoEmbedder(cfg.Index.EmbeddingDim)), bus)
	inspector := app.NewInspector(bus, indexService)

	program := tea.NewProgram(tui.New(searchService, indexService, inspector))
	_, err = program.Run()
	return err
}
