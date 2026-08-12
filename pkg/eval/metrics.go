/*
 * 文件作用：检索评估指标 -- Recall@K / NDCG@K / MRR 纯函数（架构书 §15 检索 benchmark 骨架）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package eval 提供检索质量评估指标。
package eval

import "math"

// RecallAtK 计算前 K 个结果的召回率（相关数 / 总相关数）。
func RecallAtK(relevant map[string]bool, ranked []string, k int) float64 {

	if len(relevant) == 0 {
		return 0
	}
	if k > len(ranked) {
		k = len(ranked)
	}
	hits := 0
	for index := 0; index < k; index++ {
		if relevant[ranked[index]] {
			hits++
		}
	}
	return float64(hits) / float64(len(relevant))
}

// NDCGAtK 计算前 K 个结果的归一化折损累计增益（binary relevance）。
func NDCGAtK(relevant map[string]bool, ranked []string, k int) float64 {

	if k > len(ranked) {
		k = len(ranked)
	}
	if k == 0 {
		return 0
	}
	dcg := 0.0
	for index := 0; index < k; index++ {
		if relevant[ranked[index]] {
			dcg += 1.0 / math.Log2(float64(index+2))
		}
	}
	// IDCG：理想顺序（相关文档置于最前），取 min(k, 相关总数) 个位置
	idcg := 0.0
	idealCount := k
	if idealCount > len(relevant) {
		idealCount = len(relevant)
	}
	for index := 0; index < idealCount; index++ {
		idcg += 1.0 / math.Log2(float64(index+2))
	}
	if idcg == 0 {
		return 0
	}
	return dcg / idcg
}

// MRR 计算平均倒数排名（首个相关结果的 1/rank）。
func MRR(relevant map[string]bool, ranked []string) float64 {

	for index, doc := range ranked {
		if relevant[doc] {
			return 1.0 / float64(index+1)
		}
	}
	return 0
}
