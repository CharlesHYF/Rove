/*
 * 文件作用：TUI 模型单元测试 -- tab 切换、查询输入与视图节（纯逻辑，不跑真实终端）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package tui

import (
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"rove/internal/app"
	"rove/internal/runtime"
	"rove/pkg/document"
	"rove/pkg/retrieval"
)

// newTestModel 构造无外部依赖的 TUI 模型（services 可 nil，视图渲染不触达）。
func newTestModel() *Model {

	bus := runtime.NewBus(20)
	return New(
		&app.SearchService{},
		&app.IndexService{},
		app.NewInspector(bus, &app.IndexService{}),
	)
}

func TestModelTabSwitch(t *testing.T) {

	model := newTestModel()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	require.Equal(t, 1, updated.(*Model).tab, "Tab switches to next view")
	require.Contains(t, updated.View(), "SEARCH", "all views render a header")
}

func TestModelSearchInput(t *testing.T) {

	model := newTestModel()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("browser")})
	require.Equal(t, "browser", updated.(*Model).searchQuery)
}

func TestModelViewSections(t *testing.T) {

	view := newTestModel().View()
	require.Contains(t, view, "SEARCH")
	require.Contains(t, view, "CRAWL")
	require.Contains(t, view, "BROWSE")
	require.Contains(t, view, "INDEX")
	require.Contains(t, view, "RUNTIME")
}

func TestModelSearchResultRendering(t *testing.T) {

	model := newTestModel()
	model.searchResult = &retrieval.SearchResult{
		Query: "browser",
		Hits: []retrieval.SearchHit{
			{
				Document: &document.Document{ID: "d1", URL: "https://example.com/a", Title: "Browser Doc"},
				Chunk:    document.Chunk{ChunkID: "d1-0", DocumentID: "d1", Content: "browser agent"},
			},
		},
	}
	view := model.View()
	require.Contains(t, view, "browser")
	require.Contains(t, view, "Browser Doc")
}
