# CLI（cmd/rove）
> 模块职责：Rove 命令行入口 -- 只做装配与参数校验（读配置、组装 Fetcher/Pipeline、输出），不写业务逻辑。
> 负责人：Charles
> 系统模块：core
> 关联服务：cmd/rove、internal/config、pkg/fetch、pkg/content

## fetch 命令
### 功能描述
抓取 URL 并输出标准 Document。auto 模式下页面内容不足（SPA/需交互）时返回 fetch.escalation_required 提示（浏览器能力在 M5 提供）。

### 接口信息
| 项 | 值 |
| --- | --- |
| 命令 | `rove fetch <url> [--mode auto|http|browser] [--json]` |
| 默认模式 | auto |
| JSON 输出 | document.Document（Go 字段名序列化；协议 schema 由 M2 protocol 包定义） |

### 入参要求
| 参数 | 校验规则 | 说明 |
| --- | --- | --- |
| url | cobra.ExactArgs(1) | 目标地址 |
| --mode | auto/http/browser | browser 在 M5 前返回 browser.not_implemented |
| --json | bool | 以 JSON 输出 |

### 返回
JSON：document.Document；人类可读：title/url/canonical/language/source/chunks 摘要 + duplicate warning。
