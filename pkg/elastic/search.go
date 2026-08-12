/*
 * 文件作用：ES 检索方法 -- chunk 搜索与文档 mget 回填。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"rove/pkg/document"
	"rove/pkg/rove"
)

// ChunkHit 是 chunk 索引的命中。
type ChunkHit struct {
	Score float64
	Chunk document.Chunk
}

// SearchChunks 在 chunk 索引上执行搜索，返回命中的 chunk 与分数。
func (c *Client) SearchChunks(ctx context.Context, alias string, body []byte) ([]ChunkHit, error) {

	resp, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(alias),
		c.es.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, rove.NewError("retrieval.search", rove.CategoryIndex, true, "search %s: %v", alias, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, rove.NewError("retrieval.search", rove.CategoryIndex, true, "search %s status: %s", alias, resp.Status())
	}

	var payload struct {
		Hits struct {
			Hits []struct {
				Source map[string]any `json:"_source"`
				Score  float64        `json:"_score"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, rove.NewError("retrieval.search", rove.CategoryIndex, true, "decode search: %v", err)
	}

	hits := make([]ChunkHit, 0, len(payload.Hits.Hits))
	for _, hit := range payload.Hits.Hits {
		chunk := parseChunkSource(hit.Source)
		hits = append(hits, ChunkHit{Score: hit.Score, Chunk: chunk})
	}
	return hits, nil
}

// MGetDocuments 按 id 批量取文档，返回 document_id -> Document。
func (c *Client) MGetDocuments(ctx context.Context, alias string, ids []string) (map[string]*document.Document, error) {

	result := map[string]*document.Document{}
	if len(ids) == 0 {
		return result, nil
	}
	body, _ := json.Marshal(map[string]any{"ids": ids})
	resp, err := c.es.Mget(bytes.NewReader(body), c.es.Mget.WithContext(ctx), c.es.Mget.WithIndex(alias))
	if err != nil {
		return nil, rove.NewError("retrieval.mget", rove.CategoryIndex, true, "mget %s: %v", alias, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, rove.NewError("retrieval.mget", rove.CategoryIndex, true, "mget %s status: %s", alias, resp.Status())
	}

	var payload struct {
		Docs []struct {
			Found  bool           `json:"found"`
			Source map[string]any `json:"_source"`
		} `json:"docs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, rove.NewError("retrieval.mget", rove.CategoryIndex, true, "decode mget: %v", err)
	}
	for _, doc := range payload.Docs {
		if !doc.Found {
			continue
		}
		parsed := parseDocumentSource(doc.Source)
		result[parsed.ID] = parsed
	}
	return result, nil
}

// parseChunkSource 从 _source 还原 Chunk。
func parseChunkSource(source map[string]any) document.Chunk {

	return document.Chunk{
		ChunkID:     str(source["chunk_id"]),
		DocumentID:  str(source["document_id"]),
		Position:    int(num(source["position"])),
		HeadingPath: str(source["heading_path"]),
		Content:     str(source["content"]),
		TokenCount:  int(num(source["token_count"])),
		Language:    str(source["language"]),
	}
}

// parseDocumentSource 从 _source 还原 Document。
func parseDocumentSource(source map[string]any) *document.Document {

	sourceInfo, _ := source["source"].(map[string]any)
	return &document.Document{
		ID:           str(source["id"]),
		URL:          str(source["url"]),
		CanonicalURL: str(source["canonical_url"]),
		Title:        str(source["title"]),
		Description:  str(source["description"]),
		Content:      str(source["content"]),
		Language:     str(source["language"]),
		FetchedAt:    timeNow(source["fetched_at"]),
		PublishedAt:  timePtr(source["published_at"]),
		Source: document.Source{
			Domain: str(sourceInfo["domain"]),
			Type:   str(sourceInfo["type"]),
		},
		ContentHash: str(source["content_hash"]),
	}
}

// str 安全取字符串字段。
func str(value any) string {

	s, _ := value.(string)
	return s
}

// num 安全取数值字段。
func num(value any) float64 {

	n, _ := value.(float64)
	return n
}

// timeNow 解析 RFC3339 时间，失败返回零值。
func timeNow(value any) time.Time {

	if s, ok := value.(string); ok {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// timePtr 解析 RFC3339 时间为指针，失败返回 nil。
func timePtr(value any) *time.Time {

	if s, ok := value.(string); ok {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return &t
		}
	}
	return nil
}
