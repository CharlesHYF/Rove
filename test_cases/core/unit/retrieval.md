<!--
	文件作用：retrieval 模块单元测试用例
	创建日期：2026-08-12
-->
# retrieval 模块（core/retrieval）-- 单元测试用例

> 模块简介：BM25 召回 + 聚合 + heuristic 排序
> 测试环境：dev · 测试框架：Go testing + testify + 真实 ES

## 功能一：查询体构造

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-RET-001 | P0 | 查询体结构 | - | browser agent + 域/语言/时间过滤 | multi_match 字段 boost 正确、filter 3 条 | 与预期一致 | Pass | TestBuildSearchBody |
| TC-RET-002 | P0 | 无过滤查询体 | - | 仅 Text | 无 filter 子句 | 与预期一致 | Pass | TestBuildSearchBodyNoFilters |

## 功能二：排序特征（ranking）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-RET-003 | P0 | 新鲜度衰减 | - | 1h vs 400d 文档 | 新文档分数更高 | 与预期一致 | Pass | TestScoreFreshness |
| TC-RET-004 | P0 | 语言匹配 | - | zh 查询 vs en/zh 文档 | 匹配者分数更高 | 与预期一致 | Pass | TestScoreLanguageMatch |
| TC-RET-005 | P0 | 特征分解 | - | 配置 domain 质量 2.0 | features 含 4 键、quality=2.0、dup=0 | 与预期一致 | Pass | TestScoreFeaturesMap |

## 功能三：端到端检索

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-RET-006 | P0 | BM25 召回 | ES 已启动 | 索引 2 文档后搜 browser agent | 仅命中 d1，分数/特征/TraceID 正常 | 与预期一致 | Pass | TestSearchBM25 |
| TC-RET-007 | P0 | 域名过滤 | ES 已启动 | 搜 shared + domain=a.com | 仅命中 da | 与预期一致 | Pass | TestSearchDomainFilter |
