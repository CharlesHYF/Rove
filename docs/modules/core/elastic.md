# ES 客户端（elastic）
> 模块职责：Elasticsearch 8.x 客户端封装 -- 连接、Ping、索引 mapping 与 alias 生命周期（规格书 §5.3）。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/elastic、pkg/index、pkg/retrieval

## 客户端与索引
### 功能描述
构造 ES 客户端，幂等创建 rove-documents / rove-chunks（物理 v001 + alias）；提供底层操作（bulk/delete/count/stats/mget/search）。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| New | `New(esURL, username, password, prefix string) (*Client, error)` | prefix 默认 rove |
| Ping | `(*Client).Ping(ctx) error` | 连通性检查 |
| EnsureIndexes | `(*Client).EnsureIndexes(ctx, embeddingDim int) error` | 幂等创建索引 + alias |
| ExistsAlias | `(*Client).ExistsAlias(ctx, alias string) (bool, error)` | alias 存在性 |
| DeleteIndex | `(*Client).DeleteIndex(ctx, alias string) error` | 删除索引（含物理） |
| SearchChunks | `(*Client).SearchChunks(ctx, alias string, body []byte) ([]ChunkHit, error)` | chunk 检索 |
| MGetDocuments | `(*Client).MGetDocuments(ctx, alias string, ids []string) (map[string]*document.Document, error)` | 文档回填 |

### 入参要求
embeddingDim > 0（默认 256）；alias 名 = `<prefix>-documents` / `<prefix>-chunks`。

### 返回
错误统一 *rove.Error（category=index，可重试：连接失败/429/503）。
