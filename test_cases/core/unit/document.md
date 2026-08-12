<!--
	文件作用：document 模块单元测试用例，覆盖幂等 ID 生成的全部功能
	创建日期：2026-08-12
-->
# document 模块（core/document）-- 单元测试用例

> 模块简介：核心数据模型与幂等 ID 生成（RawDocument / Document / Chunk / NewDocumentID / ChunkID）
> 测试环境：开发环境（dev）
> 测试框架：Go testing + testify

## 功能一：文档 ID 生成（NewDocumentID）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-DOC-001 | P0 | 同 URL 幂等 | - | 同一 canonical URL 调用两次 | 两次 ID 一致，长度为 16 hex | 与预期一致 | Pass | 对应 TestNewDocumentIDDeterministic |
| TC-DOC-002 | P0 | 不同 URL 不同 ID | - | 两个不同 canonical URL | 生成不同 ID | 与预期一致 | Pass | 同上 |

## 功能二：Chunk ID 生成（ChunkID）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-DOC-003 | P0 | 拼接格式 | - | documentID="abc123", position=0/12 | 输出 "abc123-0" / "abc123-12" | 与预期一致 | Pass | 对应 TestChunkID |
