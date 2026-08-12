/*
 * 文件作用：ES 底层操作 -- 单文档删除、delete_by_query、bulk、集群健康、计数与存储大小（供 pkg/index 使用）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	esapi "github.com/elastic/go-elasticsearch/v8/esapi"

	"rove/pkg/rove"
)

// Bulk 提交 NDJSON bulk 请求（refresh=wait_for）。
func (c *Client) Bulk(ctx context.Context, payload []byte) (*Response, error) {

	resp, err := c.es.Bulk(
		bytes.NewReader(payload),
		c.es.Bulk.WithContext(ctx),
		c.es.Bulk.WithRefresh("wait_for"),
	)
	if err != nil {
		return nil, rove.NewError("index.bulk", rove.CategoryIndex, true, "bulk request: %v", err)
	}
	return newResponse(resp), nil
}

// DeleteIndexDoc 删除单文档（404 视为成功）。
func (c *Client) DeleteIndexDoc(ctx context.Context, alias, documentID string) (*Response, error) {

	resp, err := c.es.Delete(alias, documentID, c.es.Delete.WithContext(ctx), c.es.Delete.WithRefresh("wait_for"))
	if err != nil {
		return nil, rove.NewError("index.delete", rove.CategoryIndex, true, "delete %s/%s: %v", alias, documentID, err)
	}
	if resp.StatusCode == 404 {
		return newResponse(resp), nil
	}
	if resp.IsError() {
		return nil, rove.NewError("index.delete", rove.CategoryIndex, true, "delete %s/%s status: %s", alias, documentID, resp.Status())
	}
	return newResponse(resp), nil
}

// DeleteByQuery 按查询删除（refresh 后立即可见）。
func (c *Client) DeleteByQuery(ctx context.Context, alias, queryBody string) (*Response, error) {

	resp, err := c.es.DeleteByQuery([]string{alias}, strings.NewReader(queryBody),
		c.es.DeleteByQuery.WithContext(ctx), c.es.DeleteByQuery.WithRefresh(true))
	if err != nil {
		return nil, rove.NewError("index.delete_by_query", rove.CategoryIndex, true, "delete_by_query %s: %v", alias, err)
	}
	if resp.IsError() {
		return nil, rove.NewError("index.delete_by_query", rove.CategoryIndex, true, "delete_by_query %s status: %s", alias, resp.Status())
	}
	return newResponse(resp), nil
}

// ClusterHealth 返回集群健康状态（green/yellow/red）。
func (c *Client) ClusterHealth(ctx context.Context) (string, error) {

	resp, err := c.es.Cluster.Health(c.es.Cluster.Health.WithContext(ctx))
	if err != nil {
		return "", rove.NewError("es.health", rove.CategoryIndex, true, "cluster health: %v", err)
	}
	defer resp.Body.Close()
	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", rove.NewError("es.health", rove.CategoryIndex, true, "decode health: %v", err)
	}
	return payload.Status, nil
}

// Count 返回索引（alias）文档数。
func (c *Client) Count(ctx context.Context, alias string) (int64, error) {

	resp, err := c.es.Count(c.es.Count.WithContext(ctx), c.es.Count.WithIndex(alias))
	if err != nil {
		return 0, rove.NewError("es.count", rove.CategoryIndex, true, "count %s: %v", alias, err)
	}
	defer resp.Body.Close()
	var payload struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, rove.NewError("es.count", rove.CategoryIndex, true, "decode count: %v", err)
	}
	return payload.Count, nil
}

// StoreSize 返回索引（alias）总存储字节数。
func (c *Client) StoreSize(ctx context.Context, alias string) (int64, error) {

	resp, err := c.es.Indices.Stats(c.es.Indices.Stats.WithContext(ctx), c.es.Indices.Stats.WithIndex(alias))
	if err != nil {
		return 0, rove.NewError("es.stats", rove.CategoryIndex, true, "stats %s: %v", alias, err)
	}
	defer resp.Body.Close()
	var payload struct {
		Indices map[string]struct {
			Total struct {
				Store struct {
					SizeInBytes int64 `json:"size_in_bytes"`
				} `json:"store"`
			} `json:"total"`
		} `json:"indices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, rove.NewError("es.stats", rove.CategoryIndex, true, "decode stats: %v", err)
	}
	var total int64
	for _, info := range payload.Indices {
		total += info.Total.Store.SizeInBytes
	}
	return total, nil
}

// Response 是 ES 响应的精简封装。
type Response struct {
	StatusCode int
	Body       io.ReadCloser
}

// IsError 判断响应是否为错误状态（>= 400）。
func (r *Response) IsError() bool {

	return r.StatusCode >= 400
}

// newResponse 包装 go-elasticsearch 响应。
func newResponse(resp *esapi.Response) *Response {

	return &Response{StatusCode: resp.StatusCode, Body: resp.Body}
}
