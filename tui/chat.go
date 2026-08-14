/*
 * 文件作用：对话视图消息模型 -- 由检索结果构造结构化回答卡片（纯函数，无 LLM 依赖）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"rove/pkg/evidence"
	"rove/pkg/retrieval"
)

// 对话视图常量。
const (
	CHAT_TOP_K            = 5   // 每次提问检索条数
	CHAT_SOURCE_MAX       = 3   // 回答卡片最多来源数
	CHAT_ANSWER_MAX_LINES = 8   // 直接回答最多展示行数
	CHAT_ANSWER_MAX_CHARS = 300 // 直接回答最多展示字符数
	CHAT_INPUT_MAX        = 200 // 输入缓冲字符上限
	CHAT_VISIBLE_MESSAGES = 8   // 一屏最多可见消息条数
)

// 消息角色常量。
const (
	userRole      = "user"
	assistantRole = "assistant"
)

// chatSource 回答来源。
type chatSource struct {
	title string
	url   string
	score float64
}

// chatMessage 对话消息。
type chatMessage struct {
	role       string
	text       string
	sources    []chatSource
	scoreLine  string
	timingLine string
	detail     []string
	errText    string
	expanded   bool
}

// buildChatAnswer 由检索结果构造回答卡片（纯函数，便于单测）。
func buildChatAnswer(query string, items []evidence.Evidence, hits []retrieval.SearchHit, timings map[string]time.Duration) chatMessage {

	message := chatMessage{role: assistantRole}
	if len(items) == 0 {
		message.text = "没有找到相关内容，试试换种说法。"
		return message
	}
	message.text = truncateAnswer(items[0].Text)
	sourceCount := len(items)
	if sourceCount > CHAT_SOURCE_MAX {
		sourceCount = CHAT_SOURCE_MAX
	}
	for index := 0; index < sourceCount; index++ {
		message.sources = append(message.sources, chatSource{
			title: items[index].Title,
			url:   items[index].URL,
			score: items[index].Score,
		})
	}
	if len(hits) > 0 {
		message.scoreLine = formatScoreLine(hits[0].Scores)
	}
	message.timingLine = formatTimingLine(timings)
	message.detail = formatDetail(hits, timings)
	return message
}

// buildChatError 由错误构造回答卡片。
func buildChatError(err error) chatMessage {

	return chatMessage{
		role:    assistantRole,
		text:    "检索失败，请检查依赖服务（如 Elasticsearch）后重试。",
		errText: err.Error(),
	}
}

// truncateAnswer 截断过长回答：先按行数上限，再按字符上限。
func truncateAnswer(text string) string {

	lines := strings.Split(text, "\n")
	if len(lines) > CHAT_ANSWER_MAX_LINES {
		lines = lines[:CHAT_ANSWER_MAX_LINES]
		lines[CHAT_ANSWER_MAX_LINES-1] += "..."
		return strings.Join(lines, "\n")
	}
	runes := []rune(text)
	if len(runes) > CHAT_ANSWER_MAX_CHARS {
		return string(runes[:CHAT_ANSWER_MAX_CHARS]) + "..."
	}
	return text
}

// formatScoreLine 格式化顶条分数分解。
func formatScoreLine(scores map[string]float64) string {

	if len(scores) == 0 {
		return ""
	}
	parts := buildScoreParts(scores)
	if len(parts) == 0 {
		return ""
	}
	return "分数: " + strings.Join(parts, " / ")
}

// formatTimingLine 格式化总耗时（各阶段之和）。
func formatTimingLine(timings map[string]time.Duration) string {

	if len(timings) == 0 {
		return ""
	}
	var total time.Duration
	for _, duration := range timings {
		total += duration
	}
	return fmt.Sprintf("耗时 %dms", total.Milliseconds())
}

// formatDetail 构造展开区：完整分数分解与阶段耗时。
func formatDetail(hits []retrieval.SearchHit, timings map[string]time.Duration) []string {

	var detail []string
	if len(hits) > 0 {
		parts := buildScoreParts(hits[0].Scores)
		if len(parts) > 0 {
			detail = append(detail, "分数分解:")
			for _, part := range parts {
				detail = append(detail, "  "+part)
			}
		}
	}
	if len(timings) > 0 {
		detail = append(detail, "阶段耗时:")
		stageKeys := make([]string, 0, len(timings))
		for stage := range timings {
			stageKeys = append(stageKeys, stage)
		}
		sort.Strings(stageKeys)
		for _, stage := range stageKeys {
			detail = append(detail, fmt.Sprintf("  %s = %dms", stage, timings[stage].Milliseconds()))
		}
	}
	return detail
}

// buildScoreParts 按固定顺序输出分数片段（lexical/vector/fusion/rank 优先，其余按键名排序追加）。
func buildScoreParts(scores map[string]float64) []string {

	orderedLabels := []struct {
		key   string
		label string
	}{
		{key: "lexical", label: "词汇"},
		{key: "vector", label: "向量"},
		{key: "fusion", label: "融合"},
		{key: "rank", label: "排序"},
	}
	var parts []string
	consumed := map[string]bool{}
	for _, definition := range orderedLabels {
		value, ok := scores[definition.key]
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %.2f", definition.label, value))
		consumed[definition.key] = true
	}
	var rest []string
	for key := range scores {
		if !consumed[key] {
			rest = append(rest, key)
		}
	}
	sort.Strings(rest)
	for _, key := range rest {
		parts = append(parts, fmt.Sprintf("%s %.2f", key, scores[key]))
	}
	return parts
}

// trimLastRune 删除末尾一个字符（rune 安全）。
func trimLastRune(text string) string {

	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}
	return string(runes[:len(runes)-1])
}
