/*
 * 文件作用：Reranker 单元测试 -- 标题命中加分、空输入兜底与排序稳定性（PRD FR-RNK-002）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package retrieval

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
)

// newRerankHit 构造带正文/标题/分数的 hit。
func newRerankHit(documentID, title, content string, score float64) SearchHit {

	return SearchHit{
		Chunk:    document.Chunk{DocumentID: documentID, Content: content, HeadingPath: ""},
		Document: &document.Document{ID: documentID, Title: title},
		Score:    score,
		Scores:   map[string]float64{},
	}
}

func TestHeuristicRerankerTitleBoost(t *testing.T) {

	hits := []SearchHit{
		newRerankHit("d1", "Generic Page", "some unrelated text", 0.8),
		newRerankHit("d2", "Browser Agents Guide", "also unrelated", 0.8),
	}
	query := &Query{Text: "browser agents"}

	result, err := (HeuristicReranker{}).Rerank(context.Background(), query, hits)

	require.NoError(t, err)
	require.Equal(t, "d2", result[0].Document.ID, "标题命中者应排第一")
	require.Greater(t, result[0].Scores["rerank"], result[1].Scores["rerank"])
	require.Greater(t, result[0].Score, result[1].Score)
}

func TestHeuristicRerankerOverlapOnly(t *testing.T) {

	hits := []SearchHit{
		newRerankHit("d1", "Page A", "browser agents crawl the web", 0.8),
		newRerankHit("d2", "Page B", "nothing matches here", 0.8),
	}
	query := &Query{Text: "browser agents"}

	result, err := (HeuristicReranker{}).Rerank(context.Background(), query, hits)

	require.NoError(t, err)
	require.Equal(t, "d1", result[0].Document.ID, "词面重叠更高者应排第一")
	require.Equal(t, rerankWeightOverlap, result[0].Scores["rerank"])
}

func TestHeuristicRerankerEmptyInputs(t *testing.T) {

	reranker := HeuristicReranker{}

	emptyQuery, err := reranker.Rerank(context.Background(), &Query{Text: "  "}, []SearchHit{newRerankHit("d1", "T", "C", 0.5)})
	require.NoError(t, err)
	require.Len(t, emptyQuery, 1, "空查询应原样返回")

	emptyHits, err := reranker.Rerank(context.Background(), &Query{Text: "browser"}, nil)
	require.NoError(t, err)
	require.Empty(t, emptyHits)
}

func TestHeuristicRerankerNilScoresInitialized(t *testing.T) {

	hit := newRerankHit("d1", "Browser Doc", "browser content", 0.8)
	hit.Scores = nil
	result, err := (HeuristicReranker{}).Rerank(context.Background(), &Query{Text: "browser"}, []SearchHit{hit})

	require.NoError(t, err)
	require.NotNil(t, result[0].Scores, "Scores 为 nil 时应初始化")
	require.Contains(t, result[0].Scores, "rerank")
}

func TestQueryTermHitAsciiExtraction(t *testing.T) {

	require.True(t, QueryTermHit("什么是 npm", "npm 是包管理器", "标题", ""), "ASCII 词命中正文")
	require.False(t, QueryTermHit("什么是 npm", "vue 响应式系统介绍", "深入响应式系统", ""), "语料不含查询词应判未命中")
	require.True(t, QueryTermHit("what is npm", "npm install", "", ""), "英文查询命中")
	require.False(t, QueryTermHit("什么是 npm", "node 是运行时", "Go 教程", ""), "标题与正文都不含 npm")
}

func TestQueryTermHitCjkFallback(t *testing.T) {

	require.True(t, QueryTermHit("响应式系统", "vue 的响应式系统介绍", "", ""), "无 ASCII 词时退回中文整词匹配")
	require.False(t, QueryTermHit("响应式系统", "npm 包管理", "", ""), "中文整词未命中")
	require.True(t, QueryTermHit("", "任意内容", "", ""), "空查询不拦截")
}
