# HTTP 抓取（fetch）
> 模块职责：HTTP First 抓取能力 -- redirect 跟随、超时、重试、压缩、charset 转换、body 上限、SSRF 策略与 Browser Escalation 触发信号。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/fetch、pkg/content、pkg/browser(M5)

## Fetch（抓取）
### 功能描述
执行一次 HTTP 抓取并返回 RawDocument。auto 模式下 HTTP 内容不足（401/403、正文 < 500 字节、HTML script 占比 > 30%）时返回 ErrEscalationRequired（内部信号，M5 实现浏览器升级）；网络错误自动重试 3 次（退避 1s/4s）。

### 接口信息
| 项 | 值 |
| --- | --- |
| 类型 | HTTPFetcher |
| 构造函数 | `NewHTTP(opts Options) *HTTPFetcher` |
| 方法 | `Fetch(ctx context.Context, req *Request) (*document.RawDocument, error)` |

### 入参要求
| 参数 | 类型 | 校验规则 | 说明 |
| --- | --- | --- | --- |
| URL | string | 合法 URL、非 loopback/私网（allow_private=false 时） | 目标地址 |
| Mode | FetchMode | auto/http/browser | browser 在 M5 前返回 browser.not_implemented |
| MaxBodyBytes | int64 | > 0 | 默认 10MB，超限返回 fetch.body_too_large |
| Timeout | Duration | > 0 | 默认 30s |
| MaxRedirects | int | > 0 | 默认 10，超限返回 fetch.redirect_loop（不可重试） |

### 返回
*document.RawDocument；错误码见规格书 §5.6（fetch.invalid_url / fetch.ssrf_blocked / fetch.redirect_loop / fetch.body_too_large / fetch.network / browser.not_implemented）。

### 边界与异常
- SSRF：redirect 每跳重新执行网络策略，策略违规返回 fetch.ssrf_blocked（不可重试）
- charset 探测失败时保留原始字节（降级，不报错）
- context 取消立即返回 ctx.Err()
