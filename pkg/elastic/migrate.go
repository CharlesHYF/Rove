/*
 * 文件作用：ES 索引迁移底层操作 -- 版本命名、物理索引创建、reindex、alias 原子切换与全量扫描（规格书 §5.3）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"rove/pkg/rove"
)

// 迁移相关常量。
const (
	versionSuffixPrefix = "-v" // 物理索引版本后缀前缀
	scanDefaultSize     = 128  // 全量扫描默认分页大小
	migrateVersionLen   = 3    // 版本号位数（-v001 格式）
)

// CurrentIndex 返回 alias 当前指向的物理索引名；不存在返回空串。
func (c *Client) CurrentIndex(ctx context.Context, alias string) (string, error) {

	indices, err := c.resolveAlias(ctx, alias)
	if err != nil {
		return "", err
	}
	if len(indices) == 0 {
		return "", nil
	}
	return indices[0], nil
}

// NextVersionName 计算下一个物理索引名：-vNNN 后缀递增（纯函数，便于单测）。
func NextVersionName(alias, current string) (string, error) {

	prefix := alias + versionSuffixPrefix
	if !strings.HasPrefix(current, prefix) {
		return "", fmt.Errorf("index %s 缺少版本后缀(-vNNN)", current)
	}
	suffix := strings.TrimPrefix(current, prefix)
	var version int
	if _, err := fmt.Sscanf(suffix, "%d", &version); err != nil || version <= 0 {
		return "", fmt.Errorf("index %s 版本后缀非法", current)
	}
	return fmt.Sprintf("%s-v%0*d", alias, migrateVersionLen, version+1), nil
}

// CreateDocumentsIndex 创建指定物理名的文档索引（mapping 与 v001 一致）。
func (c *Client) CreateDocumentsIndex(ctx context.Context, physical string) error {

	return c.createPhysical(ctx, physical, documentsMapping())
}

// CreateChunksIndex 创建指定物理名的 chunk 索引。
func (c *Client) CreateChunksIndex(ctx context.Context, physical string, embeddingDim int) error {

	return c.createPhysical(ctx, physical, chunksMapping(embeddingDim))
}

// createPhysical 按 mapping 创建物理索引（不注册 alias）。
func (c *Client) createPhysical(ctx context.Context, physical string, mapping map[string]any) error {

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
		return rove.NewError("es.create_index", rove.CategoryIndex, true, "create index %s status: %s: %s", physical, resp.Status(), readErrorBody(resp))
	}
	return nil
}

// Reindex 将 source 物理索引全量复制到 dest（完成后显式 refresh，保证计数立即可见）。
func (c *Client) Reindex(ctx context.Context, source, dest string) error {

	body := fmt.Sprintf(`{"source":{"index":%q},"dest":{"index":%q}}`, source, dest)
	resp, err := c.es.Reindex(strings.NewReader(body), c.es.Reindex.WithContext(ctx), c.es.Reindex.WithWaitForCompletion(true))
	if err != nil {
		return rove.NewError("es.reindex", rove.CategoryIndex, true, "reindex %s -> %s: %v", source, dest, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return rove.NewError("es.reindex", rove.CategoryIndex, true, "reindex %s -> %s status: %s: %s", source, dest, resp.Status(), readErrorBody(resp))
	}
	refreshResp, refreshErr := c.es.Indices.Refresh(c.es.Indices.Refresh.WithContext(ctx), c.es.Indices.Refresh.WithIndex(dest))
	if refreshErr != nil {
		return rove.NewError("es.reindex", rove.CategoryIndex, true, "refresh %s: %v", dest, refreshErr)
	}
	defer refreshResp.Body.Close()
	if refreshResp.IsError() {
		return rove.NewError("es.reindex", rove.CategoryIndex, true, "refresh %s status: %s", dest, refreshResp.Status())
	}
	return nil
}

// SwitchAlias 原子切换 alias：一次性 remove 旧物理索引并 add 新物理索引。
func (c *Client) SwitchAlias(ctx context.Context, alias, oldPhysical, newPhysical string) error {

	body := fmt.Sprintf(`{"actions":[{"remove":{"index":%q,"alias":%q}},{"add":{"index":%q,"alias":%q}}]}`,
		oldPhysical, alias, newPhysical, alias)
	resp, err := c.es.Indices.UpdateAliases(strings.NewReader(body), c.es.Indices.UpdateAliases.WithContext(ctx))
	if err != nil {
		return rove.NewError("es.update_aliases", rove.CategoryIndex, true, "switch alias %s: %v", alias, err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return rove.NewError("es.update_aliases", rove.CategoryIndex, true, "switch alias %s status: %s: %s", alias, resp.Status(), readErrorBody(resp))
	}
	return nil
}

// DeletePhysical 直接删除物理索引（404 视为成功）。
func (c *Client) DeletePhysical(ctx context.Context, physical string) error {

	resp, err := c.es.Indices.Delete([]string{physical}, c.es.Indices.Delete.WithContext(ctx))
	if err != nil {
		return rove.NewError("es.delete_index", rove.CategoryIndex, true, "delete index %s: %v", physical, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil
	}
	if resp.IsError() {
		return rove.NewError("es.delete_index", rove.CategoryIndex, true, "delete index %s status: %s: %s", physical, resp.Status(), readErrorBody(resp))
	}
	return nil
}

// ScanAllChunks 分页扫描 chunk 物理索引全部文档，visit 依次接收 _id 与 _source。
func (c *Client) ScanAllChunks(ctx context.Context, index string, size int, visit func(id string, source map[string]any) error) error {

	if size <= 0 {
		size = scanDefaultSize
	}
	from := 0
	for {
		body, err := json.Marshal(map[string]any{
			"size":  size,
			"from":  from,
			"query": map[string]any{"match_all": map[string]any{}},
		})
		if err != nil {
			return rove.NewError("es.scan", rove.CategoryIndex, true, "marshal scan body: %v", err)
		}
		resp, err := c.es.Search(
			c.es.Search.WithContext(ctx),
			c.es.Search.WithIndex(index),
			c.es.Search.WithBody(bytes.NewReader(body)),
		)
		if err != nil {
			return rove.NewError("es.scan", rove.CategoryIndex, true, "scan %s: %v", index, err)
		}
		if resp.IsError() {
			defer resp.Body.Close()
			return rove.NewError("es.scan", rove.CategoryIndex, true, "scan %s status: %s: %s", index, resp.Status(), readErrorBody(resp))
		}
		var payload struct {
			Hits struct {
				Total struct {
					Value int64 `json:"value"`
				} `json:"total"`
				Hits []struct {
					ID     string         `json:"_id"`
					Source map[string]any `json:"_source"`
				} `json:"hits"`
			} `json:"hits"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&payload)
		resp.Body.Close()
		if decodeErr != nil {
			return rove.NewError("es.scan", rove.CategoryIndex, true, "decode scan: %v", decodeErr)
		}
		for _, hit := range payload.Hits.Hits {
			if err := visit(hit.ID, hit.Source); err != nil {
				return err
			}
		}
		from += len(payload.Hits.Hits)
		if from == 0 || int64(from) >= payload.Hits.Total.Value {
			return nil
		}
	}
}
