/*
 * 文件作用：ES 搜索请求体构造 -- BM25 multi_match + 可选 filter（规格书 §4.2 boost 默认值）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package retrieval

import (
	"encoding/json"
	"time"
)

// buildSearchBody 构造 chunks 索引的 BM25 搜索体。
func buildSearchBody(q *Query, candidateSize int) ([]byte, error) {

	query := map[string]any{
		"bool": map[string]any{
			"must": []any{
				map[string]any{
					"multi_match": map[string]any{
						"query":  q.Text,
						"fields": []string{"title^3", "heading_path^2", "content"},
					},
				},
			},
		},
	}

	if filters := buildFilters(q.Filters); len(filters) > 0 {
		query["bool"].(map[string]any)["filter"] = filters
	}

	body := map[string]any{
		"size":             candidateSize,
		"track_total_hits": true,
		"query":            query,
	}
	return json.Marshal(body)
}

// buildFilters 构造 bool filter 子句。
func buildFilters(f Filters) []any {

	var filters []any
	if f.Domain != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"domain": f.Domain}})
	}
	if f.Language != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"language": f.Language}})
	}
	if f.PublishedFrom != nil || f.PublishedTo != nil {
		rangeClause := map[string]any{}
		if f.PublishedFrom != nil {
			rangeClause["gte"] = f.PublishedFrom.UTC().Format(time.RFC3339)
		}
		if f.PublishedTo != nil {
			rangeClause["lte"] = f.PublishedTo.UTC().Format(time.RFC3339)
		}
		filters = append(filters, map[string]any{"range": map[string]any{"published_at": rangeClause}})
	}
	return filters
}
