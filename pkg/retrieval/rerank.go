/*
 * 文件作用：Reranker -- Top-N 重排序的可替换接口与确定性启发式实现（词面重叠 + 标题命中，PRD FR-RNK-002）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package retrieval

import (
	"context"
	"sort"
	"strings"
)

// 启发式重排序权重。
const (
	rerankWeightOverlap = 0.15 // 词面重叠加分上限
	rerankWeightTitle   = 0.10 // 标题命中加分
	rerankWeightHeading = 0.05 // 标题路径命中加分
)

// Reranker 对 Top-N 再排序的可替换接口（cross-encoder/LLM/custom 留待后续）。
type Reranker interface {
	Rerank(ctx context.Context, query *Query, hits []SearchHit) ([]SearchHit, error)
}

// HeuristicReranker 确定性启发式重排序：查询词与正文的词面重叠 + 标题/标题路径命中，无模型。
type HeuristicReranker struct{}

// Rerank 对 hits 原地加分并按最终分数降序重排。
func (HeuristicReranker) Rerank(ctx context.Context, query *Query, hits []SearchHit) ([]SearchHit, error) {

	terms := tokenizeQuery(query.Text)
	if len(terms) == 0 || len(hits) == 0 {
		return hits, nil
	}
	for index := range hits {
		content := strings.ToLower(hits[index].Chunk.Content)
		title := strings.ToLower(hits[index].Document.Title)
		heading := strings.ToLower(hits[index].Chunk.HeadingPath)
		matched := 0
		titleHit := false
		headingHit := false
		for _, term := range terms {
			if strings.Contains(content, term) {
				matched++
			}
			if !titleHit && strings.Contains(title, term) {
				titleHit = true
			}
			if !headingHit && strings.Contains(heading, term) {
				headingHit = true
			}
		}
		overlap := float64(matched) / float64(len(terms))
		bonus := overlap * rerankWeightOverlap
		if titleHit {
			bonus += rerankWeightTitle
		}
		if headingHit {
			bonus += rerankWeightHeading
		}
		if hits[index].Scores == nil {
			hits[index].Scores = map[string]float64{}
		}
		hits[index].Scores["rerank"] = bonus
		hits[index].Score += bonus
	}
	sort.SliceStable(hits, func(left, right int) bool { return hits[left].Score > hits[right].Score })
	return hits, nil
}

// tokenizeQuery 分词：按空白切分并去除标点与过短 token（中文整句保留为单 token 子串匹配）。
func tokenizeQuery(text string) []string {

	var terms []string
	for _, field := range strings.Fields(strings.ToLower(text)) {
		trimmed := strings.Trim(field, ".,;:!?()[]{}\"'-")
		if len([]rune(trimmed)) >= 2 {
			terms = append(terms, trimmed)
		}
	}
	return terms
}
