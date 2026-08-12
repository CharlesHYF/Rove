/*
 * 文件作用：BM25 检索 -- chunk 召回 -> document 聚合（最高 chunk 分）-> mget 回填文档信息 -> ranker 排序（规格书 §4.2）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package retrieval 实现 BM25 召回与 document 级聚合。
package retrieval

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"time"

	"rove/pkg/document"
	"rove/pkg/elastic"
	"rove/pkg/ranking"
)

const defaultTopK = 10

// candidateMultiplier 候选 chunk 倍率（召回 -> 聚合裁剪）。
const candidateMultiplier = 5

// Filters 检索过滤条件。
type Filters struct {
	Domain        string
	Language      string
	PublishedFrom *time.Time
	PublishedTo   *time.Time
}

// Query 检索请求。
type Query struct {
	Text    string
	Filters Filters
	TopK    int
}

// SearchHit 一条 document 级结果。
type SearchHit struct {
	Chunk    document.Chunk
	Document *document.Document
	Score    float64
	Scores   map[string]float64
}

// SearchResult 检索输出。
type SearchResult struct {
	Query   string
	Hits    []SearchHit
	Timings map[string]time.Duration
	TraceID string
}

// Retriever 执行 BM25 检索。
type Retriever struct {
	es     *elastic.Client
	ranker ranking.Features
}

// New 构造 Retriever。
func New(es *elastic.Client) *Retriever {

	return &Retriever{
		es:     es,
		ranker: ranking.Features{Now: time.Now()},
	}
}

// Search 执行 BM25 召回 -> 聚合 -> 排序。
func (r *Retriever) Search(ctx context.Context, q *Query) (*SearchResult, error) {

	timings := map[string]time.Duration{}
	start := time.Now()

	topK := q.TopK
	if topK <= 0 {
		topK = defaultTopK
	}
	candidateSize := topK * candidateMultiplier

	body, err := buildSearchBody(q, candidateSize)
	if err != nil {
		return nil, err
	}
	hits, err := r.es.SearchChunks(ctx, r.es.ChunksAlias(), body)
	if err != nil {
		return nil, err
	}
	timings["lexical_retrieve"] = time.Since(start)

	// chunk 按 document_id 聚合，保留最高分
	grouped := r.aggregate(hits)

	// mget 回填文档信息
	docs, err := r.es.MGetDocuments(ctx, r.es.DocumentsAlias(), grouped.ids())
	if err != nil {
		return nil, err
	}

	// 排序
	rankStart := time.Now()
	result := r.rank(q, grouped, docs, topK)
	timings["rank"] = time.Since(rankStart)

	return &SearchResult{
		Query:   q.Text,
		Hits:    result,
		Timings: timings,
		TraceID: newTraceID(),
	}, nil
}

// chunkHit 是 ES chunk 命中。
type chunkHit struct {
	Score float64
	Chunk document.Chunk
}

// groupedHits 按 document 聚合后的候选。
type groupedHits struct {
	byDoc map[string]chunkHit // document_id -> 最高分 chunk
}

// ids 返回全部 document_id。
func (g *groupedHits) ids() []string {

	ids := make([]string, 0, len(g.byDoc))
	for id := range g.byDoc {
		ids = append(ids, id)
	}
	return ids
}

// aggregate 将 chunk 命中按 document_id 聚合，保留每文档最高分 chunk。
func (r *Retriever) aggregate(rawHits []elastic.ChunkHit) *groupedHits {

	grouped := &groupedHits{byDoc: map[string]chunkHit{}}
	for _, hit := range rawHits {
		existing, ok := grouped.byDoc[hit.Chunk.DocumentID]
		if !ok || hit.Score > existing.Score {
			grouped.byDoc[hit.Chunk.DocumentID] = chunkHit{Score: hit.Score, Chunk: hit.Chunk}
		}
	}
	return grouped
}

// rank 回填文档信息并应用 ranker。
func (r *Retriever) rank(q *Query, grouped *groupedHits, docs map[string]*document.Document, topK int) []SearchHit {

	var result []SearchHit
	for documentID, hit := range grouped.byDoc {
		doc := docs[documentID]
		if doc == nil {
			continue
		}
		score, features := r.ranker.Score(hit.Score, doc, q.Filters.Language)
		scores := map[string]float64{"lexical": hit.Score, "rank": score}
		for feature, value := range features {
			scores[feature] = value
		}
		result = append(result, SearchHit{
			Chunk:    hit.Chunk,
			Document: doc,
			Score:    score,
			Scores:   scores,
		})
	}
	// 按最终分数降序，截断 topK
	sort.Slice(result, func(i, j int) bool { return result[i].Score > result[j].Score })
	if len(result) > topK {
		result = result[:topK]
	}
	return result
}

// newTraceID 生成 16 字节 hex trace id。
func newTraceID() string {

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
