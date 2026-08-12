<!--
	文件作用：index 模块单元测试用例
	创建日期：2026-08-12
-->
# index 模块（core/index）-- 单元测试用例

> 模块简介：幂等写入 / 删除 / 状态
> 测试环境：dev · 测试框架：Go testing + testify + 真实 ES

## 功能一：索引写入

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-IDX-001 | P0 | 单文档写入与状态 | ES 已启动 | fixtureDoc | documents=1, chunks=1, size>0 | 与预期一致 | Pass | TestIndexAndStatus |
| TC-IDX-002 | P0 | 重复写入幂等 | ES 已启动 | 同一 doc 写入两次 | 仍为 1 doc + 1 chunk | 与预期一致 | Pass | TestIndexIdempotent |
| TC-IDX-003 | P0 | 更新清旧 chunks | ES 已启动 | 追加 chunk 后重写 | chunks=2（无残留） | 与预期一致 | Pass | TestIndexUpdatesChunks |
| TC-IDX-004 | P0 | 删除文档 | ES 已启动 | Index 后 Delete | documents=0, chunks=0 | 与预期一致 | Pass | TestDelete |
| TC-IDX-005 | P1 | 批量写入 | ES 已启动 | IndexBulk | 写入成功 | 与预期一致 | Pass | TestIndexBulk |
