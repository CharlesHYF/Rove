<!--
	文件作用：router 模块单元测试用例
	创建日期：2026-08-15
-->
# router 模块（core/router）-- 单元测试用例

> 模块简介：Query Router 垂类推断与过滤（确定性，无 LLM）
> 测试环境：dev · 测试框架：Go testing + testify

## 功能一：垂类推断

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-RTR-001 | P0 | 代码垂类 | - | func/pip install/github 片段 | code（且覆盖 docs 标记，优先级更高） | 与预期一致 | Pass | TestRouteVerticalCode |
| TC-RTR-002 | P0 | 学术垂类 | - | arxiv/doi/citation | academic | 与预期一致 | Pass | TestRouteVerticalAcademic |
| TC-RTR-003 | P0 | 文档垂类 | - | how to/usage/guide | docs | 与预期一致 | Pass | TestRouteVerticalDocs |
| TC-RTR-004 | P0 | 默认 Web | - | 普通问题 | web | 与预期一致 | Pass | TestRouteVerticalWeb |
| TC-RTR-005 | P0 | 参数校验 | - | auto/WEB/image | auto 与大小写正常、非法值报错 | 与预期一致 | Pass | TestParseVertical |
| TC-RTR-006 | P0 | source_type 过滤 | - | buildSearchBody(SourceType=docs) | filter 含 source_type term | 与预期一致 | Pass | TestBuildSearchBodySourceTypeFilter |
