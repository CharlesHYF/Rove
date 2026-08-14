/*
 * 文件作用：Application Service -- Map 链接发现服务，仅发现 URL 不做完整抓取与索引（PRD FR-MAP-001）。
 * 创建日期：2026-08-15
 * 修改日期：2026-08-15
 */
package app

import (
	"context"
	"fmt"
	neturl "net/url"

	"rove/internal/config"
	"rove/pkg/content"
	"rove/pkg/fetch"
)

// MapResult 链接发现结果。
type MapResult struct {
	URL       string   `json:"url"`
	Canonical string   `json:"canonical"`
	InSite    []string `json:"in_site"`
	OutSite   []string `json:"out_site"`
}

// MapService 链接发现服务。
type MapService struct {
	fetchService *FetchService
}

// NewMap 按配置构造链接发现服务。
func NewMap(cfg *config.Config) *MapService {

	return &MapService{fetchService: NewFetch(cfg)}
}

// Map 抓取单页并发现链接（HTTP First + auto 浏览器升级，与 fetch 同一链路）。
func (s *MapService) Map(ctx context.Context, url string) (*MapResult, error) {

	raw, err := s.fetchService.fetchRaw(ctx, url, fetch.ModeAuto)
	if err != nil {
		return nil, err
	}
	parser, err := content.DefaultRegistry().For(raw.ContentType)
	if err != nil {
		return nil, err
	}
	parsed, err := parser.Parse(ctx, raw)
	if err != nil {
		return nil, err
	}

	result := &MapResult{
		URL:       raw.URL,
		Canonical: parsed.CanonicalURL,
		InSite:    []string{},
		OutSite:   []string{},
	}
	if result.Canonical == "" {
		result.Canonical = raw.URL
	}
	base, err := neturl.Parse(raw.URL)
	if err != nil {
		return nil, fmt.Errorf("parse base url %s: %w", raw.URL, err)
	}
	baseHost := content.Hostname(raw.URL)
	seenInSite := map[string]bool{}
	seenOutSite := map[string]bool{}
	for _, link := range parsed.Links {
		if link == "" {
			continue
		}
		// Parser 输出原始 href，可能是相对路径，先解析为绝对地址。
		resolved, resolveErr := base.Parse(link)
		if resolveErr != nil {
			continue
		}
		if resolved.Scheme != "http" && resolved.Scheme != "https" {
			continue
		}
		absolute := resolved.String()
		if resolved.Hostname() == baseHost {
			if !seenInSite[absolute] {
				seenInSite[absolute] = true
				result.InSite = append(result.InSite, absolute)
			}
			continue
		}
		if !seenOutSite[absolute] {
			seenOutSite[absolute] = true
			result.OutSite = append(result.OutSite, absolute)
		}
	}
	return result, nil
}
