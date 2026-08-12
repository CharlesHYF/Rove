/*
 * 文件作用：pkg/ranking heuristic 排序的单元测试 -- 新鲜度衰减、语言匹配与特征分解（规格书 §5.4）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package ranking

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
)

// fixtureDoc 构造指定年龄与语言的文档。
func fixtureDoc(age time.Duration, lang string) *document.Document {

	return &document.Document{
		ID:        "d1",
		Title:     "t",
		Language:  lang,
		FetchedAt: time.Now().Add(-age),
		Source:    document.Source{Domain: "example.com", Type: "web"},
	}
}

func TestScoreFreshness(t *testing.T) {

	features := Features{Now: time.Now()}
	freshScore, _ := features.Score(10.0, fixtureDoc(time.Hour, "en"), "en")
	staleScore, _ := features.Score(10.0, fixtureDoc(400*24*time.Hour, "en"), "en")
	require.Greater(t, freshScore, staleScore, "fresh doc must outrank stale doc")
}

func TestScoreLanguageMatch(t *testing.T) {

	features := Features{Now: time.Now()}
	matchScore, _ := features.Score(10.0, fixtureDoc(time.Hour, "zh"), "zh")
	mismatchScore, _ := features.Score(10.0, fixtureDoc(time.Hour, "en"), "zh")
	require.Greater(t, matchScore, mismatchScore, "language match must add score")
}

func TestScoreFeaturesMap(t *testing.T) {

	features := Features{DomainQuality: map[string]float64{"example.com": 2.0}, Now: time.Now()}
	_, featureMap := features.Score(10.0, fixtureDoc(time.Hour, "en"), "en")
	require.Contains(t, featureMap, "freshness")
	require.Contains(t, featureMap, "domain_quality")
	require.Contains(t, featureMap, "language_match")
	require.Contains(t, featureMap, "dup_penalty")
	require.Equal(t, float64(2.0), featureMap["domain_quality"], "configured domain quality must apply")
	require.Equal(t, 0.0, featureMap["dup_penalty"], "M2 不实现近重复惩罚")
}
