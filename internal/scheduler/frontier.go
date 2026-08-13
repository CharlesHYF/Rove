/*
 * 文件作用：SQLite FrontierStore -- URL 去重、入队、状态流转（规格书 §4，SQLite 仅承担 Runtime 状态）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-13
 */
// Package scheduler 提供 frontier 持久化与抓取调度。
package scheduler

import (
	"context"
	"database/sql"
	"time"

	_ "modernc.org/sqlite"

	"rove/pkg/rove"
)

// FrontierEntry 是 frontier 中的一条抓取单元（规格书 §4 字段）。
type FrontierEntry struct {
	NormalizedURL string
	Host          string
	Priority      int // 0-9，越大越优先，默认 5
	Depth         int
	SourceURL     string
	DiscoveredAt  time.Time
	NextFetchAt   time.Time
	RetryCount    int
	State         string // queued | fetching | done | failed | blocked
}

// Frontier 基于 SQLite 的持久化 frontier。
type Frontier struct {
	DB *sql.DB
}

// NewFrontier 打开（或创建）SQLite 数据库并初始化表结构。
func NewFrontier(driverPath string) (*Frontier, error) {

	db, err := sql.Open("sqlite", driverPath)
	if err != nil {
		return nil, rove.NewError("frontier.open", rove.CategoryIndex, true, "open frontier db: %v", err)
	}
	db.SetMaxOpenConns(1) // :memory: 数据库按连接隔离，必须固定单连接
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, rove.NewError("frontier.open", rove.CategoryIndex, true, "init frontier schema: %v", err)
	}
	return &Frontier{DB: db}, nil
}

// schemaSQL 建表语句。
const schemaSQL = `
CREATE TABLE IF NOT EXISTS frontier (
	normalized_url TEXT PRIMARY KEY,
	host           TEXT NOT NULL,
	priority       INTEGER NOT NULL DEFAULT 5,
	depth          INTEGER NOT NULL DEFAULT 0,
	source_url     TEXT NOT NULL DEFAULT '',
	discovered_at  TEXT NOT NULL,
	next_fetch_at  TEXT NOT NULL,
	retry_count    INTEGER NOT NULL DEFAULT 0,
	state          TEXT NOT NULL DEFAULT 'queued'
);
CREATE INDEX IF NOT EXISTS idx_frontier_due ON frontier (state, next_fetch_at, priority DESC);
`

// Close 关闭数据库。
func (f *Frontier) Close() error {

	return f.DB.Close()
}

// Enqueue 批量入队（normalized_url 主键去重；done/failed 终态重置为 queued 支持重抓，在途状态忽略），返回实际入队数。
func (f *Frontier) Enqueue(ctx context.Context, entries []FrontierEntry) (int, error) {

	tx, err := f.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, rove.NewError("frontier.query", rove.CategoryIndex, true, "begin tx: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO frontier
		(normalized_url, host, priority, depth, source_url, discovered_at, next_fetch_at, retry_count, state)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(normalized_url) DO UPDATE SET
			host = excluded.host,
			priority = excluded.priority,
			depth = excluded.depth,
			source_url = excluded.source_url,
			discovered_at = excluded.discovered_at,
			next_fetch_at = excluded.next_fetch_at,
			retry_count = excluded.retry_count,
			state = 'queued'
		WHERE frontier.state IN ('done', 'failed')`)
	if err != nil {
		return 0, rove.NewError("frontier.query", rove.CategoryIndex, true, "prepare insert: %v", err)
	}
	defer stmt.Close()

	added := 0
	for _, entry := range entries {
		result, err := stmt.ExecContext(ctx,
			entry.NormalizedURL, entry.Host, entry.Priority, entry.Depth, entry.SourceURL,
			entry.DiscoveredAt.UTC().Format(time.RFC3339), entry.NextFetchAt.UTC().Format(time.RFC3339),
			entry.RetryCount, entry.State)
		if err != nil {
			return 0, rove.NewError("frontier.query", rove.CategoryIndex, true, "insert entry: %v", err)
		}
		if rows, _ := result.RowsAffected(); rows > 0 {
			added++
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, rove.NewError("frontier.query", rove.CategoryIndex, true, "commit tx: %v", err)
	}
	return added, nil
}

// Count 统计指定状态的条目数。
func (f *Frontier) Count(ctx context.Context, state string) (int, error) {

	var count int
	if err := f.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM frontier WHERE state = ?`, state).Scan(&count); err != nil {
		return 0, rove.NewError("frontier.query", rove.CategoryIndex, true, "count: %v", err)
	}
	return count, nil
}

// UpdateState 更新条目状态、重试次数与下次可抓取时间。
func (f *Frontier) UpdateState(ctx context.Context, normalizedURL, state string, retryCount int, nextFetchAt time.Time) error {

	_, err := f.DB.ExecContext(ctx, `UPDATE frontier SET state=?, retry_count=?, next_fetch_at=? WHERE normalized_url=?`,
		state, retryCount, nextFetchAt.UTC().Format(time.RFC3339), normalizedURL)
	if err != nil {
		return rove.NewError("frontier.query", rove.CategoryIndex, true, "update state: %v", err)
	}
	return nil
}
