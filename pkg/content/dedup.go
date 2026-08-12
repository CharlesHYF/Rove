/*
 * 文件作用：三层去重实现 -- canonical / exact content hash / near duplicate（token 集合 Jaccard >= 0.8）；URL 层由 frontier 负责（M4）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"context"

	"rove/pkg/document"
)

// nearDupJaccardThreshold 是近重复判定的 Jaccard 阈值：两文档共享 >= 80% token 视为近重复。
const nearDupJaccardThreshold = 0.8

// Deduper 实现三层去重判定并登记已见文档。
type Deduper struct {
	canonical map[string]string            // canonicalURL -> documentID
	exact     map[string]string            // contentHash -> documentID
	seen      map[string]map[string]struct{} // documentID -> tokenSet
}

// NewDeduper 创建空去重器。
func NewDeduper() *Deduper {

	return &Deduper{
		canonical: map[string]string{},
		exact:     map[string]string{},
		seen:      map[string]map[string]struct{}{},
	}
}

// IsDuplicate 依次检查三层；非重复文档登记后返回 (false, "")。
func (d *Deduper) IsDuplicate(ctx context.Context, doc *document.Document) (bool, string, error) {

	if _, ok := d.canonical[doc.CanonicalURL]; ok {
		return true, "canonical_duplicate", nil
	}
	if _, ok := d.exact[doc.ContentHash]; ok {
		return true, "exact_content_duplicate", nil
	}

	tokens := tokenSet(doc.Content)
	for _, existing := range d.seen {
		if jaccard(tokens, existing) >= nearDupJaccardThreshold {
			return true, "near_duplicate", nil
		}
	}

	d.canonical[doc.CanonicalURL] = doc.ID
	d.exact[doc.ContentHash] = doc.ID
	d.seen[doc.ID] = tokens
	return false, "", nil
}
