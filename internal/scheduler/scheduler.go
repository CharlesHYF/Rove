/*
 * 文件作用：Scheduler -- 按到期时间与 host 令牌桶选择抓取条目，处理成功/失败状态流转（规格书 §4）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package scheduler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Options 调度配置。
type Options struct {
	HostConcurrency int           // 每 host 并发上限，默认 1
	CrawlDelay      time.Duration // robots crawl-delay 之外的最小抓取间隔，默认 0
	MaxRetries      int           // 失败最大重试次数，默认 5
}

// Scheduler 基于 frontier 的调度器。
type Scheduler struct {
	frontier   *Frontier
	opts       Options
	hostTokens map[string]chan struct{}
}

// NewScheduler 构造调度器。
func NewScheduler(frontier *Frontier, opts Options) *Scheduler {

	if opts.HostConcurrency <= 0 {
		opts.HostConcurrency = 1
	}
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 5
	}
	return &Scheduler{frontier: frontier, opts: opts, hostTokens: map[string]chan struct{}{}}
}

// Next 取出最早到期且 host 令牌可用的 queued 条目，并标记为 fetching；无可用条目返回 nil。
func (s *Scheduler) Next(ctx context.Context) (*FrontierEntry, error) {

	return s.claimNext(ctx)
}

// Complete 结束一个抓取：成功 -> done；失败且未超限 -> 退避后回 queued；超限 -> failed。
func (s *Scheduler) Complete(ctx context.Context, entry *FrontierEntry, err error) error {

	defer s.releaseHostToken(entry.Host)
	if err == nil {
		return s.frontier.UpdateState(ctx, entry.NormalizedURL, "done", entry.RetryCount, time.Now().UTC())
	}

	retries := entry.RetryCount + 1
	if retries > s.opts.MaxRetries {
		return s.frontier.UpdateState(ctx, entry.NormalizedURL, "failed", retries, time.Now().UTC())
	}
	nextFetch := time.Now().UTC().Add(retryBackoff(retries) + s.opts.CrawlDelay)
	return s.frontier.UpdateState(ctx, entry.NormalizedURL, "queued", retries, nextFetch)
}

// claimNext 原子认领一个到期条目（UPDATE ... RETURNING，避免并发双取），host 令牌不可用则回退。
func (s *Scheduler) claimNext(ctx context.Context) (*FrontierEntry, error) {

	var entry FrontierEntry
	var discovered, nextFetch string
	err := s.frontier.DB.QueryRowContext(ctx, `UPDATE frontier SET state='fetching'
		WHERE normalized_url = (SELECT normalized_url FROM frontier
			WHERE state='queued' AND next_fetch_at <= ? ORDER BY priority DESC, next_fetch_at LIMIT 1)
		RETURNING normalized_url, host, priority, depth, source_url, discovered_at, next_fetch_at, retry_count, state`,
		time.Now().UTC().Format(time.RFC3339)).Scan(
		&entry.NormalizedURL, &entry.Host, &entry.Priority, &entry.Depth, &entry.SourceURL,
		&discovered, &nextFetch, &entry.RetryCount, &entry.State)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("claim next: %w", err)
	}
	if !s.acquireHostToken(entry.Host) {
		// host 令牌不可用，回退为 queued（下次到期即可再认领）
		_ = s.frontier.UpdateState(ctx, entry.NormalizedURL, "queued", entry.RetryCount, time.Now().UTC())
		return nil, nil
	}
	entry.DiscoveredAt = parseTime(discovered)
	entry.NextFetchAt = parseTime(nextFetch)
	return &entry, nil
}

// acquireHostToken 尝试占用 host 令牌（满则返回 false，不阻塞）。
func (s *Scheduler) acquireHostToken(host string) bool {

	tokens, ok := s.hostTokens[host]
	if !ok {
		tokens = make(chan struct{}, s.opts.HostConcurrency)
		s.hostTokens[host] = tokens
	}
	select {
	case tokens <- struct{}{}:
		return true
	default:
		return false
	}
}

// releaseHostToken 释放 host 令牌。
func (s *Scheduler) releaseHostToken(host string) {

	if tokens, ok := s.hostTokens[host]; ok {
		select {
		case <-tokens:
		default:
		}
	}
}

// parseTime 解析 RFC3339 时间字符串，失败返回零值。
func parseTime(value string) time.Time {

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

// retryBackoff 返回第 attempt 次重试的退避时长（2^attempt 秒）。
func retryBackoff(attempt int) time.Duration {

	return time.Duration(1<<uint(attempt)) * time.Second
}
