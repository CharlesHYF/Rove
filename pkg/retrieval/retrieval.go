/*
 * 文件作用：检索 -- BM25 + 可选向量腿，RRF Fusion 后 document 聚合、回填与 heuristic 排序（规格书 §4.2 / §5.1）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package retrieval 实现 BM25/向量召回、RRF 融合与 document 级聚合。
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
	SourceType    string
	PublishedFrom *time.Time
	PublishedTo   *time.Time
}

// Query 检索请求。
type Query struct {
	Text          string
	Filters       Filters
	TopK          int
	IncludeVector bool     // 配置了 embedder 时默认 true（由 Retriever 决定）
	Vertical      Vertical // 垂类路由：auto 自动推断（默认），docs/code/academic 应用 source_type 过滤
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
	Query    string
	Vertical Vertical
	Hits     []SearchHit
	Timings  map[string]time.Duration
	TraceID  string
}

// Retriever 执行 BM25 + 可选向量检索。
type Retriever struct {
	es       *elastic.Client
	ranker   ranking.Features
	embedder Embedder
	Reranker Reranker // 可选重排序（nil 则跳过，历史行为不变）
}

// New 构造 Retriever；embedders 可传 0 或 1 个（无则不启用向量腿）。
func New(es *elastic.Client, embedders ...Embedder) *Retriever {

	retriever := &Retriever{
		es:     es,
		ranker: ranking.Features{Now: time.Now()},
	}
	if len(embedders) > 0 {
		retriever.embedder = embedders[0]
	}
	return retriever
}

// Search 执行 BM25（+ 向量）召回 -> RRF Fusion -> document 聚合 -> 排序。
func (r *Retriever) Search(ctx context.Context, q *Query) (*SearchResult, error) {

	timings := map[string]time.Duration{}

	// 垂类路由（确定性，无 LLM）：auto 推断；仅 docs/code/academic 应用 source_type 过滤，web 为免过滤兜底。
	processorStart := time.Now()
	vertical := q.Vertical
	if vertical == VerticalAuto {
		vertical = RouteVertical(q.Text)
	}
	effectiveQuery := *q
	if vertical != VerticalWeb {
		effectiveQuery.Filters.SourceType = string(vertical)
	}
	timings["query_processor"] = time.Since(processorStart)

	topK := q.TopK
	if topK <= 0 {
		topK = defaultTopK
	}
	candidateSize := topK * candidateMultiplier

	// 向量腿（可选）
	var legs []namedLeg
	includeVector := q.IncludeVector
	if !includeVector && r.embedder != nil {
		includeVector = true // 配置了 embedder 默认启用
	}
	if includeVector && r.embedder != nil {
		vectorStart := time.Now()
		vectors, err := r.embedder.Embed(ctx, []string{q.Text})
		if err != nil {
			return nil, err
		}
		knnHits, err := r.es.SearchChunksKNN(ctx, r.es.ChunksAlias(), vectors[0], candidateSize, topK*10)
		if err != nil {
			return nil, err
		}
		timings["vector_retrieve"] = time.Since(vectorStart)
		legs = append(legs, namedLeg{Name: "vector", Hits: knnHits})
	}

	// BM25 腿
	lexicalStart := time.Now()
	body, err := buildSearchBody(&effectiveQuery, candidateSize)
	if err != nil {
		return nil, err
	}
	lexicalHits, err := r.es.SearchChunks(ctx, r.es.ChunksAlias(), body)
	if err != nil {
		return nil, err
	}
	timings["lexical_retrieve"] = time.Since(lexicalStart)
	legs = append(legs, namedLeg{Name: "lexical", Hits: lexicalHits})

	// RRF Fusion
	fused := rrfMerge(legs, rankConstant)

	// 按 document_id 聚合（保留最高 fusion chunk）
	grouped := r.aggregateFused(fused)

	// mget 回填文档信息
	docs, err := r.es.MGetDocuments(ctx, r.es.DocumentsAlias(), grouped.ids())
	if err != nil {
		return nil, err
	}

	// 排序
	rankStart := time.Now()
	result := r.rankFused(q, grouped, docs, topK)
	timings["rank"] = time.Since(rankStart)

	// 可选重排序（PRD FR-RNK-002）
	if r.Reranker != nil {
		rerankStart := time.Now()
		result, err = r.Reranker.Rerank(ctx, q, result)
		if err != nil {
			return nil, err
		}
		timings["rerank"] = time.Since(rerankStart)
	}

	return &SearchResult{
		Query:    q.Text,
		Vertical: vertical,
		Hits:     result,
		Timings:  timings,
		TraceID:  newTraceID(),
	}, nil
}

// fusedGrouped 按 document 聚合后的候选。
type fusedGrouped struct {
	byDoc map[string]fusedChunk // document_id -> 最高 fusion chunk
}

// ids 返回全部 document_id。
func (g *fusedGrouped) ids() []string {

	ids := make([]string, 0, len(g.byDoc))
	for id := range g.byDoc {
		ids = append(ids, id)
	}
	return ids
}

// aggregateFused 按 document_id 聚合，保留每文档最高 fusion chunk。
func (r *Retriever) aggregateFused(fused []fusedChunk) *fusedGrouped {

	grouped := &fusedGrouped{byDoc: map[string]fusedChunk{}}
	for _, fc := range fused {
		existing, ok := grouped.byDoc[fc.Chunk.DocumentID]
		if !ok || fc.Fusion > existing.Fusion {
			grouped.byDoc[fc.Chunk.DocumentID] = fc
		}
	}
	return grouped
}

// rankFused 回填文档信息并应用 ranker（base = fusion 分）。
func (r *Retriever) rankFused(q *Query, grouped *fusedGrouped, docs map[string]*document.Document, topK int) []SearchHit {

	var result []SearchHit
	for documentID, fc := range grouped.byDoc {
		doc := docs[documentID]
		if doc == nil {
			continue
		}
		score, features := r.ranker.Score(fc.Fusion, doc, q.Filters.Language)
		scores := map[string]float64{"fusion": fc.Fusion, "rank": score}
		for leg, legScore := range fc.LegScores {
			scores[leg] = legScore
		}
		for feature, value := range features {
			scores[feature] = value
		}
		result = append(result, SearchHit{
			Chunk:    fc.Chunk,
			Document: doc,
			Score:    score,
			Scores:   scores,
		})
	}
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
