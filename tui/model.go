/*
 * 文件作用：Rove TUI -- 五视图 Runtime Debugger（Search/Crawl/Browse/Index/Runtime，规格书 §11）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package tui 提供 Bubble Tea 终端界面，订阅 Runtime Telemetry。
package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"rove/internal/app"
	"rove/internal/runtime"
	"rove/pkg/retrieval"
)

// viewTitles 五视图标题。
var viewTitles = []string{"SEARCH", "CRAWL", "BROWSE", "INDEX", "RUNTIME"}

// Model 是 TUI 根模型。
type Model struct {
	tab          int
	searchQuery  string
	searchResult *retrieval.SearchResult
	searchErr    string
	indexStatus  string
	events       []runtime.Event
	search       *app.SearchService
	indexService *app.IndexService
	inspector    *app.RuntimeInspector
}

// New 构造 TUI 模型。
func New(search *app.SearchService, indexService *app.IndexService, inspector *app.RuntimeInspector) *Model {

	return &Model{search: search, indexService: indexService, inspector: inspector}
}

// Init 启动时刷新索引状态。
func (m *Model) Init() tea.Cmd {

	return m.refreshIndexStatus
}

// searchMsg 是异步搜索结果。
type searchMsg struct {
	result *retrieval.SearchResult
	err    error
}

// refreshIndexStatus 刷新索引状态（tea.Cmd）。
func (m *Model) refreshIndexStatus() tea.Msg {

	if m.inspector == nil {
		return "index: unavailable"
	}
	status, err := m.inspector.Status(context.Background())
	if err != nil {
		return fmt.Sprintf("index: %v", err)
	}
	return fmt.Sprintf("index: health=%s docs=%d chunks=%d", status.Health, status.Documents, status.Chunks)
}

// runSearch 异步执行搜索（tea.Cmd）。
func (m *Model) runSearch() tea.Cmd {

	return func() tea.Msg {
		result, err := m.search.Search(context.Background(), &retrieval.Query{Text: m.searchQuery, TopK: 10})
		return searchMsg{result: result, err: err}
	}
}

// Update 处理消息。
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch message := msg.(type) {
	case tea.KeyMsg:
		switch message.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyTab:
			m.tab = (m.tab + 1) % len(viewTitles)
		case tea.KeyEnter:
			if m.tab == 0 && m.searchQuery != "" {
				m.searchErr = ""
				return m, m.runSearch()
			}
		case tea.KeyRunes:
			if m.tab == 0 {
				m.searchQuery += string(message.Runes)
			}
		case tea.KeyBackspace:
			if m.tab == 0 && len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			}
		case tea.KeySpace:
			if m.tab == 0 {
				m.searchQuery += " "
			}
		}
	case searchMsg:
		if message.err != nil {
			m.searchResult = nil
			m.searchErr = message.err.Error()
		} else {
			m.searchResult = message.result
			m.searchErr = ""
		}
	case string:
		m.indexStatus = message
	}
	if m.inspector != nil {
		m.events = m.inspector.Events(20)
	}
	return m, nil
}

// View 渲染当前视图。
func (m *Model) View() string {

	var sb strings.Builder
	sb.WriteString("Rove TUI -- Runtime Debugger\n")
	sb.WriteString(renderTabs(m.tab) + "\n\n")
	switch m.tab {
	case 0:
		sb.WriteString(renderSearch(m))
	case 1:
		sb.WriteString(renderCrawl(m))
	case 2:
		sb.WriteString(renderBrowse(m))
	case 3:
		sb.WriteString(renderIndex(m))
	case 4:
		sb.WriteString(renderRuntime(m))
	}
	return lipgloss.NewStyle().Padding(1).Render(sb.String())
}
