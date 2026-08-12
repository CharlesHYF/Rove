/*
 * 文件作用：幂等索引器 -- 将标准 Document 写入 ES（先清旧 chunks 再 bulk 写入），并支持删除与状态查询（NFR-007）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package index 负责 Document/Chunk 到 Elasticsearch 的幂等写入。
package index

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rove/pkg/document"
	"rove/pkg/elastic"
	"rove/pkg/rove"
)

const (
	maxBulkRetries = 1
	retryBackoff   = 2 * time.Second
)

// Indexer 封装 ES 写入操作。
type Indexer struct {
	es       *elastic.Client
	embedder Embedder // 可选：索引时填充 chunk 向量
}

// Status 描述索引规模与健康。
type Status struct {
	Health    string
	Documents int64
	Chunks    int64
	SizeBytes int64
}

// New 构造 Indexer；embedders 可传 0 或 1 个 Embedder（无则跳过向量填充）。
func New(es *elastic.Client, embedders ...Embedder) *Indexer {

	indexer := &Indexer{es: es}
	if len(embedders) > 0 {
		indexer.embedder = embedders[0]
	}
	return indexer
}

// Index 幂等写入单个文档：先清旧 chunks，填充向量（若有 embedder），再 bulk 写入。
func (i *Indexer) Index(ctx context.Context, doc *document.Document) error {

	if err := i.deleteChunks(ctx, doc.ID); err != nil {
		return err
	}
	if err := i.embedChunks(ctx, doc); err != nil {
		return err
	}
	return i.bulkWrite(ctx, doc)
}

// embedChunks 为尚无向量的 chunk 计算嵌入（embedder 未配置时跳过）。
func (i *Indexer) embedChunks(ctx context.Context, doc *document.Document) error {

	if i.embedder == nil {
		return nil
	}
	var pending []*document.Chunk
	var pendingTexts []string
	for index := range doc.Chunks {
		chunk := &doc.Chunks[index]
		if chunk.Embedding == nil {
			pending = append(pending, chunk)
			pendingTexts = append(pendingTexts, chunk.Content)
		}
	}
	if len(pending) == 0 {
		return nil
	}
	vectors, err := i.embedder.Embed(ctx, pendingTexts)
	if err != nil {
		return rove.NewError("index.embedding", rove.CategoryEmbedding, true, "embed chunks: %v", err)
	}
	if len(vectors) != len(pending) {
		return rove.NewError("index.embedding", rove.CategoryEmbedding, true, "embedder returned %d vectors for %d texts", len(vectors), len(pending))
	}
	for index := range pending {
		pending[index].Embedding = vectors[index]
	}
	return nil
}

// IndexBulk 批量写入多个文档（逐个 Index，保证每个文档内部幂等）。
func (i *Indexer) IndexBulk(ctx context.Context, docs []*document.Document) error {

	for _, doc := range docs {
		if err := i.Index(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// Delete 删除文档及其全部 chunks。
func (i *Indexer) Delete(ctx context.Context, documentID string) error {

	if _, err := i.es.DeleteIndexDoc(ctx, i.es.DocumentsAlias(), documentID); err != nil {
		return err
	}
	return i.deleteChunks(ctx, documentID)
}

// Status 汇总索引规模与集群健康。
func (i *Indexer) Status(ctx context.Context) (*Status, error) {

	health, err := i.es.ClusterHealth(ctx)
	if err != nil {
		return nil, err
	}
	documents, err := i.es.Count(ctx, i.es.DocumentsAlias())
	if err != nil {
		return nil, err
	}
	chunks, err := i.es.Count(ctx, i.es.ChunksAlias())
	if err != nil {
		return nil, err
	}
	sizeBytes, err := i.es.StoreSize(ctx, i.es.DocumentsAlias())
	if err != nil {
		return nil, err
	}
	return &Status{Health: health, Documents: documents, Chunks: chunks, SizeBytes: sizeBytes}, nil
}

// deleteChunks 删除指定文档的全部 chunk（refresh 确保后续写入可覆盖）。
func (i *Indexer) deleteChunks(ctx context.Context, documentID string) error {

	body := fmt.Sprintf(`{"query":{"term":{"document_id":%q}}}`, documentID)
	_, err := i.es.DeleteByQuery(ctx, i.es.ChunksAlias(), body)
	return err
}

// bulkWrite 组装并提交 bulk 请求，失败重试一次。
func (i *Indexer) bulkWrite(ctx context.Context, doc *document.Document) error {

	payload := i.buildBulkPayload(doc)
	var lastErr error
	for attempt := 0; attempt <= maxBulkRetries; attempt++ {
		resp, err := i.es.Bulk(ctx, payload)
		if err != nil {
			lastErr = err
			time.Sleep(retryBackoff)
			continue
		}
		if resp.IsError() {
			lastErr = rove.NewError("index.bulk", rove.CategoryIndex, true, "bulk status: %d", resp.StatusCode)
			time.Sleep(retryBackoff)
			continue
		}
		return nil
	}
	return lastErr
}

// buildBulkPayload 生成 NDJSON bulk 请求体。
func (i *Indexer) buildBulkPayload(doc *document.Document) []byte {

	var buf bytes.Buffer
	writeLine := func(action map[string]any, source map[string]any) {
		actionJSON, _ := json.Marshal(action)
		buf.Write(actionJSON)
		buf.WriteByte('\n')
		if source != nil {
			sourceJSON, _ := json.Marshal(source)
			buf.Write(sourceJSON)
			buf.WriteByte('\n')
		}
	}

	writeLine(
		map[string]any{"index": map[string]any{"_index": i.es.DocumentsAlias(), "_id": doc.ID}},
		documentSource(doc),
	)
	for _, chunk := range doc.Chunks {
		writeLine(
			map[string]any{"index": map[string]any{"_index": i.es.ChunksAlias(), "_id": chunk.ChunkID}},
			chunkSource(doc, chunk),
		)
	}
	return buf.Bytes()
}

// documentSource 序列化文档索引字段。
func documentSource(doc *document.Document) map[string]any {

	source := map[string]any{
		"id":            doc.ID,
		"url":           doc.URL,
		"canonical_url": doc.CanonicalURL,
		"title":         doc.Title,
		"description":   doc.Description,
		"content":       doc.Content,
		"language":      doc.Language,
		"fetched_at":    doc.FetchedAt.UTC().Format(time.RFC3339),
		"source": map[string]any{
			"domain": doc.Source.Domain,
			"type":   doc.Source.Type,
		},
		"content_hash": doc.ContentHash,
	}
	if doc.PublishedAt != nil {
		source["published_at"] = doc.PublishedAt.UTC().Format(time.RFC3339)
	}
	if len(doc.Metadata) > 0 {
		source["metadata"] = doc.Metadata
	}
	return source
}

// chunkSource 序列化 chunk 索引字段。
func chunkSource(doc *document.Document, chunk document.Chunk) map[string]any {

	source := map[string]any{
		"chunk_id":     chunk.ChunkID,
		"document_id":  chunk.DocumentID,
		"title":        doc.Title,
		"content":      chunk.Content,
		"heading_path": chunk.HeadingPath,
		"position":     chunk.Position,
		"language":     chunk.Language,
		"domain":       doc.Source.Domain,
		"token_count":  chunk.TokenCount,
	}
	if doc.PublishedAt != nil {
		source["published_at"] = doc.PublishedAt.UTC().Format(time.RFC3339)
	}
	if chunk.Embedding != nil {
		source["embedding"] = chunk.Embedding
	}
	return source
}
