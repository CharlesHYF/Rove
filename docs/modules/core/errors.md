# 错误模型（errors）
> 模块职责：定义 Rove 统一的 machine-readable 错误 -- 类别、可重试性与 JSON 序列化契约（NFR-001）。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/rove、全链路

## 错误构造
### 功能描述
所有错误路径返回 *rove.Error，禁止 silent failure；JSON 序列化字段为 code/category/retryable/message/details。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| NewError | `NewError(code string, category ErrorCategory, retryable bool, format string, args ...any) *Error` | 构造错误 |
| WithDetails | `(*Error).WithDetails(details map[string]any) *Error` | 附加结构化细节 |

### 入参要求
| 参数 | 说明 |
| --- | --- |
| code | 稳定错误码，如 fetch.ssrf_blocked |
| category | network/timeout/policy/content/browser/index/embedding |
| retryable | 是否可安全重试 |

### 返回
*rove.Error（实现 error 接口）。
