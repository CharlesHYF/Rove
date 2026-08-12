<!--
	文件作用：browser 模块单元测试用例
	创建日期：2026-08-12
-->
# browser 模块（core/browser）-- 单元测试用例

> 模块简介：浏览器渲染 / 页面状态 / 升级接线
> 测试环境：dev · 测试框架：Go testing + testify + 真实 Chromium（无浏览器时 Skip）

## 功能一：渲染与页面状态

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-BRW-001 | P0 | SPA 渲染抓取 | 浏览器可用 | JS 渲染页面 | RawDocument.FetchMethod=browser、Body 含渲染内容 | 与预期一致 | Pass | TestBrowserFetchesRenderedHTML |
| TC-BRW-002 | P0 | 稳定 element id | 浏览器可用 | 同一表单页两次加载 | 交互元素 id 两次一致；链接/标题/文本完整 | 与预期一致 | Pass | TestPageStateStableIDs |
| TC-BRW-003 | P0 | CLI browser 模式 | 浏览器可用 | fetch --mode browser | 输出渲染后 Dynamic Title | 与预期一致 | Pass | TestFetchBrowserMode |
| TC-BRW-004 | P0 | auto 自动升级 | 浏览器可用 | fetch --mode auto（SPA shell） | 自动升级浏览器，输出渲染内容（A2） | 与预期一致 | Pass | TestFetchAutoEscalation |
| TC-BRW-005 | P0 | E2E 同一管线 | 浏览器+ES 可用 | SPA -> browser fetch -> pipeline -> index -> search | 渲染内容可检索，标题 SPA E2E | 与预期一致 | Pass | TestBrowserEscalationPipeline |
