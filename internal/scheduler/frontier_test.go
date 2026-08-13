/*
 * 文件作用：internal/scheduler frontier 的单元测试 -- 入队/去重/状态流转（SQLite :memory:）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-13
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

// frontierRow 是测试查询用的一条 frontier 记录快照。
type frontierRow struct {
	state       string
	nextFetchAt string
	priority    int
}

// readFrontierRow 读取单条 frontier 记录，便于断言状态流转细节。
func readFrontierRow(t *testing.T, frontier *Frontier, normalizedURL string) frontierRow {

	var row frontierRow
	err := frontier.DB.QueryRow(`SELECT state, next_fetch_at, priority FROM frontier WHERE normalized_url = ?`, normalizedURL).
		Scan(&row.state, &row.nextFetchAt, &row.priority)
	require.NoError(t, err)
	return row
}

// enqueueEntry 构造一条 queued 状态的入队条目。
func enqueueEntry(normalizedURL string, at time.Time) FrontierEntry {

	return FrontierEntry{NormalizedURL: normalizedURL, Host: "a.com", Priority: 5, Depth: 0, DiscoveredAt: at, NextFetchAt: at, State: "queued"}
}

// TestFrontierRequeueDoneSeed 回归：已 done 的 seed 重新入队须重置为 queued 并刷新 next_fetch_at。
func TestFrontierRequeueDoneSeed(t *testing.T) {

	frontier, _ := NewFrontier(":memory:")
	defer frontier.Close()
	ctx := context.Background()
	now := time.Now().UTC()
	url := "https://a.com/1"
	added, err := frontier.Enqueue(ctx, []FrontierEntry{enqueueEntry(url, now)})
	require.NoError(t, err)
	require.Equal(t, 1, added)
	require.NoError(t, frontier.UpdateState(ctx, url, "done", 0, now))

	later := now.Add(time.Hour)
	added, err = frontier.Enqueue(ctx, []FrontierEntry{enqueueEntry(url, later)})
	require.NoError(t, err)
	require.Equal(t, 1, added, "done 状态须重新入队")
	require.Equal(t, "queued", readFrontierRow(t, frontier, url).state)
	require.Equal(t, later.Format(time.RFC3339), readFrontierRow(t, frontier, url).nextFetchAt, "next_fetch_at 须刷新")
}

// TestFrontierRequeueFailedSeed 回归：failed 状态同样重置为 queued。
func TestFrontierRequeueFailedSeed(t *testing.T) {

	frontier, _ := NewFrontier(":memory:")
	defer frontier.Close()
	ctx := context.Background()
	now := time.Now().UTC()
	url := "https://a.com/2"
	_, _ = frontier.Enqueue(ctx, []FrontierEntry{enqueueEntry(url, now)})
	require.NoError(t, frontier.UpdateState(ctx, url, "failed", 5, now))

	added, err := frontier.Enqueue(ctx, []FrontierEntry{enqueueEntry(url, now)})
	require.NoError(t, err)
	require.Equal(t, 1, added, "failed 状态须重新入队")
	require.Equal(t, "queued", readFrontierRow(t, frontier, url).state)
}

// TestFrontierEnqueueSkipsInFlight 回归：queued/fetching 状态不重置，避免重复抓取。
func TestFrontierEnqueueSkipsInFlight(t *testing.T) {

	frontier, _ := NewFrontier(":memory:")
	defer frontier.Close()
	ctx := context.Background()
	now := time.Now().UTC()

	queuedURL := "https://a.com/queued"
	fetchingURL := "https://a.com/fetching"
	_, _ = frontier.Enqueue(ctx, []FrontierEntry{enqueueEntry(queuedURL, now)})
	_, _ = frontier.Enqueue(ctx, []FrontierEntry{enqueueEntry(fetchingURL, now)})
	require.NoError(t, frontier.UpdateState(ctx, fetchingURL, "fetching", 0, now))

	added, err := frontier.Enqueue(ctx, []FrontierEntry{enqueueEntry(queuedURL, now), enqueueEntry(fetchingURL, now)})
	require.NoError(t, err)
	require.Equal(t, 0, added, "queued/fetching 在途条目必须忽略")
	require.Equal(t, "queued", readFrontierRow(t, frontier, queuedURL).state)
	require.Equal(t, "fetching", readFrontierRow(t, frontier, fetchingURL).state)
}
