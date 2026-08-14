/*
 * 文件作用：索引迁移器 -- create vNext -> reindex/重嵌入 -> 计数校验 -> alias 原子切换（规格书 §5.3）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package index

import (
	"context"
	"encoding/json"
	"strings"

	"rove/pkg/elastic"
	"rove/pkg/retrieval"
	"rove/pkg/rove"
)

// 迁移批量常量。
const migrateEmbedBatchSize = 64

// MigrateResult 迁移结果。
type MigrateResult struct {
	DocumentsOld   string
	DocumentsNew   string
	ChunksOld      string
	ChunksNew      string
	DocumentsMoved int64
	ChunksMoved    int64
}

// Migrator 索引迁移器：documents 走 ES reindex；chunks 因 embedding 不入 _source，需扫描重嵌入回填。
type Migrator struct {
	client   *elastic.Client
	embedder retrieval.Embedder
}

// NewMigrator 构造迁移器。
func NewMigrator(client *elastic.Client, embedder retrieval.Embedder) *Migrator {

	return &Migrator{client: client, embedder: embedder}
}

// Migrate 执行迁移：旧索引不存在返回错误（先 rove index init）。
func (m *Migrator) Migrate(ctx context.Context, embeddingDim int, deleteOld bool) (*MigrateResult, error) {

	documentsAlias := m.client.DocumentsAlias()
	chunksAlias := m.client.ChunksAlias()

	documentsOld, err := m.client.CurrentIndex(ctx, documentsAlias)
	if err != nil {
		return nil, err
	}
	if documentsOld == "" {
		return nil, rove.NewError("index.migrate", rove.CategoryIndex, false, "documents index not found, run rove index init first")
	}
	chunksOld, err := m.client.CurrentIndex(ctx, chunksAlias)
	if err != nil {
		return nil, err
	}
	if chunksOld == "" {
		return nil, rove.NewError("index.migrate", rove.CategoryIndex, false, "chunks index not found, run rove index init first")
	}

	documentsNew, err := elastic.NextVersionName(documentsAlias, documentsOld)
	if err != nil {
		return nil, err
	}
	chunksNew, err := elastic.NextVersionName(chunksAlias, chunksOld)
	if err != nil {
		return nil, err
	}

	if err := m.client.CreateDocumentsIndex(ctx, documentsNew); err != nil {
		return nil, err
	}
	if err := m.client.CreateChunksIndex(ctx, chunksNew, embeddingDim); err != nil {
		return nil, err
	}

	if err := m.client.Reindex(ctx, documentsOld, documentsNew); err != nil {
		return nil, err
	}
	chunksMoved, err := m.migrateChunks(ctx, chunksOld, chunksNew)
	if err != nil {
		return nil, err
	}

	// 校验：新旧计数一致才允许切换 alias（失败保持旧 alias 不动）。
	documentsMoved, err := m.client.Count(ctx, documentsNew)
	if err != nil {
		return nil, err
	}
	documentsOldCount, err := m.client.Count(ctx, documentsOld)
	if err != nil {
		return nil, err
	}
	if documentsMoved != documentsOldCount {
		return nil, rove.NewError("index.migrate", rove.CategoryIndex, false, "documents count mismatch: old=%d new=%d", documentsOldCount, documentsMoved)
	}
	chunksOldCount, err := m.client.Count(ctx, chunksOld)
	if err != nil {
		return nil, err
	}
	if chunksMoved != chunksOldCount {
		return nil, rove.NewError("index.migrate", rove.CategoryIndex, false, "chunks count mismatch: old=%d new=%d", chunksOldCount, chunksMoved)
	}

	if err := m.client.SwitchAlias(ctx, documentsAlias, documentsOld, documentsNew); err != nil {
		return nil, err
	}
	if err := m.client.SwitchAlias(ctx, chunksAlias, chunksOld, chunksNew); err != nil {
		return nil, err
	}

	if deleteOld {
		if err := m.client.DeletePhysical(ctx, documentsOld); err != nil {
			return nil, err
		}
		if err := m.client.DeletePhysical(ctx, chunksOld); err != nil {
			return nil, err
		}
	}

	return &MigrateResult{
		DocumentsOld:   documentsOld,
		DocumentsNew:   documentsNew,
		ChunksOld:      chunksOld,
		ChunksNew:      chunksNew,
		DocumentsMoved: documentsMoved,
		ChunksMoved:    chunksMoved,
	}, nil
}

// chunkPending 待写入新索引的 chunk。
type chunkPending struct {
	id         string
	documentID string
	content    string
	source     map[string]any
}

// migrateChunks 扫描旧 chunk 索引，重嵌入向量并批量写入新索引（回填 source_type，供垂类路由使用）。
func (m *Migrator) migrateChunks(ctx context.Context, oldIndex, newIndex string) (int64, error) {

	var moved int64
	var batch []chunkPending
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		texts := make([]string, len(batch))
		documentIDs := make([]string, len(batch))
		for index := range batch {
			texts[index] = batch[index].content
			documentIDs[index] = batch[index].documentID
		}
		embeddings, err := m.embedder.Embed(ctx, texts)
		if err != nil {
			return err
		}
		if len(embeddings) != len(batch) {
			return rove.NewError("index.migrate", rove.CategoryIndex, false, "embedder returned %d vectors for %d texts", len(embeddings), len(batch))
		}
		// 迁移时 alias 仍指向旧 documents 索引，按 id 回填 source_type。
		documents, err := m.client.MGetDocuments(ctx, m.client.DocumentsAlias(), uniqueStrings(documentIDs))
		if err != nil {
			return err
		}

		var payload strings.Builder
		for index := range batch {
			source := cloneSource(batch[index].source)
			source["embedding"] = embeddings[index]
			sourceType := "web"
			if doc, ok := documents[batch[index].documentID]; ok && doc.Source.Type != "" {
				sourceType = doc.Source.Type
			}
			source["source_type"] = sourceType
			action, _ := json.Marshal(map[string]any{"index": map[string]any{"_index": newIndex, "_id": batch[index].id}})
			body, _ := json.Marshal(source)
			payload.Write(action)
			payload.WriteByte('\n')
			payload.Write(body)
			payload.WriteByte('\n')
		}
		resp, err := m.client.Bulk(ctx, []byte(payload.String()))
		if err != nil {
			return err
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		if resp.IsError() {
			return rove.NewError("index.migrate", rove.CategoryIndex, true, "bulk into %s status: %d", newIndex, resp.StatusCode)
		}
		moved += int64(len(batch))
		batch = nil
		return nil
	}

	err := m.client.ScanAllChunks(ctx, oldIndex, migrateEmbedBatchSize, func(id string, source map[string]any) error {
		chunkID := id
		if value, ok := source["chunk_id"].(string); ok && value != "" {
			chunkID = value
		}
		content, _ := source["content"].(string)
		documentID, _ := source["document_id"].(string)
		batch = append(batch, chunkPending{id: chunkID, documentID: documentID, content: content, source: source})
		if len(batch) >= migrateEmbedBatchSize {
			if flushErr := flush(); flushErr != nil {
				return flushErr
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	if err := flush(); err != nil {
		return 0, err
	}
	return moved, nil
}

// uniqueStrings 去重字符串列表（保持顺序）。
func uniqueStrings(values []string) []string {

	seen := map[string]bool{}
	var result []string
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

// cloneSource 浅拷贝 chunk 源字段映射。
func cloneSource(source map[string]any) map[string]any {

	clone := make(map[string]any, len(source)+2)
	for key, value := range source {
		clone[key] = value
	}
	return clone
}
