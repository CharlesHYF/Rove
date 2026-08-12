<!--
	文件作用：elastic 模块单元测试用例
	创建日期：2026-08-12
-->
# elastic 模块（core/elastic）-- 单元测试用例

> 模块简介：ES 客户端与索引 alias 管理
> 测试环境：dev · 测试框架：Go testing + testify + 真实 ES（ROVE_TEST_ES_URL）

## 功能一：客户端与索引初始化

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-EL-001 | P0 | 非法 URL 报错 | - | "://bad" | 返回错误 | 与预期一致 | Pass | TestNewClientInvalidURL |
| TC-EL-002 | P0 | Ping 连通性 | ES 已启动 | - | 无错误 | 与预期一致 | Pass | TestPing（不可达则 Skip） |
| TC-EL-003 | P0 | alias 命名 | - | prefix=rove-x | rove-x-documents / rove-x-chunks | 与预期一致 | Pass | TestAliasNames |
| TC-EL-004 | P0 | 幂等建索引 | ES 已启动 | EnsureIndexes(256) x2 | 首次创建、重复调用不报错、alias 存在 | 与预期一致 | Pass | TestEnsureIndexes |
