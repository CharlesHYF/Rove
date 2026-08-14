/*
 * 文件作用：TUI 视图渲染 -- 对话首页与五调试视图，内容来自注入的 services 与 telemetry 事件。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-15
 */
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"rove/internal/runtime"
)

var (
	activeTabStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	inactiveTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// renderTabs 渲染视图标签栏。
func renderTabs(active int) string {

	var tabs []string
	for index, title := range viewTitles {
		if index == active {
			tabs = append(tabs, activeTabStyle.Render(title))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(title))
		}
	}
	return strings.Join(tabs, "  ")
}

// renderChat 渲染对话视图（默认首页）。
func renderChat(m *Model) string {

	var sb strings.Builder
	sb.WriteString("输入问题后回车检索（自有索引，无需 API Key）；Ctrl+E 展开分数，Ctrl+L 清空，Tab 切视图\n\n")

	total := len(m.chatMessages)
	end := total - m.chatScroll
	if end < 0 {
		end = 0
	}
	start := end - CHAT_VISIBLE_MESSAGES
	if start < 0 {
		start = 0
	}
	if m.chatScroll > 0 && start > 0 {
		sb.WriteString(fmt.Sprintf("(向上翻页 %d 屏)\n", m.chatScroll))
	}
	for index := start; index < end; index++ {
		sb.WriteString(renderChatMessage(m.chatMessages[index]) + "\n")
	}
	if m.chatLoading {
		sb.WriteString("检索中...\n")
	}
	sb.WriteString("\n你: " + m.chatInput)
	return sb.String()
}

// renderChatMessage 渲染单条对话消息。
func renderChatMessage(message chatMessage) string {

	if message.role == userRole {
		return "你: " + message.text
	}
	var sb strings.Builder
	sb.WriteString("答: " + message.text + "\n")
	if len(message.sources) > 0 {
		sb.WriteString("来源:\n")
		for index, source := range message.sources {
			sb.WriteString(fmt.Sprintf(" %d. %s -- %s（相关度 %.2f）\n", index+1, source.title, source.url, source.score))
		}
	}
	if message.scoreLine != "" {
		sb.WriteString(message.scoreLine + "\n")
	}
	if message.timingLine != "" {
		sb.WriteString(message.timingLine + "\n")
	}
	if message.expanded {
		for _, line := range message.detail {
			sb.WriteString(line + "\n")
		}
	}
	if message.errText != "" {
		sb.WriteString("错误: " + message.errText + "\n")
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

// renderSearch 渲染检索视图。
func renderSearch(m *Model) string {

	var sb strings.Builder
	sb.WriteString("query> " + m.searchQuery + "\n\n")
	if m.searchErr != "" {
		sb.WriteString("error: " + m.searchErr + "\n\n")
	}
	if m.searchResult == nil {
		sb.WriteString("(输入查询后回车执行；Tab 切换视图)\n")
		return sb.String()
	}
	sb.WriteString(fmt.Sprintf("query: %s\n", m.searchResult.Query))
	sb.WriteString(fmt.Sprintf("trace %s  timings: ", m.searchResult.TraceID))
	for key, value := range m.searchResult.Timings {
		sb.WriteString(fmt.Sprintf("%s=%dms ", key, value.Milliseconds()))
	}
	sb.WriteString("\n\n")
	for rank, hit := range m.searchResult.Hits {
		sb.WriteString(fmt.Sprintf("%d. %s [%.3f] fusion=%.3f\n   %s\n", rank+1, hit.Document.Title, hit.Score, hit.Scores["fusion"], hit.Document.URL))
	}
	return sb.String()
}

// renderCrawl 渲染抓取视图。
func renderCrawl(m *Model) string {

	var sb strings.Builder
	sb.WriteString("crawl events:\n")
	for _, event := range m.events {
		if event.Type == runtime.Crawl {
			sb.WriteString(fmt.Sprintf("  %s %s\n", event.At.Format("15:04:05"), event.Message))
		}
	}
	if strings.Contains(sb.String(), "crawl events:\n  ") == false {
		sb.WriteString("  (无 crawl 事件，运行 rove crawl 后可见)\n")
	}
	return sb.String()
}

// renderBrowse 渲染浏览视图。
func renderBrowse(m *Model) string {

	var sb strings.Builder
	sb.WriteString("browse events:\n")
	for _, event := range m.events {
		if event.Type == runtime.Fetch {
			sb.WriteString(fmt.Sprintf("  %s %s\n", event.At.Format("15:04:05"), event.Message))
		}
	}
	sb.WriteString("  (用 `rove browse <url>` 在 CLI 查看 Page State)\n")
	return sb.String()
}

// renderIndex 渲染索引视图。
func renderIndex(m *Model) string {

	status := m.indexStatus
	if status == "" {
		status = "index: loading..."
	}
	var sb strings.Builder
	sb.WriteString(status + "\n\nindex events:\n")
	for _, event := range m.events {
		if event.Type == runtime.Index {
			sb.WriteString(fmt.Sprintf("  %s %s\n", event.At.Format("15:04:05"), event.Message))
		}
	}
	return sb.String()
}

// renderRuntime 渲染运行时事件日志。
func renderRuntime(m *Model) string {

	var sb strings.Builder
	sb.WriteString("runtime events (最近 20 条):\n")
	for _, event := range m.events {
		sb.WriteString(fmt.Sprintf("  %s [%s] %s\n", event.At.Format("15:04:05"), event.Type, event.Message))
	}
	return sb.String()
}
