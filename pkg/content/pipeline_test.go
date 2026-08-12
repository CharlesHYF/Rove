/*
 * 文件作用：Content Pipeline 的单元测试 -- 验证 RawDocument 到标准 Document 的完整转换、去重语义与错误路由。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
	"rove/pkg/rove"
)

// testPipeline 构造默认管线。
func testPipeline() *Pipeline {

	return NewPipeline(DefaultRegistry(), &Canonicalizer{}, NewDeduper(), NewChunker(800, 100))
}

// readFixture 读取测试夹具文件。
func readFixture(t *testing.T, name string) []byte {

	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return data
}

func TestPipelineProducesDocument(t *testing.T) {

	raw := &document.RawDocument{
		ID:          "raw-1",
		URL:         "https://example.com/article",
		StatusCode:  200,
		ContentType: "text/html",
		Body:        readFixture(t, "page.html"),
		FetchedAt:   time.Now().UTC(),
	}

	doc, result, err := testPipeline().Process(context.Background(), raw, "web")
	require.NoError(t, err)
	require.False(t, result.IsDuplicate, "first process must not be duplicate")
	require.Equal(t, "Rove Test Page", doc.Title)
	require.Equal(t, "https://example.com/rove-test", doc.CanonicalURL)
	require.Equal(t, "example.com", doc.Source.Domain)
	require.Equal(t, "web", doc.Source.Type)
	require.Len(t, doc.ContentHash, 64, "ContentHash must be sha256 hex")
	require.NotEmpty(t, doc.Chunks, "expected chunks")

	for index, chunk := range doc.Chunks {
		require.Equal(t, document.ChunkID(doc.ID, index), chunk.ChunkID)
		require.Equal(t, doc.ID, chunk.DocumentID)
	}
}

func TestPipelineDuplicate(t *testing.T) {

	pipeline := testPipeline()
	ctx := context.Background()
	makeRaw := func(url string) *document.RawDocument {
		return &document.RawDocument{
			ID:          "raw-" + url,
			URL:         url,
			StatusCode:  200,
			ContentType: "text/plain",
			Body:        []byte("identical content here"),
			FetchedAt:   time.Now().UTC(),
		}
	}

	_, result, err := pipeline.Process(ctx, makeRaw("https://x.com/1"), "web")
	require.NoError(t, err)
	require.False(t, result.IsDuplicate)

	_, result, err = pipeline.Process(ctx, makeRaw("https://x.com/1"), "web")
	require.NoError(t, err)
	require.True(t, result.IsDuplicate)
	require.Equal(t, "canonical_duplicate", result.Reason)

	_, result, err = pipeline.Process(ctx, makeRaw("https://x.com/2"), "web")
	require.NoError(t, err)
	require.True(t, result.IsDuplicate)
	require.Equal(t, "exact_content_duplicate", result.Reason)
}

func TestPipelineUnsupportedContentType(t *testing.T) {

	raw := &document.RawDocument{ID: "raw-1", URL: "https://x.com/f", ContentType: "application/unknown", Body: []byte("x")}
	_, _, err := testPipeline().Process(context.Background(), raw, "web")

	var re *rove.Error
	require.ErrorAs(t, err, &re)
	require.Equal(t, "content.unsupported_type", re.Code)
}
