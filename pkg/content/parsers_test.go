/*
 * 文件作用：pkg/content 解析层的单元测试 -- 覆盖注册表路由与 html/text/json/markdown 四种 Parser。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
)

// rawDoc 构造最小 RawDocument 测试夹具。
func rawDoc(t *testing.T, body []byte, contentType string) *document.RawDocument {

	t.Helper()
	return &document.RawDocument{
		ID:          "raw-1",
		URL:         "https://example.com/page",
		StatusCode:  200,
		ContentType: contentType,
		Body:        body,
	}
}

func TestRegistryFor(t *testing.T) {

	registry := DefaultRegistry()
	for _, contentType := range []string{"text/html", "text/html; charset=utf-8", "text/plain", "application/json", "text/markdown"} {
		_, err := registry.For(contentType)
		require.NoError(t, err, "For(%q) must resolve", contentType)
	}
	_, err := registry.For("application/unknown")
	require.Error(t, err, "unknown content type must error")
}

func TestHTMLParser(t *testing.T) {

	data, err := os.ReadFile(filepath.Join("testdata", "page.html"))
	require.NoError(t, err)

	parser, err := DefaultRegistry().For("text/html")
	require.NoError(t, err)

	pc, err := parser.Parse(context.Background(), rawDoc(t, data, "text/html"))
	require.NoError(t, err)
	require.Equal(t, "Rove Test Page", pc.Title)
	require.Equal(t, "a test page for rove", pc.Description)
	require.Equal(t, "https://example.com/rove-test", pc.CanonicalURL)
	require.Equal(t, "en", pc.Language)
	require.NotNil(t, pc.PublishedAt, "published time must be parsed")
	require.Equal(t, 2026, pc.PublishedAt.Year())
	require.Contains(t, pc.Links, "/page2")
	require.Contains(t, pc.Content, "main content of the test page")
	require.Contains(t, pc.Markdown, "Hello Rove")
}

func TestTextParser(t *testing.T) {

	parser, _ := DefaultRegistry().For("text/plain")
	pc, err := parser.Parse(context.Background(), rawDoc(t, []byte("  hello\nworld  "), "text/plain"))
	require.NoError(t, err)
	require.Equal(t, "hello\nworld", pc.Content)
}

func TestJSONParser(t *testing.T) {

	parser, _ := DefaultRegistry().For("application/json")
	pc, err := parser.Parse(context.Background(), rawDoc(t, []byte(`{"title":"api doc","a":1}`), "application/json"))
	require.NoError(t, err)
	require.Equal(t, "api doc", pc.Title)
	require.Contains(t, pc.Content, `"a": 1`)
}

func TestMarkdownParser(t *testing.T) {

	parser, _ := DefaultRegistry().For("text/markdown")
	pc, err := parser.Parse(context.Background(), rawDoc(t, []byte("# Title\n\nsome body"), "text/markdown"))
	require.NoError(t, err)
	require.Equal(t, "Title", pc.Title)
	require.Contains(t, pc.Content, "some body")
}
