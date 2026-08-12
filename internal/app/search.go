/*
 * 文件作用：Application Service -- 统一 Search / Index 能力，供 CLI/TUI 使用（规格书 §11 分层）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package app 提供跨界面复用的应用服务。
package app

import (
	"context"

	"rove/pkg/document"
	"rove/pkg/index"
	"rove/pkg/retrieval"
)

// SearchService 检索服务。
type SearchService struct {
	retriever *retrieval.Retriever
}

// NewSearch 构造检索服务。
func NewSearch(retriever *retrieval.Retriever) *SearchService {

	return &SearchService{retriever: retriever}
}

// Search 执行检索。
func (s *SearchService) Search(ctx context.Context, q *retrieval.Query) (*retrieval.SearchResult, error) {

	return s.retriever.Search(ctx, q)
}

// IndexService 索引服务。
type IndexService struct {
	indexer *index.Indexer
}

// NewIndex 构造索引服务。
func NewIndex(indexer *index.Indexer) *IndexService {

	return &IndexService{indexer: indexer}
}

// Index 幂等写入单个文档。
func (s *IndexService) Index(ctx context.Context, doc *document.Document) error {

	return s.indexer.Index(ctx, doc)
}

// IndexBulk 批量写入。
func (s *IndexService) IndexBulk(ctx context.Context, docs []*document.Document) error {

	return s.indexer.IndexBulk(ctx, docs)
}

// Delete 删除文档。
func (s *IndexService) Delete(ctx context.Context, documentID string) error {

	return s.indexer.Delete(ctx, documentID)
}

// Status 索引状态。
func (s *IndexService) Status(ctx context.Context) (*index.Status, error) {

	return s.indexer.Status(ctx)
}
