<!--
	文件作用：cli 模块单元测试用例，覆盖 fetch 命令的 JSON 与人类可读输出
	创建日期：2026-08-12
-->
# cli 模块（core/cli）-- 单元测试用例

> 模块简介：rove fetch 命令（装配 Fetcher + Pipeline，输出 Document JSON 或摘要）
> 测试环境：开发环境（dev）
> 测试框架：Go testing + testify + httptest

## 功能一：fetch 命令（rove fetch）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-CLI-001 | P0 | JSON 输出 | httptest HTML 页面、ROVE_ALLOW_PRIVATE=true | rove fetch <url> --json | 输出含 "Title": "CLI Test Page" 与 "Chunks" | 与预期一致 | Pass | 对应 TestFetchCommandJSON |
| TC-CLI-002 | P0 | 人类可读摘要 | 同上 | rove fetch <url> | 输出 title/canonical/language/source/chunks 摘要 | 与预期一致 | Pass | 对应 TestFetchCommandHuman |
