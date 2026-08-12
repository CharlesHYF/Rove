# 证据输出（evidence）
> 模块职责：将排序结果转为 Agent 可消费的可追溯证据（规格书 §3 / §09）。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/evidence、pkg/retrieval

## 证据构建
### 功能描述
将检索结果转换为 Evidence（document/chunk/url/title/text/score/rank_features/source），分数按最高分归一化到 0-1；不生成业务结论。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| New | `New() *Engine` | - |
| Build | `(*Engine).Build(ctx, result *retrieval.SearchResult) ([]Evidence, error)` | 空结果返回空列表 |

### 入参要求
result.Hits 非空；Score 可为任意实数（归一化处理）。

### 返回
[]Evidence；Score 在 0-1 之间，RankFeatures 携带分数分解与排序特征。
