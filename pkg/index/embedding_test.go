/*
 * 文件作用：pkg/index 嵌入接线的集成测试 -- 可选 embedder 填充 chunk 向量，未配置时跳过。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package index

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/retrieval"
)

func TestIndexWithEmbedder(t *testing.T) {

	ctx := context.Background()
	_, client, _ := testSetup(t)
	indexer := New(client, retrieval.NewPseudoEmbedder())

	doc := fixtureDoc()
	require.Nil(t, doc.Chunks[0].Embedding, "fixture chunk starts without embedding")

	require.NoError(t, indexer.Index(ctx, doc))
	require.NotNil(t, doc.Chunks[0].Embedding, "embedder must fill chunk embedding")
	require.Len(t, doc.Chunks[0].Embedding, 256)
}

func TestIndexWithoutEmbedder(t *testing.T) {

	ctx := context.Background()
	_, client, _ := testSetup(t)
	indexer := New(client) // 无 embedder

	doc := fixtureDoc()
	require.NoError(t, indexer.Index(ctx, doc))
	require.Nil(t, doc.Chunks[0].Embedding, "chunk embedding stays nil without embedder")
}
