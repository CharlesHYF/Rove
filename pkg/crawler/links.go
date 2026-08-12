/*
 * 文件作用：链接发现与抓取策略 -- 解析相对/绝对链接，规范化（含去 tracking）后按策略放行（规格书 §4）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package crawler

import (
	"context"
	"net/url"
	"strings"

	"rove/pkg/content"
)

// Policy 抓取策略。
type Policy struct {
	MaxDepth        int
	DomainAllowlist []string
	SameHostOnly    bool // 仅抓取与 base 同 host 的链接（默认 true）
}

// CandidateLinks 从 ParsedContent 提取可入队的链接：解析 -> 规范化（含去 tracking）-> 策略过滤。
func (p *Policy) CandidateLinks(baseURL string, parsed *content.ParsedContent) ([]string, error) {

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	baseHost := strings.ToLower(base.Hostname())

	var links []string
	for _, rawLink := range parsed.Links {
		resolved := resolveLink(base, rawLink)
		if resolved == "" {
			continue
		}
		if resolved == baseURL {
			continue // 自链接/纯片段，跳过
		}
		host := resolvedHost(resolved)
		if p.SameHostOnly && !strings.EqualFold(host, baseHost) {
			continue
		}
		if len(p.DomainAllowlist) > 0 && !p.AllowedDomain(host) {
			continue
		}
		links = append(links, resolved)
	}
	return links, nil
}

// AllowedDomain 判断 host 是否在 allowlist 中；allowlist 为空时放行一切。
func (p *Policy) AllowedDomain(host string) bool {

	if len(p.DomainAllowlist) == 0 {
		return true
	}
	for _, allowed := range p.DomainAllowlist {
		if strings.EqualFold(host, allowed) {
			return true
		}
	}
	return false
}

// resolveLink 解析相对/绝对链接，过滤非法 scheme 与片段，规范化后返回。
func resolveLink(base *url.URL, rawLink string) string {

	rawLink = strings.TrimSpace(rawLink)
	if strings.HasPrefix(rawLink, "#") {
		return "" // 纯片段链接无抓取价值
	}
	ref, err := url.Parse(rawLink)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(ref)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}
	if resolved.Fragment != "" {
		resolved.Fragment = ""
	}
	normalized, err := NormalizeURL(resolved.String())
	if err != nil {
		return ""
	}
	return normalized
}

// resolvedHost 提取 hostname（小写）。
func resolvedHost(urlStr string) string {

	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}

// NormalizeURL 规范化 URL（复用 Canonicalizer，无 rel=canonical）。
func NormalizeURL(rawURL string) (string, error) {

	canonicalizer := &content.Canonicalizer{}
	return canonicalizer.Canonicalize(context.Background(), rawURL, "")
}
