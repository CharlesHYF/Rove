/*
 * 文件作用：RuntimeInspector -- 向 TUI 暴露索引状态与 telemetry 事件（规格书 §13）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package app

import (
	"context"

	"rove/internal/runtime"
	"rove/pkg/index"
)

// RuntimeInspector 聚合运行时观测数据。
type RuntimeInspector struct {
	bus          *runtime.Bus
	indexService *IndexService
}

// NewInspector 构造 RuntimeInspector。
func NewInspector(bus *runtime.Bus, indexService *IndexService) *RuntimeInspector {

	return &RuntimeInspector{bus: bus, indexService: indexService}
}

// Status 返回索引状态。
func (r *RuntimeInspector) Status(ctx context.Context) (*index.Status, error) {

	return r.indexService.Status(ctx)
}

// Events 返回最近事件。
func (r *RuntimeInspector) Events(limit int) []runtime.Event {

	return r.bus.Recent(limit)
}

// Subscribe 返回事件流。
func (r *RuntimeInspector) Subscribe() <-chan runtime.Event {

	return r.bus.Subscribe()
}
