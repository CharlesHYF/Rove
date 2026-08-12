# 浏览器（browser）
> 模块职责：Chromium 渲染与页面状态 -- HTTP 内容不足时的升级路径（规格书 §5.5 / FR-BRW-002）。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/browser、pkg/fetch、cmd/rove

## 渲染抓取与页面状态
### 功能描述
浏览器进程复用（每 fetch 新 context 隔离 Cookie），渲染 SPA 后返回 RawDocument（FetchMethod=browser）进入统一管线；Page State 输出文本/链接/交互元素（DOM 路径稳定 id：结构不变时跨加载一致，动态 DOM 变化时降级为快照内路径）。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| NewManager | `NewManager(Options) (*Manager, error)` | 发现可执行文件并启动分配器 |
| Fetch | `(*Manager).Fetch(ctx, url) (*document.RawDocument, error)` | 渲染 HTML + 最终 URL |
| Browse | `(*Manager).Browse(ctx, url) (*PageState, error)` | 页面状态（含稳定 id） |
| DiscoverExecutable | `DiscoverExecutable(explicit string) (string, error)` | 显式路径 -> 常见路径 -> PATH |

### 入参要求
可执行文件可用（ROVE_BROWSER_EXECUTABLE 或自动发现）；Timeout 默认 30s。

### 返回
错误：browser.not_found（不可重试）/ browser.navigation（可重试）；StatusCode 记 200（决策 3）。
