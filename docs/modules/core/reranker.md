# Reranker（pkg/retrieval）
> 模块职责：Top-N 重排序 -- 确定性启发式（词面重叠 + 标题/标题路径命中加分，无模型）与可替换接口（PRD FR-RNK-002，P1）。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/retrieval、cmd/rove

## Reranker
### 功能描述
`rove search --rerank` 启用启发式重排序：对 Ranker 输出的 Top-N，按查询词与 chunk 正文的词面重叠比例（上限 0.15）、标题命中（0.10）、标题路径命中（0.05）加分后重排，重排分数写入 Scores["rerank"]（JSON 分数分解可见）。Reranker 为可替换接口（cross-encoder/LLM/custom 留待后续接入），默认关闭（nil 即跳过，历史行为不变）。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| Reranker | `Rerank(ctx, q *Query, hits []SearchHit) ([]SearchHit, error)` | 重排序接口 |
| HeuristicReranker | `(HeuristicReranker).Rerank(...)` | 确定性启发式实现 |
| Search | `(*Retriever).Search(...)` | Reranker 非 nil 时自动应用并记录 rerank 耗时 |

### 入参要求
查询文本非空；hits 可为空（原样返回）。

### 返回
按最终分数降序的 hits；每条 Scores 增加 rerank 特征；Timings 增加 rerank 阶段耗时。
