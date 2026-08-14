<!--
	文件作用：reranker 模块单元测试用例
	创建日期：2026-08-15
-->
# reranker 模块（core/reranker）-- 单元测试用例

> 模块简介：Reranker 确定性启发式重排序
> 测试环境：dev · 测试框架：Go testing + testify

## 功能一：启发式重排序

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-RRK-001 | P0 | 标题命中加分 | 两条同分 hit，一条标题含查询词 | Rerank | 标题命中者排第一且 rerank 分更高 | 与预期一致 | Pass | TestHeuristicRerankerTitleBoost |
| TC-RRK-002 | P0 | 词面重叠排序 | 两条同分 hit 重叠度不同 | Rerank | 重叠高者排第一，rerank=0.15 | 与预期一致 | Pass | TestHeuristicRerankerOverlapOnly |
| TC-RRK-003 | P0 | 空输入兜底 | 空查询/空 hits | Rerank | 原样返回不报错 | 与预期一致 | Pass | TestHeuristicRerankerEmptyInputs |
| TC-RRK-004 | P0 | nil Scores 初始化 | hit.Scores=nil | Rerank | Scores 非 nil 且含 rerank | 与预期一致 | Pass | TestHeuristicRerankerNilScoresInitialized |
