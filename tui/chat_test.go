/*
 * 文件作用：对话视图回答卡片构造纯函数测试 -- 空结果、来源截断、分数分解、错误兜底。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
	"rove/pkg/evidence"
	"rove/pkg/retrieval"
)

func TestBuildChatAnswerEmpty(t *testing.T) {

	message := buildChatAnswer("无结果查询", nil, nil, nil)
	require.Equal(t, assistantRole, message.role)
	require.Contains(t, message.text, "没有找到相关内容")
	require.Empty(t, message.sources)
}

func TestBuildChatAnswerSourcesAndScores(t *testing.T) {

	items := []evidence.Evidence{
		{Title: "文档一", URL: "https://example.com/a", Text: "直接回答内容", Score: 1.0},
		{Title: "文档二", URL: "https://example.com/b", Text: "候选二", Score: 0.8},
		{Title: "文档三", URL: "https://example.com/c", Text: "候选三", Score: 0.6},
		{Title: "文档四", URL: "https://example.com/d", Text: "候选四", Score: 0.4},
	}
	hits := []retrieval.SearchHit{
		{
			Chunk:    document.Chunk{Content: "查询直接回答内容", HeadingPath: ""},
			Document: &document.Document{Title: "文档一"},
			Scores:   map[string]float64{"lexical": 0.62, "vector": 0.45, "fusion": 0.71, "rank": 0.68},
		},
	}
	timings := map[string]time.Duration{
		"lexical_retrieve": 18 * time.Millisecond,
		"vector_retrieve":  31 * time.Millisecond,
		"rank":             4 * time.Millisecond,
	}

	message := buildChatAnswer("查询", items, hits, timings)

	require.False(t, message.lowConfidence, "词面命中时应正常作答")
	require.Equal(t, "直接回答内容", message.text)
	require.Len(t, message.sources, CHAT_SOURCE_MAX, "来源最多 3 条")
	require.Equal(t, "文档一", message.sources[0].title)
	require.Equal(t, "https://example.com/a", message.sources[0].url)
	require.Contains(t, message.scoreLine, "词汇 0.62")
	require.Contains(t, message.scoreLine, "向量 0.45")
	require.Contains(t, message.scoreLine, "融合 0.71")
	require.Contains(t, message.scoreLine, "排序 0.68")
	require.Contains(t, message.timingLine, "53ms", "耗时应为各阶段之和")
	require.NotEmpty(t, message.detail, "展开区应含分数分解与阶段耗时")
}

func TestBuildChatAnswerTruncatesLongAnswer(t *testing.T) {

	lines := make([]string, 12)
	for index := range lines {
		lines[index] = "内容行"
	}
	items := []evidence.Evidence{{Title: "长文", Text: strings.Join(lines, "\n")}}

	message := buildChatAnswer("查询", items, nil, nil)

	require.LessOrEqual(t, len(strings.Split(message.text, "\n")), CHAT_ANSWER_MAX_LINES)
	require.Contains(t, message.text, "...")
}

func TestBuildChatAnswerTruncatesLongChars(t *testing.T) {

	long := strings.Repeat("字", 400)
	items := []evidence.Evidence{{Title: "超长单行", Text: long}}

	message := buildChatAnswer("查询", items, nil, nil)

	require.LessOrEqual(t, len([]rune(message.text)), CHAT_ANSWER_MAX_CHARS+3)
	require.Contains(t, message.text, "...")
}

func TestBuildChatError(t *testing.T) {

	message := buildChatError(errors.New("连接 Elasticsearch 失败"))

	require.Equal(t, assistantRole, message.role)
	require.Contains(t, message.text, "检索失败")
	require.Equal(t, "连接 Elasticsearch 失败", message.errText)
}

func TestBuildChatAnswerNoTermMatch(t *testing.T) {

	items := []evidence.Evidence{
		{Title: "深入响应式系统", URL: "https://cn.vuejs.org/guide/extras/reactivity-in-depth", Text: "Vue 最标志性的功能就是响应式系统", Score: 1.0},
	}
	hits := []retrieval.SearchHit{
		{
			Chunk:    document.Chunk{Content: "Vue 最标志性的功能就是响应式系统", HeadingPath: ""},
			Document: &document.Document{Title: "深入响应式系统"},
			Scores:   map[string]float64{"fusion": 0.02},
		},
	}

	message := buildChatAnswer("什么是 npm", items, hits, nil)

	require.True(t, message.lowConfidence, "查询词未命中顶部结果应降级")
	require.Contains(t, message.text, "没有找到与问题直接相关的内容")
	require.Len(t, message.sources, 1, "候选来源仍保留供参考")
}

func TestBuildChatAnswerTermMatchNormal(t *testing.T) {

	items := []evidence.Evidence{
		{Title: "npm 文档", URL: "https://docs.npmjs.com", Text: "npm 是 JavaScript 的包管理器", Score: 1.0},
	}
	hits := []retrieval.SearchHit{
		{
			Chunk:    document.Chunk{Content: "npm 是 JavaScript 的包管理器", HeadingPath: ""},
			Document: &document.Document{Title: "npm 文档"},
			Scores:   map[string]float64{"fusion": 0.03},
		},
	}

	message := buildChatAnswer("什么是 npm", items, hits, nil)

	require.False(t, message.lowConfidence)
	require.Equal(t, "npm 是 JavaScript 的包管理器", message.text)
}
