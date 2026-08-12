/*
 * 文件作用：pkg/elastic 客户端的单元测试 -- 非法 URL、Ping 连通性与 alias 命名契约。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package elastic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// testESURL 读取集成测试 ES 地址；未设置时使用默认地址。
func testESURL(t *testing.T) string {

	t.Helper()
	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	return url
}

func TestNewClientInvalidURL(t *testing.T) {

	_, err := New("://bad", "", "", "rove-test")
	require.Error(t, err, "invalid URL must error")
}

func TestPing(t *testing.T) {

	client, err := New(testESURL(t), "", "", "rove-test-"+randSuffix())
	require.NoError(t, err)

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Skipf("ES 不可达，跳过集成测试: %v", err)
	}
	require.NoError(t, client.Ping(ctx))
}

func TestAliasNames(t *testing.T) {

	client, err := New(testESURL(t), "", "", "rove-x")
	require.NoError(t, err)
	require.Equal(t, "rove-x-documents", client.DocumentsAlias())
	require.Equal(t, "rove-x-chunks", client.ChunksAlias())
}

// randSuffix 生成测试索引唯一后缀。
func randSuffix() string {

	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
