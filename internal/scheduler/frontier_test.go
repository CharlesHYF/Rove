/*
 * 文件作用：internal/scheduler frontier 的单元测试 -- 入队/去重/状态流转（SQLite :memory:）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFrontierEnqueueAndCount(t *testing.T) {

	frontier, err := NewFrontier(":memory:")
	require.NoError(t, err)
	defer frontier.Close()

	now := time.Now().UTC()
	added, err := frontier.Enqueue(context.Background(), []FrontierEntry{
		{NormalizedURL: "https://a.com/1", Host: "a.com", Priority: 5, Depth: 0, DiscoveredAt: now, NextFetchAt: now, State: "queued"},
		{NormalizedURL: "https://a.com/2", Host: "a.com", Priority: 5, Depth: 1, DiscoveredAt: now, NextFetchAt: now, State: "queued"},
	})
	require.NoError(t, err)
	require.Equal(t, 2, added)

	count, err := frontier.Count(context.Background(), "queued")
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestFrontierDedup(t *testing.T) {

	frontier, _ := NewFrontier(":memory:")
	defer frontier.Close()
	ctx := context.Background()
	now := time.Now().UTC()
	entry := FrontierEntry{NormalizedURL: "https://a.com/1", Host: "a.com", Priority: 5, DiscoveredAt: now, NextFetchAt: now, State: "queued"}

	added, err := frontier.Enqueue(ctx, []FrontierEntry{entry})
	require.NoError(t, err)
	require.Equal(t, 1, added)

	added, err = frontier.Enqueue(ctx, []FrontierEntry{entry})
	require.NoError(t, err)
	require.Equal(t, 0, added, "duplicate URL must be ignored")
}

func TestFrontierUpdateState(t *testing.T) {

	frontier, _ := NewFrontier(":memory:")
	defer frontier.Close()
	ctx := context.Background()
	now := time.Now().UTC()
	entry := FrontierEntry{NormalizedURL: "https://a.com/1", Host: "a.com", Priority: 5, DiscoveredAt: now, NextFetchAt: now, State: "queued"}
	_, _ = frontier.Enqueue(ctx, []FrontierEntry{entry})

	require.NoError(t, frontier.UpdateState(ctx, entry.NormalizedURL, "done", 0, now))

	count, err := frontier.Count(ctx, "done")
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
