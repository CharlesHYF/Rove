/*
 * 文件作用：PseudoEmbedder 的单元测试 -- 确定性、维度、相似度方向与空文本行为（规格书 §5.2）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package retrieval

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPseudoEmbedderDeterministic(t *testing.T) {

	ctx := context.Background()
	embedder := NewPseudoEmbedder()
	require.Equal(t, 256, embedder.Dim())

	texts := []string{"browser agent infrastructure", "browser agent infrastructure"}
	vectors, err := embedder.Embed(ctx, texts)
	require.NoError(t, err)
	require.Len(t, vectors, 2)
	require.Equal(t, vectors[0], vectors[1], "same text must produce identical vector")
	require.Len(t, vectors[0], 256)
}

func TestPseudoEmbedderSimilarity(t *testing.T) {

	ctx := context.Background()
	embedder := NewPseudoEmbedder()

	base := "Rove is an open web infrastructure for AI agents. It crawls, indexes and retrieves."
	near := "Rove is an open web infrastructure for AI agents. It crawls, indexes and retrieves results."
	far := "completely different topic about cooking recipes with tomatoes"

	vectors, err := embedder.Embed(ctx, []string{base, near, far})
	require.NoError(t, err)

	similar := Cosine(vectors[0], vectors[1])
	unrelated := Cosine(vectors[0], vectors[2])
	require.Greater(t, similar, unrelated, "near text must be more similar than unrelated text")
	require.Greater(t, similar, 0.5, "near-duplicate must have high cosine")
}

func TestPseudoEmbedderEmptyText(t *testing.T) {

	ctx := context.Background()
	embedder := NewPseudoEmbedder()
	vectors, err := embedder.Embed(ctx, []string{""})
	require.NoError(t, err)
	require.Len(t, vectors, 1)
	require.Len(t, vectors[0], 256, "empty text must still yield a 256-dim vector")
}

func TestPseudoEmbedderCustomDim(t *testing.T) {

	ctx := context.Background()
	embedder := NewPseudoEmbedder(128)
	require.Equal(t, 128, embedder.Dim())

	vectors, err := embedder.Embed(ctx, []string{"browser agent", "browser agent"})
	require.NoError(t, err)
	require.Len(t, vectors, 2)
	require.Len(t, vectors[0], 128, "custom dim must produce 128-dim vectors")
	require.Equal(t, vectors[0], vectors[1], "same text must produce identical vector for custom dim")

	// 缺省 / 非法维度回退到默认 256
	require.Equal(t, 256, NewPseudoEmbedder().Dim())
	require.Equal(t, 256, NewPseudoEmbedder(0).Dim())
}
