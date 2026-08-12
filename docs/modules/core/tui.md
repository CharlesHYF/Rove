# TUI（tui）
> 模块职责：五视图 Runtime Debugger（Search/Crawl/Browse/Index/Runtime），订阅 Runtime Telemetry（规格书 §11）。
> 负责人：Charles
> 系统模块：core
> 关联服务：tui、internal/app、internal/runtime

## 终端界面
### 功能描述
Bubble Tea 五视图：Search（输入 + 结果 + 分数分解 + 耗时）、Crawl（crawl 事件流）、Browse（fetch 事件 + 指引）、Index（状态 + index 事件）、Runtime（事件日志）。Tab 切换视图；Search 视图输入查询回车异步执行。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| New | `tui.New(search, indexService, inspector) *Model` | - |
| Init | `(*Model).Init() tea.Cmd` | 刷新索引状态 |
| Update | `(*Model).Update(tea.Msg) (tea.Model, tea.Cmd)` | 键盘/异步消息 |
| View | `(*Model).View() string` | 渲染当前视图 |

### 入参要求
services 可 nil（视图降级）；inspector 提供事件与状态。

### 返回
无；Ctrl+C/Esc 退出。
