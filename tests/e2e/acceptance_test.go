/*
 * 文件作用：全量验收 -- A1-A7 汇总断言（规格书 §9），复用各里程碑 E2E 逻辑。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
	"rove/pkg/elastic"
	"rove/pkg/evidence"
	"rove/pkg/index"
	"rove/pkg/retrieval"
)

// newIndexerForTest 构造带 PseudoEmbedder 的索引器。
func newIndexerForTest(client *elastic.Client) *index.Indexer {

	return index.New(client, retrieval.NewPseudoEmbedder())
}

// newRetrieverForTest 构造带 PseudoEmbedder 的检索器。
func newRetrieverForTest(client *elastic.Client) *retrieval.Retriever {

	return retrieval.New(client, retrieval.NewPseudoEmbedder())
}

// acceptanceFixtureDoc 构造验收测试文档。
func acceptanceFixtureDoc() *document.Document {

	now := time.Now().UTC()
	return &document.Document{
		ID: "accept-1", URL: "https://accept.example.com/doc", CanonicalURL: "https://accept.example.com/doc",
		Title: "Acceptance Doc", Content: "acceptance hybrid document with keyword and semantics",
		Markdown: "# Acceptance Doc\n\nacceptance hybrid document with keyword and semantics",
		Language: "en", FetchedAt: now, Source: document.Source{Domain: "accept.example.com", Type: "web"}, ContentHash: "h-accept",
		Chunks: []document.Chunk{
			{ChunkID: "accept-1-0", DocumentID: "accept-1", Position: 0, HeadingPath: "Acceptance Doc", Content: "acceptance hybrid document with keyword and semantics", TokenCount: 7, Language: "en"},
		},
	}
}

// TestAcceptanceA3A4A5A7 汇总断言：自有 ES 检索（A3）、Hybrid 分数分解（A4）、Evidence 完整性（A5）、无 API Key（A7）。
// A1/A2 由 TestCrawlEndToEnd / TestBrowserEscalationPipeline 覆盖。
func TestAcceptanceA3A4A5A7(t *testing.T) {

	url := os.Getenv("ROVE_TEST_ES_URL")
	if url == "" {
		url = "http://localhost:9200"
	}
	client, err := elastic.New(url, "", "", "rove-e2e-acceptance")
	require.NoError(t, err)
	if err := client.Ping(context.Background()); err != nil {
		t.Skipf("ES 不可达，跳过: %v", err)
	}
	require.NoError(t, client.EnsureIndexes(context.Background(), 256))

	// 准备数据：索引一个文档（无 API Key，PseudoEmbedder -- A7）
	indexer := newIndexerForTest(client)
	doc := acceptanceFixtureDoc()
	require.NoError(t, indexer.Index(context.Background(), doc))

	// A3 + A4：Hybrid 检索（Scores 分解 lexical/vector/fusion）
	retriever := newRetrieverForTest(client)
	result, err := retriever.Search(context.Background(), &retrieval.Query{Text: "acceptance hybrid document", TopK: 5})
	require.NoError(t, err)
	require.NotEmpty(t, result.Hits, "A3: results from own ES")
	first := result.Hits[0]
	require.Contains(t, first.Scores, "lexical", "A4: lexical leg")
	require.Contains(t, first.Scores, "vector", "A4: vector leg")
	require.Contains(t, first.Scores, "fusion", "A4: fusion score")

	// A5：Evidence 完整性
	items, err := evidence.New().Build(context.Background(), result)
	require.NoError(t, err)
	require.NotEmpty(t, items)
	item := items[0]
	require.NotEmpty(t, item.URL, "A5: url")
	require.NotEmpty(t, item.Text, "A5: relevant chunk text")
	require.NotEmpty(t, item.Source.Domain, "A5: source")
	require.Positive(t, item.Score, "A5: score")
}
