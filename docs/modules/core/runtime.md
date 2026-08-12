# 运行时遥测（runtime）
> 模块职责：内部 Runtime Telemetry 事件总线 -- TUI 订阅与最近事件查询（规格书 §13；OTel 导出器为 P2 延后项）。
> 负责人：Charles
> 系统模块：core
> 关联服务：internal/runtime、internal/app、tui

## 事件总线
### 功能描述
业务事件（search/fetch/crawl/index）广播与环形缓冲；Publish 非阻塞（满丢最旧），Subscribe 返回事件流。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| NewBus | `NewBus(capacity int) *Bus` | - |
| Publish | `(*Bus).Publish(event Event)` | 非阻塞广播 |
| Subscribe | `(*Bus).Subscribe() <-chan Event` | 事件流 |
| Recent | `(*Bus).Recent(limit int) []Event` | 最近 N 条（新->旧） |
| Close | `(*Bus).Close()` | 关闭并通知订阅者 |

### 入参要求
Event{Type, TraceID, Message, Fields, At}；At 为空自动填充。

### 返回
无；错误不适用（非阻塞设计）。
