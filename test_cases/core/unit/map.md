<!--
	文件作用：map 模块单元测试用例
	创建日期：2026-08-15
-->
# map 模块（core/map）-- 单元测试用例

> 模块简介：Map 链接发现服务（URL Discovery）
> 测试环境：dev · 测试框架：Go testing + testify

## 功能一：链接发现

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-MAP-001 | P0 | 站内/站外划分与去重 | httptest 页面（相对/绝对/重复链接） | Map(url) | 相对链接解析为绝对地址；站内 2 条（去重）、站外 1 条；canonical 来自 rel=canonical | 与预期一致 | Pass | TestMapDiscoversLinks |
| TC-MAP-002 | P0 | 不支持类型报错 | httptest 返回 octet-stream | Map(url) | 返回机器可读错误，不 silent failure | 与预期一致 | Pass | TestMapUnsupportedContent |
