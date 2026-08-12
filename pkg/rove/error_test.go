/*
 * 文件作用：pkg/rove 错误模型的单元测试 -- 验证错误构造、JSON 序列化契约与可重试性语义。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package rove

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewErrorMessage(t *testing.T) {

	e := NewError("fetch.invalid_url", CategoryContent, false, "invalid url: %s", "://bad")
	require.Equal(t, "fetch.invalid_url", e.Code)
	require.Equal(t, CategoryContent, e.Category)
	require.False(t, e.Retryable)
	require.Equal(t, "invalid url: ://bad", e.Error())
}

func TestErrorJSONShape(t *testing.T) {

	e := NewError("fetch.ssrf_blocked", CategoryPolicy, false, "blocked url: %s", "127.0.0.1").
		WithDetails(map[string]any{"host": "127.0.0.1"})
	data, err := json.Marshal(e)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))

	for _, key := range []string{"code", "category", "retryable", "message", "details"} {
		require.Contains(t, m, key, "missing json field %q", key)
	}
}

func TestRetryableCategories(t *testing.T) {

	require.True(t, NewError("x", CategoryNetwork, true, "boom").Retryable)
	require.False(t, NewError("x", CategoryPolicy, false, "blocked").Retryable)
}
