/*
 * 文件作用：pkg/index 幂等写入的集成测试 -- 写入/幂等/更新清旧 chunks/删除/批量（NFR-007）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package index

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
	"rove/pkg/elastic"
)

// testSetup 返回隔离的索引器与测试索引前缀；ES 不可达时跳过。
func testSetup(t *testing.T) (*Indexer, *elastic.Client, string) {

	t.Helper()
	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	prefix := "rove-test-" + randSuffix()
	client, err := elastic.New(url, "", "", prefix)
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过集成测试: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))
	return New(client), client, prefix
}

// randSuffix 生成测试索引唯一后缀。
func randSuffix() string {

	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// fixtureDoc 构造测试文档。
func fixtureDoc() *document.Document {

	now := time.Now().UTC()
	return &document.Document{
		ID:           "doc-1",
		URL:          "https://example.com/guide",
		CanonicalURL: "https://example.com/guide",
		Title:        "Rove Guide",
		Content:      "Rove is an open web infrastructure for AI agents.",
		Markdown:     "# Rove Guide\n\nRove is an open web infrastructure for AI agents.",
		Language:     "en",
		FetchedAt:    now,
		Source:       document.Source{Domain: "example.com", Type: "web"},
		ContentHash:  "hash-1",
		Chunks: []document.Chunk{
			{ChunkID: "doc-1-0", DocumentID: "doc-1", Position: 0, HeadingPath: "Rove Guide", Content: "Rove is an open web infrastructure for AI agents.", TokenCount: 9, Language: "en"},
		},
	}
}

func TestIndexAndStatus(t *testing.T) {

	ctx := context.Background()
	indexer, _, _ := testSetup(t)

	require.NoError(t, indexer.Index(ctx, fixtureDoc()))

	status, err := indexer.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), status.Documents)
	require.Equal(t, int64(1), status.Chunks)
	require.Greater(t, status.SizeBytes, int64(0))
}

func TestIndexIdempotent(t *testing.T) {

	ctx := context.Background()
	indexer, _, _ := testSetup(t)

	require.NoError(t, indexer.Index(ctx, fixtureDoc()))
	require.NoError(t, indexer.Index(ctx, fixtureDoc())) // 重复写入（同 ID，内容相同）

	status, err := indexer.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), status.Documents, "document must be overwritten not duplicated")
	require.Equal(t, int64(1), status.Chunks, "chunks must be cleaned before rewrite")
}

func TestIndexUpdatesChunks(t *testing.T) {

	ctx := context.Background()
	indexer, _, _ := testSetup(t)

	doc := fixtureDoc()
	require.NoError(t, indexer.Index(ctx, doc))

	// 更新：追加一个 chunk 后重写
	doc.Chunks = append(doc.Chunks, document.Chunk{ChunkID: "doc-1-1", DocumentID: "doc-1", Position: 1, Content: "second chunk", TokenCount: 2})
	require.NoError(t, indexer.Index(ctx, doc))

	status, err := indexer.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), status.Documents)
	require.Equal(t, int64(2), status.Chunks, "stale chunks must be removed, new chunks added")
}

func TestDelete(t *testing.T) {

	ctx := context.Background()
	indexer, _, _ := testSetup(t)

	require.NoError(t, indexer.Index(ctx, fixtureDoc()))
	require.NoError(t, indexer.Delete(ctx, "doc-1"))

	status, err := indexer.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(0), status.Documents)
	require.Equal(t, int64(0), status.Chunks)
}

func TestIndexBulk(t *testing.T) {

	ctx := context.Background()
	indexer, _, _ := testSetup(t)

	require.NoError(t, indexer.IndexBulk(ctx, []*document.Document{fixtureDoc()}))

	status, err := indexer.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), status.Documents)
	require.Equal(t, int64(1), status.Chunks)
}
