/*
 * 文件作用：Canonicalizer 的单元测试 -- 覆盖 scheme/host 归并、默认端口、fragment、tracking 参数与 rel=canonical 解析。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanonicalize(t *testing.T) {

	canonicalizer := &Canonicalizer{}
	cases := []struct {
		name   string
		rawURL string
		canon  string
		want   string
	}{
		{"scheme case", "HTTP://Example.com/A", "", "http://example.com/A"},
		{"default port stripped", "https://example.com:443/a", "", "https://example.com/a"},
		{"fragment stripped", "https://example.com/a#section", "", "https://example.com/a"},
		{"trailing slash", "https://example.com/a/", "", "https://example.com/a"},
		{"root slash kept", "https://example.com/", "", "https://example.com/"},
		{"tracking removed", "https://example.com/a?utm_source=x&q=1&fbclid=z", "", "https://example.com/a?q=1"},
		{"rel canonical absolute", "https://example.com/a", "https://cdn.example.com/real", "https://cdn.example.com/real"},
		{"rel canonical relative", "https://example.com/a/b", "../real", "https://example.com/real"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			got, err := canonicalizer.Canonicalize(context.Background(), tc.rawURL, tc.canon)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestCanonicalizeInvalid(t *testing.T) {

	canonicalizer := &Canonicalizer{}
	_, err := canonicalizer.Canonicalize(context.Background(), "://bad", "")
	require.Error(t, err, "invalid url must error")
}
