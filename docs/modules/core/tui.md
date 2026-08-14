# TUI（tui）
> 模块职责：对话式查询首页 + 五视图 Runtime Debugger（Search/Crawl/Browse/Index/Runtime），订阅 Runtime Telemetry（规格书 §11）。
> 负责人：Charles
> 系统模块：core
> 关联服务：tui、internal/app、internal/runtime

## 终端界面
### 功能描述
启动默认进入"对话"视图：输入自然语言问题回车，异步检索自有 ES 索引（无 LLM、无外部依赖），以结构化回答卡片展示直接回答、来源（最多 3 条：标题/URL/相关度）、分数分解与总耗时；Ctrl+E 展开完整分数与阶段耗时，Ctrl+L 清空会话，上/下方向键翻阅历史输入，PgUp/PgDn 翻页。Tab/数字键切换其余五视图：搜索（调试视角结果与分数）、抓取（crawl 事件流）、浏览（fetch 事件与指引）、索引（状态与事件）、运行时（事件日志）。

### 回答规则（确定性，无 LLM）
- 检索 TopK=5；有结果时直接回答 = 第 1 条 Evidence 文本（最多 8 行、300 字符截断）；无结果回复"没有找到相关内容，试试换种说法"；检索出错展示错误摘要，界面不崩。
- 诚实兜底：查询关键词（ASCII 词优先，无则退回中文整词）在顶部结果的正文/标题/标题路径中完全没有词面命中时，不做权威回答，提示"当前语料可能未覆盖"并将结果降级为"可能相关，仅供参考"（lowConfidence）。
- 相关度 = Evidence 归一化分数（0-1）；分数分解来自 SearchHit.Scores（lexical/vector/fusion/rank）；耗时 = SearchResult.Timings 各阶段之和。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| New | `tui.New(search, indexService, inspector) *Model` | - |
| Init | `(*Model).Init() tea.Cmd` | 刷新索引状态 |
| Update | `(*Model).Update(tea.Msg) (tea.Model, tea.Cmd)` | 键盘/异步消息（对话输入、历史翻阅、展开、清空） |
| View | `(*Model).View() string` | 渲染当前视图 |
| buildChatAnswer | `buildChatAnswer(query, items, hits, timings) chatMessage` | 由检索结果构造回答卡片（纯函数） |
| buildChatError | `buildChatError(err) chatMessage` | 由错误构造回答卡片（纯函数） |

### 入参要求
services 可 nil（视图降级）；inspector 提供事件与状态；对话检索依赖 SearchService.Search 与 SearchService.Evidence。

### 返回
无；Ctrl+C/Esc 退出。
