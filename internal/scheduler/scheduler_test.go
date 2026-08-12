/*
 * 文件作用：internal/scheduler 调度器的单元测试 -- 取号/完成/失败退避（SQLite :memory:）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rove/pkg/rove"
)

// newTestError 构造可重试的测试错误。
func newTestError(message string) error {

	return rove.NewError("test.fetch", rove.CategoryNetwork, true, "%s", message)
}

var errTestFetch = newTestError("fetch failed")

func TestRetryBackoff(t *testing.T) {

	require.Equal(t, 1*time.Second, retryBackoff(0))
	require.Equal(t, 2*time.Second, retryBackoff(1))
	require.Equal(t, 8*time.Second, retryBackoff(3))
}

func TestSchedulerNextAndComplete(t *testing.T) {

	frontier, _ := NewFrontier(":memory:")
	defer frontier.Close()
	ctx := context.Background()
	now := time.Now().UTC()

	entry := FrontierEntry{NormalizedURL: "https://a.com/1", Host: "a.com", Priority: 8, Depth: 0, DiscoveredAt: now, NextFetchAt: now, State: "queued"}
	_, _ = frontier.Enqueue(ctx, []FrontierEntry{entry})

	sched := NewScheduler(frontier, Options{HostConcurrency: 1})
	got, err := sched.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "https://a.com/1", got.NormalizedURL)
	require.Equal(t, "a.com", got.Host)
	require.Equal(t, 8, got.Priority)

	// Next 后条目进入 fetching，不可重复取
	_, err = sched.Next(ctx)
	require.NoError(t, err)
	count, _ := frontier.Count(ctx, "fetching")
	require.Equal(t, 1, count)

	// 成功 -> done
	require.NoError(t, sched.Complete(ctx, got, nil))
	count, _ = frontier.Count(ctx, "done")
	require.Equal(t, 1, count)
}

func TestSchedulerFailureBackoff(t *testing.T) {

	frontier, _ := NewFrontier(":memory:")
	defer frontier.Close()
	ctx := context.Background()
	now := time.Now().UTC()

	entry := FrontierEntry{NormalizedURL: "https://a.com/1", Host: "a.com", Priority: 5, Depth: 0, DiscoveredAt: now, NextFetchAt: now, State: "queued"}
	_, _ = frontier.Enqueue(ctx, []FrontierEntry{entry})

	sched := NewScheduler(frontier, Options{HostConcurrency: 1, MaxRetries: 2})
	got, _ := sched.Next(ctx)
	require.NoError(t, sched.Complete(ctx, got, errTestFetch))

	// 重试 1 次后仍在 queued，且 next_fetch_at 延后
	count, _ := frontier.Count(ctx, "queued")
	require.Equal(t, 1, count)

	got, _ = sched.Next(ctx)
	require.NoError(t, sched.Complete(ctx, got, errTestFetch))
	count, _ = frontier.Count(ctx, "failed")
	require.Equal(t, 1, count, "retry limit exceeded -> failed")
}
