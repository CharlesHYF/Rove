/*
 * 文件作用：pkg/retrieval 查询体构造的单元测试 -- multi_match boost 与 filter 语义（规格书 §4.2）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package retrieval

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBuildSearchBody(t *testing.T) {

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	body, err := buildSearchBody(&Query{
		Text: "browser agent",
		Filters: Filters{
			Domain:        "example.com",
			Language:      "en",
			PublishedFrom: &from,
		},
		TopK: 10,
	}, 50)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(body, &parsed))
	require.Equal(t, float64(50), parsed["size"])

	boolQuery := parsed["query"].(map[string]any)["bool"].(map[string]any)
	must := boolQuery["must"].([]any)
	multiMatch := must[0].(map[string]any)["multi_match"].(map[string]any)
	require.Equal(t, "browser agent", multiMatch["query"])
	require.Equal(t, []any{"title^3", "heading_path^2", "content"}, multiMatch["fields"])

	filter := boolQuery["filter"].([]any)
	require.Len(t, filter, 3, "domain + language + date range filters")
}

func TestBuildSearchBodyNoFilters(t *testing.T) {

	body, err := buildSearchBody(&Query{Text: "hello", TopK: 10}, 50)
	require.NoError(t, err)
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(body, &parsed))
	boolQuery := parsed["query"].(map[string]any)["bool"].(map[string]any)
	require.NotContains(t, boolQuery, "filter")
}

func TestBuildSearchBodySourceTypeFilter(t *testing.T) {

	body, err := buildSearchBody(&Query{Text: "hello", TopK: 10, Filters: Filters{SourceType: "docs"}}, 50)
	require.NoError(t, err)
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(body, &parsed))
	boolQuery := parsed["query"].(map[string]any)["bool"].(map[string]any)
	filter := boolQuery["filter"].([]any)
	require.Len(t, filter, 1)
	require.Equal(t, map[string]any{"term": map[string]any{"source_type": "docs"}}, filter[0])
}
