/*
 * 文件作用：internal/runtime 事件总线的单元测试 -- 发布/订阅/环形缓冲。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package runtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBusPublishSubscribe(t *testing.T) {

	bus := NewBus(10)
	defer bus.Close()

	bus.Publish(Event{Type: Search, TraceID: "t1", Message: "query=rove"})
	bus.Publish(Event{Type: Index, TraceID: "t2", Message: "indexed doc-1"})

	recent := bus.Recent(10)
	require.Len(t, recent, 2)
	require.Equal(t, Search, recent[0].Type)
	require.Equal(t, "t2", recent[1].TraceID)

	ch := bus.Subscribe()
	bus.Publish(Event{Type: Fetch, TraceID: "t3", Message: "fetched"})
	select {
	case event := <-ch:
		require.Equal(t, Fetch, event.Type)
	case <-time.After(time.Second):
		t.Fatal("subscriber must receive event")
	}
}

func TestBusBufferOverflow(t *testing.T) {

	bus := NewBus(2)
	defer bus.Close()
	for index := 0; index < 5; index++ {
		bus.Publish(Event{Type: Index, Message: "m"})
	}
	recent := bus.Recent(10)
	require.Len(t, recent, 2, "buffer must keep most recent 2")
}
