/*
 * 文件作用：RRF Fusion 纯函数测试 -- 双腿贡献求和、去重与排序（规格书 §5.1，Go 侧实现）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package retrieval

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
	"rove/pkg/elastic"
)

// chunkHit 构造测试 chunk 命中。
func chunkHit(id string, score float64) elastic.ChunkHit {

	return elastic.ChunkHit{Score: score, Chunk: document.Chunk{ChunkID: id, DocumentID: "doc-" + id, Content: id}}
}

func TestRRFMerge(t *testing.T) {

	legs := []namedLeg{
		{Name: "lexical", Hits: []elastic.ChunkHit{chunkHit("a", 3.0), chunkHit("b", 2.0)}},
		{Name: "vector", Hits: []elastic.ChunkHit{chunkHit("b", 0.9), chunkHit("c", 0.8)}},
	}

	fused := rrfMerge(legs, 60)
	require.Len(t, fused, 3, "union of both legs")

	byID := map[string]fusedChunk{}
	for _, fc := range fused {
		byID[fc.Chunk.ChunkID] = fc
	}
	// b 双腿命中：lexical 第 2（1/62）+ vector 第 1（1/61）
	require.InDelta(t, 1.0/61.0+1.0/62.0, byID["b"].Fusion, 0.0001)
	// a 仅 lexical 腿第 1 名：fusion = 1/61
	require.InDelta(t, 1.0/61.0, byID["a"].Fusion, 0.0001)
	// 腿分数分解存在
	require.Contains(t, byID["b"].LegScores, "lexical")
	require.Contains(t, byID["b"].LegScores, "vector")
	// 排序：双腿命中者优先
	require.Equal(t, "b", fused[0].Chunk.ChunkID)
}

func TestRRFMergeEmpty(t *testing.T) {

	fused := rrfMerge(nil, 60)
	require.Empty(t, fused)
}
