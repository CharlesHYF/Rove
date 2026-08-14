# Query Router（pkg/retrieval Router）
> 模块职责：确定性查询垂类推断（web/docs/code/academic）与检索路由（PRD FR-QRY-001，P1，无 LLM）。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/retrieval、pkg/elastic、cmd/rove、protocol/mcp

## Query Router
### 功能描述
查询垂类路由：`rove search --vertical auto|web|docs|code|academic`（默认 auto）。auto 模式用确定性规则（代码标记 -> 学术标记 -> 文档标记 -> 默认 web，均无 LLM）推断垂类并在结果中回显。检索落点：chunks 索引新增 `source_type` keyword 字段；仅 docs/code/academic 垂类应用 source_type 过滤，web 为免过滤兜底（与历史行为一致）。搜索体沿用现有 buildFilters（BM25 腿），KNN 腿当前不支持过滤（与 domain/language 过滤的既有语义一致）。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| RouteVertical | `RouteVertical(text string) Vertical` | 确定性垂类推断 |
| ParseVertical | `ParseVertical(s string) (Vertical, error)` | 解析并校验垂类参数（auto/web/docs/code/academic） |
| Search | `(*Retriever).Search(ctx, q *Query) (*SearchResult, error)` | q.Vertical 生效；结果回显 Vertical |

### 入参要求
vertical 取值限定 auto/web/docs/code/academic；非法值返回错误（不 silent failure）。

### 返回
SearchResult.Vertical 回显实际生效垂类；JSON 输出含 vertical 字段。

### 兼容性说明
source_type 字段仅在新建索引（rove index init）或迁移（rove index migrate）后的 chunks 索引上存在；旧索引上 docs/code/academic 过滤可能返回空（无该字段的文档不命中 term 过滤），web/auto 不受影响。
