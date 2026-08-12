# 索引器（index）
> 模块职责：Document/Chunk 到 ES 的幂等写入、删除与状态汇总（NFR-007）。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/index、pkg/elastic

## 写入/删除/状态
### 功能描述
先清旧 chunks 再 bulk 写入 doc + chunks（refresh=wait_for），保证幂等；重复写入不产生重复 chunk。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| New | `New(es *elastic.Client) *Indexer` | - |
| Index | `(*Indexer).Index(ctx, doc) error` | 单文档幂等 |
| IndexBulk | `(*Indexer).IndexBulk(ctx, docs []*document.Document) error` | 批量 |
| Delete | `(*Indexer).Delete(ctx, documentID string) error` | 删文档 + chunks |
| Status | `(*Indexer).Status(ctx) (*Status, error)` | 规模与健康 |

### 入参要求
doc.ID / chunk.ChunkID 非空；索引须先 EnsureIndexes。

### 返回
Status{Health, Documents, Chunks, SizeBytes}；错误 code：index.bulk / index.delete / index.status。
