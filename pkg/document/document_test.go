/*
 * 文件作用：pkg/document 数据模型的单元测试 -- 验证幂等 ID 生成的确定性与格式。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package document

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDocumentIDDeterministic(t *testing.T) {

	a := NewDocumentID("https://example.com/doc")
	b := NewDocumentID("https://example.com/doc")
	require.Equal(t, a, b, "same canonical URL must produce same id")
	require.Len(t, a, 16, "id must be 16 hex chars")
	require.NotEqual(t, NewDocumentID("https://example.com/other"), a, "different URLs must produce different ids")
}

func TestChunkID(t *testing.T) {

	require.Equal(t, "abc123-0", ChunkID("abc123", 0))
	require.Equal(t, "abc123-12", ChunkID("abc123", 12))
}
