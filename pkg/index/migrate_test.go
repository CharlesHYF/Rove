/*
 * 文件作用：索引迁移集成测试 -- create vNext -> 迁移/重嵌入 -> 校验 -> alias 切换 -> 可检索 -> 删旧（ES 不可达跳过）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package index

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/elastic"
	"rove/pkg/retrieval"
)

func TestMigrateEndToEnd(t *testing.T) {

	ctx := context.Background()
	indexer, client, _ := testSetup(t)
	require.NoError(t, indexer.Index(ctx, fixtureDoc()))

	embedder := retrieval.NewPseudoEmbedder(256)
	migrator := NewMigrator(client, embedder)

	result, err := migrator.Migrate(ctx, 256, false)
	require.NoError(t, err)
	require.Equal(t, client.DocumentsAlias()+"-v001", result.DocumentsOld)
	require.Equal(t, client.DocumentsAlias()+"-v002", result.DocumentsNew)
	require.Equal(t, int64(1), result.DocumentsMoved)
	require.Equal(t, int64(1), result.ChunksMoved)

	current, err := client.CurrentIndex(ctx, client.DocumentsAlias())
	require.NoError(t, err)
	require.Equal(t, result.DocumentsNew, current, "alias 应切换到 v002")

	// 迁移后检索仍可用（向量已重嵌入回填）
	retriever := retrieval.New(client, embedder)
	searchResult, err := retriever.Search(ctx, &retrieval.Query{Text: "open web infrastructure", TopK: 5})
	require.NoError(t, err)
	require.NotEmpty(t, searchResult.Hits, "迁移后 Hybrid 检索应可命中")

	// 二次迁移 + 删旧：旧物理索引应被删除
	second, err := migrator.Migrate(ctx, 256, true)
	require.NoError(t, err)
	require.Equal(t, client.DocumentsAlias()+"-v003", second.DocumentsNew)
	require.Error(t, client.Reindex(ctx, second.DocumentsOld, "rove-migrate-gone-dest"), "旧物理索引已删除，reindex 应失败")
}

func TestMigrateRequiresExistingIndexes(t *testing.T) {

	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	client, err := elastic.New(url, "", "", "rove-migrate-empty-"+randSuffix())
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过集成测试: %v", err)
	}

	migrator := NewMigrator(client, retrieval.NewPseudoEmbedder(256))
	_, err = migrator.Migrate(context.Background(), 256, false)
	require.Error(t, err, "未初始化索引时应报错而不是静默跳过")
}
