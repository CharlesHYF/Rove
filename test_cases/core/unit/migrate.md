<!--
	文件作用：migrate 模块单元测试用例
	创建日期：2026-08-15
-->
# migrate 模块（core/migrate）-- 单元测试用例

> 模块简介：索引迁移（schema 演进工具）
> 测试环境：dev（集成用例需 ES，不可达跳过）· 测试框架：Go testing + testify

## 功能一：版本命名（纯函数）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-MIG-001 | P0 | 版本递增 | - | v001 / v009 | v002 / v010（补零进位） | 与预期一致 | Pass | TestNextVersionName |
| TC-MIG-002 | P0 | 非法版本报错 | - | 无后缀 / vabc | 返回错误 | 与预期一致 | Pass | TestNextVersionNameInvalid |

## 功能二：迁移链路（ES 集成，不可达跳过）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-MIG-003 | P0 | 全链路迁移 | 已 init + 1 文档 | Migrate(256, false) | v001->v002、计数 1、alias 切换、检索可命中 | 与预期一致 | Pass | TestMigrateEndToEnd |
| TC-MIG-004 | P0 | 删旧 | 二次迁移 | Migrate(256, true) | v003 生成，旧物理索引已删除（reindex 404） | 与预期一致 | Pass | TestMigrateEndToEnd |
| TC-MIG-005 | P0 | 未初始化报错 | 未 init 的空前缀 | Migrate | 返回 index.migrate 错误 | 与预期一致 | Pass | TestMigrateRequiresExistingIndexes |
| TC-MIG-006 | P0 | 物理索引生命周期 | 创建 v001/v002 | Create/Reindex/Switch/Delete | alias 指向 v002，删除成功 | 与预期一致 | Pass | TestPhysicalIndexLifecycle |
