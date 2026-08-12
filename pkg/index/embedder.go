/*
 * 文件作用：索引侧最小 Embedder 接口 -- 解耦 index 与具体向量实现（依赖倒置，retrieval.Embedder 自动满足）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package index

import "context"

// Embedder 为 chunk 提供向量（索引写入前填充）。
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}
