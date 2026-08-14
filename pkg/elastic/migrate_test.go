/*
 * 文件作用：ES 迁移底层操作测试 -- 版本命名纯函数与物理索引生命周期（含 ES 集成，不可达则跳过）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package elastic

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNextVersionName(t *testing.T) {

	next, err := NextVersionName("rove-chunks", "rove-chunks-v001")
	require.NoError(t, err)
	require.Equal(t, "rove-chunks-v002", next)

	next, err = NextVersionName("rove-documents", "rove-documents-v009")
	require.NoError(t, err)
	require.Equal(t, "rove-documents-v010", next, "版本号应补零进位")
}

func TestNextVersionNameInvalid(t *testing.T) {

	_, err := NextVersionName("rove-chunks", "rove-chunks")
	require.Error(t, err, "缺少版本后缀应报错")

	_, err = NextVersionName("rove-chunks", "rove-chunks-vabc")
	require.Error(t, err, "非法版本号应报错")
}

// esURL 读取测试用 ES 地址（CI 注入 ROVE_TEST_ES_URL）。
func esURL() string {

	if url := os.Getenv("ROVE_TEST_ES_URL"); url != "" {
		return url
	}
	return "http://localhost:9200"
}

// TestPhysicalIndexLifecycle 集成验证：创建物理索引 -> reindex -> alias 切换 -> 删除（ES 不可达跳过）。
func TestPhysicalIndexLifecycle(t *testing.T) {

	prefix := "rove-migrate-elastic-" + t.Name()
	client, err := New(esURL(), "", "", prefix)
	require.NoError(t, err)
	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skipf("ES 不可达，跳过: %v", err)
	}

	oldIndex := prefix + "-documents-v001"
	newIndex := prefix + "-documents-v002"
	require.NoError(t, client.CreateDocumentsIndex(ctx, oldIndex))
	require.NoError(t, client.CreateDocumentsIndex(ctx, newIndex))
	t.Cleanup(func() {
		_ = client.DeletePhysical(context.Background(), oldIndex)
		_ = client.DeletePhysical(context.Background(), newIndex)
	})

	require.NoError(t, client.Reindex(ctx, oldIndex, newIndex))
	require.NoError(t, client.SwitchAlias(ctx, prefix+"-documents", oldIndex, newIndex))

	current, err := client.CurrentIndex(ctx, prefix+"-documents")
	require.NoError(t, err)
	require.Equal(t, newIndex, current, "alias 应原子切换到新物理索引")

	require.NoError(t, client.DeletePhysical(ctx, oldIndex))
}
