# 检索与排序（retrieval + ranking）
> 模块职责：BM25 召回 -> document 聚合 -> heuristic 排序（规格书 §4.2 / §5.4）。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/retrieval、pkg/ranking、pkg/elastic

## 检索
### 功能描述
chunk 索引 multi_match（title^3 > heading_path^2 > content）召回候选（TopK x 5），按 document_id 聚合（保留最高 chunk 分），mget 回填文档信息，叠加 freshness/domain_quality/language 特征排序。

> M3 起：配置 Embedder（默认 PseudoEmbedder）后启用向量腿，BM25 + KNN 经 RRF Fusion（rank_constant=60）合并，
> Scores 分解为 lexical / vector / fusion / rank + heuristic 特征（规格书 §5.1）。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| New | `New(es *elastic.Client) *Retriever` | - |
| Search | `(*Retriever).Search(ctx, q *Query) (*SearchResult, error)` | 完整检索流 |

### 入参要求
Query.Text 非空；TopK 缺省 10；Filters 可选（domain/language/时间范围）。

### 返回
SearchResult{Hits（含 Chunk/Document/Score/Scores 分解）, Timings, TraceID}；权重默认值见规格书 §5.4。
