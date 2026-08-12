/*
 * 文件作用：Chunker 的单元测试 -- 覆盖 heading_path、position、词窗口切分与 overlap 语义。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
)

func TestChunkerHeadingPathAndPosition(t *testing.T) {

	markdown := "# Install\n\nRun this command.\n\n## Quick Start\n\nDo the thing.\n\n# FAQ\n\nQuestion here."
	doc := &document.Document{ID: "doc1", Content: markdown, Markdown: markdown}

	chunks, err := NewChunker(1000, 0).Chunk(context.Background(), doc)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(chunks), 3, "each heading section must become a chunk")

	for index, chunk := range chunks {
		require.Equal(t, document.ChunkID("doc1", index), chunk.ChunkID)
		require.Equal(t, "doc1", chunk.DocumentID)
		require.Equal(t, index, chunk.Position)
		require.Positive(t, chunk.TokenCount)
	}

	require.Equal(t, "Install", chunks[0].HeadingPath)
	require.Equal(t, "Install > Quick Start", chunks[1].HeadingPath)
	require.Equal(t, "FAQ", chunks[2].HeadingPath)
}

func TestChunkerMaxTokens(t *testing.T) {

	markdown := "# A\n\n" + strings.Repeat("word ", 500) + "\n\n# B\n\n" + strings.Repeat("term ", 500)
	doc := &document.Document{ID: "doc2", Content: markdown, Markdown: markdown}

	chunks, err := NewChunker(300, 0).Chunk(context.Background(), doc)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(chunks), 4, "500-word block must split into >= 2 chunks each")
	for _, chunk := range chunks {
		require.LessOrEqual(t, chunk.TokenCount, 300, "chunk must not exceed max tokens")
	}
}

func TestChunkerOverlap(t *testing.T) {

	markdown := "# A\n\n" + strings.Repeat("word ", 1000)
	chunks, err := NewChunker(300, 80).Chunk(context.Background(), &document.Document{ID: "doc3", Markdown: markdown})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(chunks), 2, "long text must split")
	require.Contains(t, chunks[1].Content, "word", "overlap seed must carry previous content")
	require.NotEqual(t, strings.TrimSpace(chunks[0].Content), strings.TrimSpace(chunks[1].Content), "adjacent chunks must differ")
}
