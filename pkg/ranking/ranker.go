/*
 * 文件作用：heuristic Ranker -- 在 BM25 分数上叠加 freshness / domain_quality / language 特征（规格书 §5.4 默认权重）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package ranking 提供无模型的启发式排序特征。
package ranking

import (
	"math"
	"time"

	"rove/pkg/document"
)

const (
	weightFreshness     = 0.1
	weightDomainQuality = 0.1
	weightLanguage      = 0.05
	dupPenalty          = -0.2
	freshnessHalfLife   = 90 * 24 * time.Hour
)

// Features 携带排序所需的上下文。
type Features struct {
	DomainQuality map[string]float64 // domain -> quality 分数，未配置默认 1.0
	Now           time.Time
}

// Score 计算最终分数与特征分解。base 为 BM25 原始分数。
func (f Features) Score(base float64, doc *document.Document, queryLanguage string) (float64, map[string]float64) {

	freshness := freshnessScore(doc, f.Now)
	quality := f.domainQuality(doc.Source.Domain)
	language := languageMatch(queryLanguage, doc.Language)

	finalScore := base + freshness*weightFreshness + (quality-1.0)*weightDomainQuality + language*weightLanguage
	features := map[string]float64{
		"freshness":      freshness,
		"domain_quality": quality,
		"language_match": language,
		"dup_penalty":    0, // M2 不实现近重复惩罚（M3 起）
	}
	_ = dupPenalty
	return finalScore, features
}

// freshnessScore 指数衰减：exp(-age/halfLife)，0..1。
func freshnessScore(doc *document.Document, now time.Time) float64 {

	age := now.Sub(doc.FetchedAt)
	if doc.PublishedAt != nil {
		age = now.Sub(*doc.PublishedAt)
	}
	if age <= 0 {
		return 1.0
	}
	return math.Exp(-age.Seconds() / freshnessHalfLife.Seconds())
}

// domainQuality 查询域名质量分，未配置返回 1.0。
func (f Features) domainQuality(domain string) float64 {

	if quality, ok := f.DomainQuality[domain]; ok {
		return quality
	}
	return 1.0
}

// languageMatch 查询语言与文档语言匹配得 1，不匹配或未知得 0。
func languageMatch(queryLanguage, docLanguage string) float64 {

	if queryLanguage == "" || docLanguage == "" || queryLanguage == docLanguage {
		return 1.0
	}
	return 0.0
}
