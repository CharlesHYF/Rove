/*
 * 文件作用：pkg/eval 检索评估指标的单元测试 -- Recall@K / NDCG@K / MRR。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package eval

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecallAtK(t *testing.T) {

	relevant := map[string]bool{"a": true, "b": true, "c": true}
	require.InDelta(t, 2.0/3.0, RecallAtK(relevant, []string{"a", "b"}, 2), 0.0001, "recall@2 against 3 relevant is 2/3")
	require.InDelta(t, 2.0/3.0, RecallAtK(relevant, []string{"a", "b", "x"}, 3), 0.0001)
	require.Equal(t, 0.0, RecallAtK(relevant, []string{"x", "y"}, 2))
}

func TestNDCGAtK(t *testing.T) {

	relevant := map[string]bool{"a": true, "b": true, "c": true}
	ideal := NDCGAtK(relevant, []string{"a", "b", "c"}, 3)
	require.InDelta(t, 1.0, ideal, 0.0001, "perfect ranking gives NDCG 1")
	partial := NDCGAtK(relevant, []string{"c", "x", "a"}, 3)
	require.Less(t, partial, ideal)
}

func TestMRR(t *testing.T) {

	relevant := map[string]bool{"b": true}
	require.InDelta(t, 0.5, MRR(relevant, []string{"a", "b", "c"}), 0.0001)
	require.Equal(t, 0.0, MRR(relevant, []string{"x", "y"}))
}
