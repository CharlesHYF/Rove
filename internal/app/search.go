/*
 * 文件作用：Application Service -- 统一 Search / Index 能力，供 CLI/TUI 使用（规格书 §11 分层）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package app 提供跨界面复用的应用服务。
package app

import (
	"context"

	"rove/internal/runtime"
	"rove/pkg/document"
	"rove/pkg/evidence"
	"rove/pkg/index"
	"rove/pkg/retrieval"
)

// SearchService 检索服务。
type SearchService struct {
	retriever *retrieval.Retriever
	bus       *runtime.Bus
}

// NewSearch 构造检索服务；bus 可传 nil（不发布事件）。
func NewSearch(retriever *retrieval.Retriever, buses ...*runtime.Bus) *SearchService {

	service := &SearchService{retriever: retriever}
	if len(buses) > 0 {
		service.bus = buses[0]
	}
	return service
}

// Search 执行检索并发布 telemetry 事件。
func (s *SearchService) Search(ctx context.Context, q *retrieval.Query) (*retrieval.SearchResult, error) {

	result, err := s.retriever.Search(ctx, q)
	if s.bus != nil {
		fields := map[string]string{"hits": "0"}
		if result != nil {
			fields["hits"] = itoa(len(result.Hits))
			for key, value := range result.Timings {
				fields[key] = itoa(int(value.Milliseconds()))
			}
		}
		message := "query=" + q.Text
		s.bus.Publish(runtime.Event{Type: runtime.Search, Message: message, Fields: fields})
	}
	return result, err
}

// itoa 简易整数转字符串（避免 strconv 依赖）。
func itoa(value int) string {

	if value == 0 {
		return "0"
	}
	var digits []byte
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

// Evidence 将检索结果转为证据列表。
func (s *SearchService) Evidence(ctx context.Context, result *retrieval.SearchResult) ([]evidence.Evidence, error) {

	return evidence.New().Build(ctx, result)
}

// IndexService 索引服务。
type IndexService struct {
	indexer *index.Indexer
	bus     *runtime.Bus
}

// NewIndex 构造索引服务；bus 可传 nil。
func NewIndex(indexer *index.Indexer, buses ...*runtime.Bus) *IndexService {

	service := &IndexService{indexer: indexer}
	if len(buses) > 0 {
		service.bus = buses[0]
	}
	return service
}

// Index 幂等写入单个文档并发布事件。
func (s *IndexService) Index(ctx context.Context, doc *document.Document) error {

	err := s.indexer.Index(ctx, doc)
	if s.bus != nil {
		s.bus.Publish(runtime.Event{
			Type:    runtime.Index,
			Message: "indexed " + doc.ID,
			Fields:  map[string]string{"chunks": itoa(len(doc.Chunks))},
		})
	}
	return err
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
