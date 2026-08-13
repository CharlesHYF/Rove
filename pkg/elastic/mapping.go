/*
 * 文件作用：ES 索引 mapping 与 alias 生命周期 -- 创建物理索引 v001 并注册 alias（规格书 §5.3）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"rove/pkg/rove"
)

// documentsMapping 返回文档索引 mapping。
func documentsMapping() map[string]any {

	return map[string]any{
		"settings": map[string]any{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]any{
			"properties": map[string]any{
				"id":            map[string]any{"type": "keyword"},
				"url":           map[string]any{"type": "keyword"},
				"canonical_url": map[string]any{"type": "keyword"},
				"title":         map[string]any{"type": "text", "fields": map[string]any{"keyword": map[string]any{"type": "keyword"}}},
				"description":   map[string]any{"type": "text"},
				"content":       map[string]any{"type": "text"},
				"language":      map[string]any{"type": "keyword"},
				"published_at":  map[string]any{"type": "date"},
				"fetched_at":    map[string]any{"type": "date"},
				"source": map[string]any{
					"properties": map[string]any{
						"domain": map[string]any{"type": "keyword"},
						"type":   map[string]any{"type": "keyword"},
					},
				},
				"content_hash": map[string]any{"type": "keyword"},
				"metadata":     map[string]any{"type": "flattened"},
			},
		},
	}
}

// chunksMapping 返回 chunk 索引 mapping（embedding 为 M3 预留，dims 由 embeddingDim 决定）。
func chunksMapping(embeddingDim int) map[string]any {

	return map[string]any{
		"settings": map[string]any{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]any{
			// embedding 仅用于 HNSW 向量检索，不落入 _source：避免存储翻倍与每次检索回传全量向量。
			"_source": map[string]any{
				"excludes": []string{"embedding"},
			},
			"properties": map[string]any{
				"chunk_id":     map[string]any{"type": "keyword"},
				"document_id":  map[string]any{"type": "keyword"},
				"title":        map[string]any{"type": "text"},
				"content":      map[string]any{"type": "text"},
				"heading_path": map[string]any{"type": "text", "fields": map[string]any{"keyword": map[string]any{"type": "keyword"}}},
				"position":     map[string]any{"type": "integer"},
				"language":     map[string]any{"type": "keyword"},
				"published_at": map[string]any{"type": "date"},
				"domain":       map[string]any{"type": "keyword"},
				"token_count":  map[string]any{"type": "integer"},
				"embedding":    map[string]any{"type": "dense_vector", "dims": embeddingDim, "index": true, "similarity": "cosine"},
			},
		},
	}
}

// EnsureIndexes 幂等创建文档与 chunk 索引（物理 v001 + alias）。
func (c *Client) EnsureIndexes(ctx context.Context, embeddingDim int) error {

	if err := c.ensureIndex(ctx, c.DocumentsAlias(), documentsMapping()); err != nil {
		return err
	}
	return c.ensureIndex(ctx, c.ChunksAlias(), chunksMapping(embeddingDim))
}

// ensureIndex 单个索引：alias 已存在则跳过；否则创建物理索引并注册 alias。
func (c *Client) ensureIndex(ctx context.Context, alias string, mapping map[string]any) error {

	exists, err := c.ExistsAlias(ctx, alias)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	physical := fmt.Sprintf("%s-v001", alias)
	body, err := json.Marshal(mapping)
	if err != nil {
		return rove.NewError("es.mapping", rove.CategoryIndex, false, "marshal mapping: %v", err)
	}

	resp, err := c.es.Indices.Create(physical, c.es.Indices.Create.WithContext(ctx), c.es.Indices.Create.WithBody(strings.NewReader(string(body))))
	if err != nil {
		return rove.NewError("es.create_index", rove.CategoryIndex, true, "create index %s: %v", physical, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return rove.NewError("es.create_index", rove.CategoryIndex, true, "create index %s status: %s", physical, resp.Status())
	}

	aliasBody := fmt.Sprintf(`{"actions":[{"add":{"index":%q,"alias":%q}}]}`, physical, alias)
	aliasResp, err := c.es.Indices.UpdateAliases(strings.NewReader(aliasBody), c.es.Indices.UpdateAliases.WithContext(ctx))
	if err != nil {
		return rove.NewError("es.update_aliases", rove.CategoryIndex, true, "add alias %s: %v", alias, err)
	}
	defer aliasResp.Body.Close()
	if aliasResp.IsError() {
		return rove.NewError("es.update_aliases", rove.CategoryIndex, true, "add alias %s status: %s", alias, aliasResp.Status())
	}
	return nil
}

// ExistsAlias 判断 alias 是否存在（404 视为不存在）。
func (c *Client) ExistsAlias(ctx context.Context, alias string) (bool, error) {

	indices, err := c.resolveAlias(ctx, alias)
	if err != nil {
		return false, err
	}
	return len(indices) > 0, nil
}

// resolveAlias 返回 alias 指向的物理索引名列表（404 视为空；ES 不接受直接 DELETE alias，需先解析）。
func (c *Client) resolveAlias(ctx context.Context, alias string) ([]string, error) {

	resp, err := c.es.Indices.GetAlias(c.es.Indices.GetAlias.WithContext(ctx), c.es.Indices.GetAlias.WithName(alias))
	if err != nil {
		return nil, rove.NewError("es.get_alias", rove.CategoryIndex, true, "get alias %s: %v", alias, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, nil
	}
	if resp.IsError() {
		return nil, rove.NewError("es.get_alias", rove.CategoryIndex, true, "get alias %s status: %s", alias, resp.Status())
	}
	// GET _alias/<name> 返回 { "<物理索引>": { "aliases": {...} } }，键即物理索引名。
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, rove.NewError("es.get_alias", rove.CategoryIndex, true, "decode get alias %s: %v", alias, err)
	}
	indices := make([]string, 0, len(payload))
	for name := range payload {
		indices = append(indices, name)
	}
	return indices, nil
}
