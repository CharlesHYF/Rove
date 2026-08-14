/*
 * 文件作用：TUI 模型单元测试 -- 对话交互、tab 切换、查询输入与视图渲染（纯逻辑，不跑真实终端）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-15
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

func TestModelDefaultTabIsChat(t *testing.T) {

	model := newTestModel()
	require.Equal(t, chatTabIndex, model.tab, "默认进入对话视图")
	require.Contains(t, model.View(), "对话")
}

func TestModelTabSwitch(t *testing.T) {

	model := newTestModel()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	require.Equal(t, searchTabIndex, updated.(*Model).tab, "Tab 切换到下一个视图")
	require.Contains(t, updated.View(), "搜索", "所有视图渲染中文标题")
}

func TestModelViewSections(t *testing.T) {

	view := newTestModel().View()
	require.Contains(t, view, "对话")
	require.Contains(t, view, "搜索")
	require.Contains(t, view, "抓取")
	require.Contains(t, view, "浏览")
	require.Contains(t, view, "索引")
	require.Contains(t, view, "运行时")
}

func TestModelChatInputTyping(t *testing.T) {

	model := newTestModel()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("浏览器")})
	require.Equal(t, "浏览器", updated.(*Model).chatInput)
}

func TestModelChatInputLimit(t *testing.T) {

	model := newTestModel()
	model.chatInput = string(make([]rune, CHAT_INPUT_MAX))
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	require.Equal(t, CHAT_INPUT_MAX, len([]rune(updated.(*Model).chatInput)), "输入超过上限应被忽略")
}

func TestModelChatEnterSubmits(t *testing.T) {

	model := newTestModel()
	model.chatInput = "browser agents 是什么"

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	state := updated.(*Model)
	require.NotNil(t, cmd, "回车应启动异步检索")
	require.Len(t, state.chatMessages, 1)
	require.Equal(t, userRole, state.chatMessages[0].role)
	require.Equal(t, "browser agents 是什么", state.chatMessages[0].text)
	require.Equal(t, "", state.chatInput)
	require.True(t, state.chatLoading)
	require.Equal(t, []string{"browser agents 是什么"}, state.chatHistory)
}

func TestModelChatEnterEmptyIgnored(t *testing.T) {

	model := newTestModel()
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	require.Nil(t, cmd)
	require.Empty(t, updated.(*Model).chatMessages)
}

func TestModelChatCtrlLClears(t *testing.T) {

	model := newTestModel()
	model.chatMessages = []chatMessage{{role: userRole, text: "旧问题"}, {role: assistantRole, text: "旧回答"}}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlL})

	require.Empty(t, updated.(*Model).chatMessages)
	require.Equal(t, 0, updated.(*Model).chatScroll)
}

func TestModelChatHistoryNavigation(t *testing.T) {

	model := newTestModel()
	model.chatHistory = []string{"问题一", "问题二"}
	model.historyIndex = -1

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	require.Equal(t, "问题二", updated.(*Model).chatInput)

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyUp})
	require.Equal(t, "问题一", updated.(*Model).chatInput)

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	require.Equal(t, "问题二", updated.(*Model).chatInput)

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	require.Equal(t, "", updated.(*Model).chatInput, "越过最新历史回到空白输入")
}

func TestModelChatScrollKeys(t *testing.T) {

	model := newTestModel()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	require.Equal(t, 1, updated.(*Model).chatScroll)

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	require.Equal(t, 0, updated.(*Model).chatScroll)

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	require.Equal(t, 0, updated.(*Model).chatScroll, "翻到顶部后不再变化")
}

func TestModelChatExpandToggle(t *testing.T) {

	model := newTestModel()
	model.chatMessages = []chatMessage{
		{role: userRole, text: "问题"},
		{role: assistantRole, text: "回答", detail: []string{"分数分解:", "  词汇 = 0.62"}},
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	require.True(t, updated.(*Model).chatMessages[1].expanded, "展开最新回答的分数分解")

	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	require.False(t, updated.(*Model).chatMessages[1].expanded, "再次按键折叠")
}

func TestModelChatAnswerAppends(t *testing.T) {

	model := newTestModel()
	model.chatLoading = true

	updated, _ := model.Update(chatAnswerMsg{message: chatMessage{role: assistantRole, text: "回答内容"}})

	state := updated.(*Model)
	require.False(t, state.chatLoading)
	require.Len(t, state.chatMessages, 1)
	require.Equal(t, "回答内容", state.chatMessages[0].text)
}

func TestModelChatRendering(t *testing.T) {

	model := newTestModel()
	model.chatMessages = []chatMessage{
		{role: userRole, text: "问题"},
		{role: assistantRole, text: "回答", sources: []chatSource{{title: "来源标题", url: "https://example.com", score: 0.9}}, scoreLine: "分数: 融合 0.71", timingLine: "耗时 53ms"},
	}

	view := model.View()

	require.Contains(t, view, "你: 问题")
	require.Contains(t, view, "答: 回答")
	require.Contains(t, view, "来源标题")
	require.Contains(t, view, "https://example.com")
	require.Contains(t, view, "相关度 0.90")
}

func TestModelSearchInput(t *testing.T) {

	model := newTestModel()
	model.tab = searchTabIndex
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("browser")})
	require.Equal(t, "browser", updated.(*Model).searchQuery)
}

func TestModelSearchResultRendering(t *testing.T) {

	model := newTestModel()
	model.tab = searchTabIndex
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
