/*
 * 文件作用：URL 规范化 -- 归并 scheme/host/端口/fragment/trailing slash，移除 tracking 参数并应用 rel=canonical。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"context"
	"net/url"
	"strings"

	"rove/pkg/rove"
)

// trackingParams 是需要移除的已知跟踪参数。
var trackingParams = map[string]bool{
	"utm_source": true, "utm_medium": true, "utm_campaign": true,
	"utm_term": true, "utm_content": true, "fbclid": true,
	"gclid": true, "gclsrc": true, "dclid": true,
	"mc_cid": true, "mc_eid": true, "igshid": true,
}

// Canonicalizer 负责 URL 规范化（规格书 §4.1）。
type Canonicalizer struct{}

// Canonicalize 归并 URL：小写 scheme/host、去 default port/fragment/tracking、trailing slash，
// 并应用 rel=canonical（相对时按 rawURL 解析）。
func (c *Canonicalizer) Canonicalize(ctx context.Context, rawURL, relCanonical string) (string, error) {

	target := rawURL
	if relCanonical != "" {
		base, err := url.Parse(rawURL)
		if err != nil {
			return "", rove.NewError("content.invalid_url", rove.CategoryContent, false, "invalid url: %q", rawURL)
		}
		ref, err := url.Parse(relCanonical)
		if err != nil {
			return "", rove.NewError("content.invalid_url", rove.CategoryContent, false, "invalid canonical url: %q", relCanonical)
		}
		target = base.ResolveReference(ref).String()
	}

	u, err := url.Parse(target)
	if err != nil || u.Host == "" {
		return "", rove.NewError("content.invalid_url", rove.CategoryContent, false, "invalid url: %q", target)
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Fragment = ""
	stripDefaultPort(u)

	query := u.Query()
	for key := range query {
		if trackingParams[strings.ToLower(key)] {
			query.Del(key)
		}
	}
	u.RawQuery = query.Encode()

	if len(u.Path) > 1 && strings.HasSuffix(u.Path, "/") {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}
	return u.String(), nil
}

// stripDefaultPort 移除 http:80 / https:443 默认端口。
func stripDefaultPort(u *url.URL) {

	if u.Port() == "" {
		return
	}
	if (u.Scheme == "http" && u.Port() == "80") || (u.Scheme == "https" && u.Port() == "443") {
		u.Host = u.Hostname()
	}
}
