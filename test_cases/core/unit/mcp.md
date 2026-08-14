<!--
	文件作用：mcp 模块单元测试用例
	创建日期：2026-08-15
-->
# mcp 模块（core/mcp）-- 单元测试用例

> 模块简介：MCP Server 工具注册与协议流（官方 go-sdk 内存传输）
> 测试环境：dev · 测试框架：Go testing + testify

## 功能一：MCP 协议

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-MCP-001 | P0 | 工具列表 | 四个 fake service | tools/list | 返回 rove_search/fetch/browse/crawl 四个工具 | 与预期一致 | Pass | TestServerListsFourTools |
| TC-MCP-002 | P0 | 检索工具调用 | fake 检索结果 | tools/call rove_search | JSON 文本含标题与 trace | 与预期一致 | Pass | TestCallSearchTool |
| TC-MCP-003 | P0 | 空查询报错 | - | query 为空白 | IsError=true | 与预期一致 | Pass | TestCallSearchToolEmptyQueryIsError |
| TC-MCP-004 | P0 | 后端错误转义 | fake 返回错误 | tools/call rove_search | IsError=true | 与预期一致 | Pass | TestCallSearchToolBackendErrorIsError |
| TC-MCP-005 | P0 | 抓取工具调用 | fake Document | tools/call rove_fetch | JSON 含标题与正文 | 与预期一致 | Pass | TestCallFetchTool |
| TC-MCP-006 | P0 | 非法模式报错 | - | mode=ftp | IsError=true | 与预期一致 | Pass | TestCallFetchToolInvalidModeIsError |
| TC-MCP-007 | P0 | 浏览工具调用 | fake PageState | tools/call rove_browse | JSON 含标题与文本 | 与预期一致 | Pass | TestCallBrowseTool |
| TC-MCP-008 | P0 | 抓取站点工具调用 | fake Stats | tools/call rove_crawl | JSON 含 Visited/Indexed | 与预期一致 | Pass | TestCallCrawlTool |
| TC-MCP-009 | P0 | 空依赖拒绝 | Deps 全空 | NewServer | panic 明确报错 | 与预期一致 | Pass | TestNewServerRejectsNilDeps |
