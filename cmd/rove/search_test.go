/*
 * 文件作用：cmd/rove search 与 index 命令的集成测试 -- JSON/人类可读输出与索引状态（ES 不可达时跳过）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package main

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
	"rove/pkg/elastic"
	"rove/pkg/index"
)

// esURL 读取集成测试 ES 地址。
func esURL() string {

	if url := os.Getenv("ROVE_TEST_ES_URL"); url != "" {
		return url
	}
	return "http://localhost:9200"
}

func TestSearchCommand(t *testing.T) {

	client, err := elastic.New(esURL(), "", "", "rove-cli-test")
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))

	now := time.Now().UTC()
	doc := &document.Document{
		ID: "cli-1", URL: "https://x.com/browser", CanonicalURL: "https://x.com/browser", Title: "Browser Doc",
		Content: "browser agent infrastructure", Markdown: "# Browser Doc", Language: "en", FetchedAt: now,
		Source: document.Source{Domain: "x.com", Type: "web"}, ContentHash: "h",
		Chunks: []document.Chunk{{ChunkID: "cli-1-0", DocumentID: "cli-1", Position: 0, Content: "browser agent infrastructure", TokenCount: 3, Language: "en"}},
	}
	require.NoError(t, index.New(client).Index(context.Background(), doc))

	t.Setenv("ROVE_ES_URL", esURL())
	t.Setenv("ROVE_INDEX_PREFIX", "rove-cli-test")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)

	// JSON 输出
	rootCmd.SetArgs([]string{"search", "browser", "--json"})
	require.NoError(t, rootCmd.Execute())
	require.Contains(t, out.String(), `"Title": "Browser Doc"`)
	require.Contains(t, out.String(), `"TraceID"`)

	// 人类可读输出（重置 flag 与 buffer）
	searchFlags.json = false
	out.Reset()
	rootCmd.SetArgs([]string{"search", "browser"})
	require.NoError(t, rootCmd.Execute())
	require.Contains(t, out.String(), "Browser Doc")
}

func TestIndexStatusCommand(t *testing.T) {

	client, err := elastic.New(esURL(), "", "", "rove-cli-status")
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))

	t.Setenv("ROVE_ES_URL", esURL())
	t.Setenv("ROVE_INDEX_PREFIX", "rove-cli-status")

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"index", "status"})
	require.NoError(t, rootCmd.Execute())
	require.Contains(t, out.String(), "documents")
	require.Contains(t, out.String(), "chunks")
}
