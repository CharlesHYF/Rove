/*
 * 文件作用：pkg/browser PageState 的集成测试 -- 稳定 element id（两次加载一致）与链接/文本提取。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package browser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// interactiveFixture 返回含表单/链接/按钮的静态页面。
func interactiveFixture() string {

	return `<html><head><title>Form Page</title></head><body>
<main>
<h1>Form</h1>
<a href="/page2">Link Two</a>
<a href="/page3">Link Three</a>
<form><input type="text" name="q"><button type="submit">Search</button></form>
</main></body></html>`
}

func TestPageStateStableIDs(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(interactiveFixture()))
	}))
	defer server.Close()

	manager := testManager(t)
	defer manager.Close()
	ctx := context.Background()

	first, err := manager.Browse(ctx, server.URL)
	require.NoError(t, err)
	second, err := manager.Browse(ctx, server.URL)
	require.NoError(t, err)

	require.NotEmpty(t, first.Interactive, "must find interactive elements")
	require.Equal(t, len(first.Interactive), len(second.Interactive))
	for index := range first.Interactive {
		require.Equal(t, first.Interactive[index].ID, second.Interactive[index].ID, "element id must be stable across loads (index %d)", index)
	}
	require.NotEmpty(t, first.Text)
	require.Contains(t, first.Links, server.URL+"/page2")
	require.Equal(t, "Form Page", first.Title)
	require.Equal(t, "Search", first.Interactive[len(first.Interactive)-1].Text, "button text captured")
}
