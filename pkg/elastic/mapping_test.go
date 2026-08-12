/*
 * 文件作用：pkg/elastic 索引生命周期测试 -- 幂等创建 alias + mapping（规格书 §5.3）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package elastic

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureIndexes(t *testing.T) {

	ctx := context.Background()
	prefix := "rove-test-" + randSuffix()
	client, err := New(testESURL(t), "", "", prefix)
	require.NoError(t, err)
	if err := client.Ping(ctx); err != nil {
		t.Skipf("ES 不可达，跳过集成测试: %v", err)
	}

	// 首次创建
	require.NoError(t, client.EnsureIndexes(ctx, 256))
	docsExists, err := client.ExistsAlias(ctx, client.DocumentsAlias())
	require.NoError(t, err)
	require.True(t, docsExists, "documents alias must exist")
	chunksExists, err := client.ExistsAlias(ctx, client.ChunksAlias())
	require.NoError(t, err)
	require.True(t, chunksExists, "chunks alias must exist")

	// 幂等：重复调用不报错
	require.NoError(t, client.EnsureIndexes(ctx, 256))
}
