# Index Migrate（pkg/index Migrator）
> 模块职责：索引 schema 演进工具 -- create vNext -> reindex/重嵌入 -> 计数校验 -> alias 原子切换（规格书 §5.3，P1）。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/index、pkg/elastic、cmd/rove

## 索引迁移
### 功能描述
`rove index migrate [--delete-old]` 执行 schema 演进：解析两个 alias 当前物理索引（-vNNN）-> 创建 vNext 物理索引（mapping 与当前一致）-> documents 走 ES _reindex；chunks 因 embedding 不入 _source，按分页扫描旧索引、用 Embedder 重嵌入向量并批量写入新索引（同时按 documents 索引回填 source_type，供 Query Router 使用）-> 新旧计数一致才允许 alias 原子切换（不一致保持旧 alias 不动）-> 可选删除旧物理索引。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| NewMigrator | `NewMigrator(client *elastic.Client, embedder retrieval.Embedder) *Migrator` | 构造 |
| Migrate | `(*Migrator).Migrate(ctx, embeddingDim int, deleteOld bool) (*MigrateResult, error)` | 执行迁移 |
| NextVersionName | `elastic.NextVersionName(alias, current) (string, error)` | 版本后缀递增（纯函数） |

MigrateResult：DocumentsOld/New、ChunksOld/New（物理索引名）、DocumentsMoved/ChunksMoved（迁移条数）。

### 入参要求
索引须已初始化（rove index init），否则返回 index.migrate 错误；embeddingDim 须与 Embedder.Dim() 一致（新 chunks 索引向量维度）。

### 返回
- 成功：迁移摘要（旧新索引名与条数）；
- 失败：机器可读错误；计数校验失败时不切换 alias（旧索引继续服务）。
