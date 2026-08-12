/*
 * 文件作用：Elasticsearch 客户端封装 -- 连接、Ping 健康检查与索引别名常量（规格书 §5.3）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package elastic 封装 Elasticsearch 8.x 客户端与索引生命周期管理。
package elastic

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"

	"rove/pkg/rove"
)

// Client 封装 ES 客户端，持有索引前缀。
type Client struct {
	es     *elasticsearch.Client
	prefix string
}

// New 构造 ES 客户端；prefix 为空时默认 "rove"。
func New(esURL, username, password, prefix string) (*Client, error) {

	if prefix == "" {
		prefix = "rove"
	}
	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
		Username:  username,
		Password:  password,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, rove.NewError("es.client", rove.CategoryIndex, true, "create es client: %v", err)
	}
	return &Client{es: es, prefix: prefix}, nil
}

// Ping 检查 ES 连通性。
func (c *Client) Ping(ctx context.Context) error {

	resp, err := c.es.Ping(c.es.Ping.WithContext(ctx))
	if err != nil {
		return rove.NewError("es.ping", rove.CategoryIndex, true, "es ping: %v", err)
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return rove.NewError("es.ping", rove.CategoryIndex, true, "es ping status: %s", resp.Status())
	}
	return nil
}

// DocumentsAlias 返回文档索引别名。
func (c *Client) DocumentsAlias() string {

	return fmt.Sprintf("%s-documents", c.prefix)
}

// ChunksAlias 返回 chunk 索引别名。
func (c *Client) ChunksAlias() string {

	return fmt.Sprintf("%s-chunks", c.prefix)
}
