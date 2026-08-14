/*
 * 文件作用：Rove TUI -- 对话查询首页 + 五视图 Runtime Debugger（搜索/抓取/浏览/索引/运行时，规格书 §11）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-15
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

// 视图页签索引。
const (
	chatTabIndex    = 0 // 对话（默认首页）
	searchTabIndex  = 1 // 搜索（调试视角）
	crawlTabIndex   = 2 // 抓取
	browseTabIndex  = 3 // 浏览
	indexTabIndex   = 4 // 索引
	runtimeTabIndex = 5 // 运行时
)

// viewTitles 六视图标题。
var viewTitles = []string{"对话", "搜索", "抓取", "浏览", "索引", "运行时"}

// Model 是 TUI 根模型。
type Model struct {
	tab          int
	searchQuery  string
	searchResult *retrieval.SearchResult
	searchErr    string
	indexStatus  string
	events       []runtime.Event
	chatInput    string
	chatMessages []chatMessage
	chatHistory  []string
	historyIndex int
	chatLoading  bool
	chatScroll   int
	search       *app.SearchService
	indexService *app.IndexService
	inspector    *app.RuntimeInspector
}

// New 构造 TUI 模型（默认进入对话视图）。
func New(search *app.SearchService, indexService *app.IndexService, inspector *app.RuntimeInspector) *Model {

	return &Model{search: search, indexService: indexService, inspector: inspector, historyIndex: -1}
}

// Init 启动时刷新索引状态。
func (m *Model) Init() tea.Cmd {

	return m.refreshIndexStatus
}

// searchMsg 是调试视图异步搜索结果。
type searchMsg struct {
	result *retrieval.SearchResult
	err    error
}

// chatAnswerMsg 是对话视图异步检索结果。
type chatAnswerMsg struct {
	message chatMessage
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

// runSearch 异步执行调试视图搜索（tea.Cmd）。
func (m *Model) runSearch() tea.Cmd {

	return func() tea.Msg {
		result, err := m.search.Search(context.Background(), &retrieval.Query{Text: m.searchQuery, TopK: 10})
		return searchMsg{result: result, err: err}
	}
}

// runChatSearch 异步执行对话检索（tea.Cmd），查询文本在回车时固化。
func (m *Model) runChatSearch(query string) tea.Cmd {

	return func() tea.Msg {
		result, err := m.search.Search(context.Background(), &retrieval.Query{Text: query, TopK: CHAT_TOP_K})
		if err != nil {
			return chatAnswerMsg{message: buildChatError(err)}
		}
		items, buildErr := m.search.Evidence(context.Background(), result)
		if buildErr != nil {
			return chatAnswerMsg{message: buildChatError(buildErr)}
		}
		return chatAnswerMsg{message: buildChatAnswer(query, items, result.Hits, result.Timings)}
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
			if m.tab == chatTabIndex && m.chatInput != "" && !m.chatLoading {
				return m, m.submitChatQuery()
			}
			if m.tab == searchTabIndex && m.searchQuery != "" {
				m.searchErr = ""
				return m, m.runSearch()
			}
		case tea.KeyCtrlL:
			if m.tab == chatTabIndex {
				m.chatMessages = nil
				m.chatScroll = 0
			}
		case tea.KeyCtrlE:
			if m.tab == chatTabIndex {
				m.toggleChatExpand()
			}
		case tea.KeyUp:
			if m.tab == chatTabIndex {
				m.stepChatHistory(-1)
			}
		case tea.KeyDown:
			if m.tab == chatTabIndex {
				m.stepChatHistory(1)
			}
		case tea.KeyPgUp:
			if m.tab == chatTabIndex {
				m.chatScroll++
			}
		case tea.KeyPgDown:
			if m.tab == chatTabIndex && m.chatScroll > 0 {
				m.chatScroll--
			}
		case tea.KeyRunes:
			if m.tab == chatTabIndex {
				m.appendChatInput(string(message.Runes))
			}
			if m.tab == searchTabIndex {
				m.searchQuery += string(message.Runes)
			}
		case tea.KeyBackspace:
			if m.tab == chatTabIndex {
				m.chatInput = trimLastRune(m.chatInput)
			}
			if m.tab == searchTabIndex && len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			}
		case tea.KeySpace:
			if m.tab == chatTabIndex {
				m.appendChatInput(" ")
			}
			if m.tab == searchTabIndex {
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
	case chatAnswerMsg:
		m.chatLoading = false
		m.chatMessages = append(m.chatMessages, message.message)
	case string:
		m.indexStatus = message
	}
	if m.inspector != nil {
		m.events = m.inspector.Events(20)
	}
	return m, nil
}

// submitChatQuery 固化查询并启动异步检索。
func (m *Model) submitChatQuery() tea.Cmd {

	query := m.chatInput
	m.chatMessages = append(m.chatMessages, chatMessage{role: userRole, text: query})
	m.chatHistory = append(m.chatHistory, query)
	m.historyIndex = -1
	m.chatInput = ""
	m.chatLoading = true
	if m.search == nil {
		m.chatMessages = append(m.chatMessages, buildChatError(fmt.Errorf("检索服务不可用")))
		m.chatLoading = false
		return nil
	}
	return m.runChatSearch(query)
}

// appendChatInput 追加输入（受 CHAT_INPUT_MAX 上限约束）。
func (m *Model) appendChatInput(text string) {

	if len([]rune(m.chatInput))+len([]rune(text)) > CHAT_INPUT_MAX {
		return
	}
	m.chatInput += text
}

// stepChatHistory 翻阅历史输入：-1 向上（更旧）、1 向下（更新）；未翻阅时按上键进入最新一条。
func (m *Model) stepChatHistory(direction int) {

	if len(m.chatHistory) == 0 {
		return
	}
	if m.historyIndex == -1 && direction == -1 {
		m.historyIndex = len(m.chatHistory) - 1
		m.chatInput = m.chatHistory[m.historyIndex]
		return
	}
	next := m.historyIndex + direction
	if next < -1 {
		next = -1
	}
	if next >= len(m.chatHistory) {
		next = -1
	}
	m.historyIndex = next
	if next == -1 {
		m.chatInput = ""
		return
	}
	m.chatInput = m.chatHistory[next]
}

// toggleChatExpand 展开/折叠最新一条回答。
func (m *Model) toggleChatExpand() {

	if len(m.chatMessages) == 0 {
		return
	}
	last := &m.chatMessages[len(m.chatMessages)-1]
	if last.role == assistantRole {
		last.expanded = !last.expanded
	}
}

// View 渲染当前视图。
func (m *Model) View() string {

	var sb strings.Builder
	sb.WriteString("Rove 终端 -- 对话查询与调试器\n")
	sb.WriteString(renderTabs(m.tab) + "\n\n")
	switch m.tab {
	case chatTabIndex:
		sb.WriteString(renderChat(m))
	case searchTabIndex:
		sb.WriteString(renderSearch(m))
	case crawlTabIndex:
		sb.WriteString(renderCrawl(m))
	case browseTabIndex:
		sb.WriteString(renderBrowse(m))
	case indexTabIndex:
		sb.WriteString(renderIndex(m))
	case runtimeTabIndex:
		sb.WriteString(renderRuntime(m))
	}
	return lipgloss.NewStyle().Padding(1).Render(sb.String())
}
