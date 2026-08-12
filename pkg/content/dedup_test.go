/*
 * 文件作用：去重与 Jaccard 相似度的单元测试 -- 覆盖近重复判定与三层去重的层序语义。
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

func TestJaccardSimilarity(t *testing.T) {

	base := "The quick brown fox jumps over the lazy dog near the river" // 10 个唯一 token
	near := strings.Replace(base, "quick", "fast", 1)                    // 仅 1 词不同 -> 9/11 = 0.818
	far := "quantum entanglement and neural networks"

	setBase := tokenSet(base)
	require.GreaterOrEqual(t, jaccard(setBase, tokenSet(near)), nearDupJaccardThreshold, "near-duplicate must exceed threshold")
	require.Less(t, jaccard(setBase, tokenSet(far)), nearDupJaccardThreshold, "unrelated texts must be below threshold")
	require.Equal(t, 0.0, jaccard(nil, tokenSet(far)), "empty set similarity must be 0")
}

func TestDeduperLayers(t *testing.T) {

	deduper := NewDeduper()
	ctx := context.Background()
	base := "Rove is an open web infrastructure for AI agents. It crawls, indexes and retrieves."

	firstDoc := docFixture("https://a.com/1", base)
	dup, _, err := deduper.IsDuplicate(ctx, firstDoc)
	require.NoError(t, err)
	require.False(t, dup, "first doc must not be duplicate")

	// canonical 重复
	dup, reason, _ := deduper.IsDuplicate(ctx, docFixture("https://a.com/1", base))
	require.True(t, dup)
	require.Equal(t, "canonical_duplicate", reason)

	// exact 内容重复（不同 URL）
	dup, reason, _ = deduper.IsDuplicate(ctx, docFixture("https://a.com/2", base))
	require.True(t, dup)
	require.Equal(t, "exact_content_duplicate", reason)

	// 近重复（1 词变异）
	near := strings.Replace(base, "crawls", "scans", 1)
	dup, reason, _ = deduper.IsDuplicate(ctx, docFixture("https://a.com/3", near))
	require.True(t, dup)
	require.Equal(t, "near_duplicate", reason)

	// 不同内容通过
	dup, _, err = deduper.IsDuplicate(ctx, docFixture("https://b.com/1", "completely different topic about cooking recipes"))
	require.NoError(t, err)
	require.False(t, dup, "distinct doc must pass")
}

// docFixture 构造测试 Document。
func docFixture(url, content string) *document.Document {

	return &document.Document{
		ID:           document.NewDocumentID(url),
		URL:          url,
		CanonicalURL: url,
		Content:      content,
		ContentHash:  strings.TrimSpace(content),
	}
}
