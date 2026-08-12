<!--
	文件作用：fetch 模块单元测试用例，覆盖 redirect / 重试 / body 上限 / SSRF / charset / escalation 全部功能
	创建日期：2026-08-12
-->
# fetch 模块（core/fetch）-- 单元测试用例

> 模块简介：HTTP First 抓取（redirect 跟随、超时重试、charset 转换、body 上限、SSRF 策略、escalation 信号）
> 测试环境：开发环境（dev）
> 测试框架：Go testing + testify + httptest

## 功能一：HTTP 抓取（Fetch）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-FET-001 | P0 | 跟随重定向 | httptest 双路径服务器 | /start -> 301 -> /end | 最终 URL 为 /end、StatusCode 200、Timings.Total > 0 | 与预期一致 | Pass | 对应 TestFetchFollowsRedirect |
| TC-FET-002 | P0 | 重定向环上限 | MaxRedirects=5，服务器无限跳转 | /loop 无限 302 | 返回含 "redirect" 的错误（不可重试） | 与预期一致 | Pass | 对应 TestFetchMaxRedirects |
| TC-FET-003 | P0 | body 超限 | MaxBodyBytes=1024，响应 2048 字节 | 超限响应 | 返回 fetch.body_too_large | 与预期一致 | Pass | 对应 TestFetchMaxBody |
| TC-FET-004 | P0 | SSRF 阻止 loopback | AllowPrivate=false（默认） | 127.0.0.1 地址 | 返回 fetch.ssrf_blocked（含 "blocked"） | 与预期一致 | Pass | 对应 TestFetchSSRFBlockLoopback |
| TC-FET-005 | P0 | GBK 转 UTF-8 | charset=gbk 响应 | "你好" 的 GBK 字节 | Body 转换为 UTF-8 "你好" | 与预期一致 | Pass | 对应 TestFetchCharsetGBK |
| TC-FET-006 | P0 | auto 模式内容不足 | 短 HTML + script | 不足正文 | 返回 ErrEscalationRequired（errors.Is 可判） | 与预期一致 | Pass | 对应 TestFetchAutoInsufficient |
| TC-FET-007 | P1 | browser 模式未实现 | - | Mode=browser | 返回 browser.not_implemented（含 "M5"） | 与预期一致 | Pass | 对应 TestFetchModeBrowserNotImplemented |
| TC-FET-008 | P1 | 网络错误自动重试 | 前 2 次连接被重置 | 正常请求 | 第 3 次成功，calls >= 3 | 与预期一致 | Pass | 对应 TestFetchRetriesNetworkError |
