/*
 * 文件作用：pkg/crawler 链接策略的单元测试 -- 过滤、规范化与域名策略。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package crawler

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/content"
)

func TestCandidateLinks(t *testing.T) {

	policy := Policy{MaxDepth: 3, SameHostOnly: true}
	parsed := &content.ParsedContent{
		Links: []string{
			"/page2",
			"https://example.com/page3?utm_source=x",
			"https://other.com/external", // 跨域，默认排除
			"javascript:void(0)",          // 非法 scheme，排除
			"#fragment",                   // 片段，排除
		},
	}
	links, err := policy.CandidateLinks("https://example.com/start", parsed)
	require.NoError(t, err)
	require.Equal(t, []string{"https://example.com/page2", "https://example.com/page3"}, links)
}

func TestPolicyAllowedDomain(t *testing.T) {

	noAllowlist := Policy{}
	require.True(t, noAllowlist.AllowedDomain("anything.com"), "empty allowlist allows all")

	allowlist := Policy{DomainAllowlist: []string{"a.com", "b.com"}}
	require.True(t, allowlist.AllowedDomain("a.com"))
	require.False(t, allowlist.AllowedDomain("c.com"))
}

func TestNormalizeURL(t *testing.T) {

	normalized, err := NormalizeURL("HTTPS://Example.com/A/?utm_source=x#frag")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/A", normalized, "path case is preserved, scheme/host lowercased, tracking/fragment stripped")
}
