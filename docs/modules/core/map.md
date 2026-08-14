# Map（internal/app MapService）
> 模块职责：URL Discovery -- 抓取单页并发现站内/站外链接，不强制完整内容抓取与索引（PRD FR-MAP-001，P1）。
> 负责人：Charles
> 系统模块：core
> 关联服务：internal/app、pkg/fetch、pkg/content

## Map 服务
### 功能描述
`rove map <url>` 抓取单页（HTTP First，auto 模式内容不足时浏览器升级，与 fetch 同一升级链路），解析后输出发现的链接，按宿主划分为站内（in_site）与站外（out_site）并去重。仅做 URL Discovery：不进入 Crawler、不建索引、不输出完整 Document。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| NewMap | `NewMap(cfg *config.Config) *MapService` | 按配置构造 |
| Map | `(*MapService).Map(ctx, url) (*MapResult, error)` | 抓取并发现链接 |

MapResult：`url`（抓取地址）、`canonical`（rel=canonical，空则取抓取地址）、`in_site`/`out_site`（去重后的链接列表）。

### 入参要求
url 必填；需网络可达；私网抓取须配置 allow_private=true。

### 返回
- 成功：MapResult（JSON 或人类可读摘要）；
- 失败：机器可读错误（fetch/content 类别），不 silent failure。
