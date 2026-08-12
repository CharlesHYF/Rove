/*
 * 文件作用：RRF Fusion -- 合并 BM25 与向量双腿候选，分数可分解（规格书 §5.1，Go 侧实现，Charles 已确认）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package retrieval

import (
	"sort"

	"rove/pkg/document"
	"rove/pkg/elastic"
)

// rankConstant RRF 常数（与 ES 原生默认一致）。
const rankConstant = 60

// namedLeg 是一条带名称的候选腿。
type namedLeg struct {
	Name string
	Hits []elastic.ChunkHit
}

// fusedChunk 融合后的候选 chunk。
type fusedChunk struct {
	Chunk     document.Chunk
	LegScores map[string]float64 // 各腿原始分数（lexical/vector）
	Fusion    float64            // RRF 融合分
}

// rrfMerge 按 Reciprocal Rank Fusion 合并双腿：每条腿贡献 1/(rankConstant+rank)，rank 从 1 起。
func rrfMerge(legs []namedLeg, rankConstant int) []fusedChunk {

	byChunk := map[string]*fusedChunk{}
	var order []string

	for _, leg := range legs {
		for rank, hit := range leg.Hits {
			fused, ok := byChunk[hit.Chunk.ChunkID]
			if !ok {
				fused = &fusedChunk{Chunk: hit.Chunk, LegScores: map[string]float64{}}
				byChunk[hit.Chunk.ChunkID] = fused
				order = append(order, hit.Chunk.ChunkID)
			}
			fused.LegScores[leg.Name] = hit.Score
			fused.Fusion += 1.0 / float64(rankConstant+rank+1)
		}
	}

	result := make([]fusedChunk, 0, len(order))
	for _, id := range order {
		result = append(result, *byChunk[id])
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Fusion > result[j].Fusion })
	return result
}
