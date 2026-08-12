/*
 * 文件作用：Runtime Telemetry 事件总线 -- 业务事件广播与最近事件查询（TUI 订阅，规格书 §13）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package runtime

import (
	"sync"
	"time"
)

// EventType 事件类别。
type EventType string

const (
	Search EventType = "search"
	Fetch  EventType = "fetch"
	Crawl  EventType = "crawl"
	Index  EventType = "index"
)

// Event 是一条运行时事件。
type Event struct {
	Type    EventType
	TraceID string
	Message string
	Fields  map[string]string
	At      time.Time
}

// Bus 是事件总线（环形缓冲 + 订阅广播）。
type Bus struct {
	mu       sync.Mutex
	buffer   []Event
	capacity int
	subs     map[chan Event]struct{}
	closed   bool
}

// NewBus 构造事件总线。
func NewBus(capacity int) *Bus {

	return &Bus{buffer: make([]Event, 0, capacity), capacity: capacity, subs: map[chan Event]struct{}{}}
}

// Publish 广播事件；缓冲满时丢弃最旧；channel 满时跳过订阅者（不阻塞）。
func (b *Bus) Publish(event Event) {

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	if event.At.IsZero() {
		event.At = time.Now()
	}
	b.buffer = append(b.buffer, event)
	if len(b.buffer) > b.capacity {
		b.buffer = b.buffer[len(b.buffer)-b.capacity:]
	}
	for subscriber := range b.subs {
		select {
		case subscriber <- event:
		default:
		}
	}
}

// Subscribe 返回事件流。
func (b *Bus) Subscribe() <-chan Event {

	channel := make(chan Event, 16)
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed {
		b.subs[channel] = struct{}{}
	}
	return channel
}

// Recent 返回最近 N 条事件（新->旧）。
func (b *Bus) Recent(limit int) []Event {

	b.mu.Lock()
	defer b.mu.Unlock()
	count := limit
	if count > len(b.buffer) {
		count = len(b.buffer)
	}
	result := make([]Event, 0, count)
	for index := len(b.buffer) - count; index < len(b.buffer); index++ {
		result = append(result, b.buffer[index])
	}
	return result
}

// Close 关闭总线并通知订阅者。
func (b *Bus) Close() {

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for subscriber := range b.subs {
		close(subscriber)
	}
	b.subs = map[chan Event]struct{}{}
}
