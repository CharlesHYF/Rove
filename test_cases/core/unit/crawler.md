<!--
	文件作用：crawler 模块单元测试用例
	创建日期：2026-08-12
-->
# crawler 模块（core/crawler）-- 单元测试用例

> 模块简介：frontier / 调度 / robots / 链接策略 / 编排 / PDF
> 测试环境：dev · 测试框架：Go testing + testify

## 功能一：frontier 与调度

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-CRW-001 | P0 | 入队与计数 | - | 2 条记录 | queued=2 | 与预期一致 | Pass | TestFrontierEnqueueAndCount |
| TC-CRW-002 | P0 | URL 去重 | - | 重复入队 | 第二次 added=0 | 与预期一致 | Pass | TestFrontierDedup |
| TC-CRW-003 | P0 | 状态流转 | - | 入队后置 done | done=1 | 与预期一致 | Pass | TestFrontierUpdateState |
| TC-CRW-004 | P0 | 调度取号与完成 | - | queued 1 条 | Next 取到并置 fetching，Complete 后 done | 与预期一致 | Pass | TestSchedulerNextAndComplete |
| TC-CRW-005 | P0 | 失败退避 | MaxRetries=1 | 连续失败 | 第 2 次失败后 state=failed | 与预期一致 | Pass | TestSchedulerFailureBackoff |
| TC-CRW-006 | P0 | 退避公式 | - | attempt 0/1/3 | 1s/2s/8s | 与预期一致 | Pass | TestRetryBackoff |

## 功能二：robots 与链接策略

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-CRW-007 | P0 | robots 放行/拒绝 | httptest robots.txt | /public 与 /private | 前者允许、后者拒绝、crawl-delay=2s | 与预期一致 | Pass | TestRobotsIsAllowed |
| TC-CRW-008 | P0 | robots 缺失 fail-open | 404 | 任意路径 | 允许 | 与预期一致 | Pass | TestRobotsMissingFile |
| TC-CRW-009 | P0 | 链接过滤与规范化 | - | 跨域/js/片段/utm/自链接 | 仅同 host 规范化链接入队 | 与预期一致 | Pass | TestCandidateLinks |
| TC-CRW-010 | P1 | 域名策略 | - | allowlist 与非白名单 | 正确放行/拒绝 | 与预期一致 | Pass | TestPolicyAllowedDomain |
| TC-CRW-011 | P1 | URL 规范化 | - | 大小写/utm/片段 | scheme/host 小写、tracking 剥离 | 与预期一致 | Pass | TestNormalizeURL |

## 功能三：编排与 PDF

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-CRW-012 | P0 | 站点图抓取 | httptest 3 页 | seed=/ | visited>=3、indexed>=3 | 与预期一致 | Pass | TestCrawlerVisitsSite |
| TC-CRW-013 | P0 | PDF 解析 | 程序化 PDF | Hello Rove 文本 | 提取成功 | 与预期一致 | Pass | TestPDFParser |
| TC-CRW-014 | P1 | PDF 非法输入 | - | 非 PDF 字节 | 返回 pdf 错误 | 与预期一致 | Pass | TestPDFParserInvalid |
| TC-CRW-015 | P0 | CLI crawl 端到端 | ES 已启动 | seed -> crawl -> search | visited/indexed 输出、搜索命中 Crawl Home | 与预期一致 | Pass | TestCrawlCommand |
| TC-CRW-016 | P0 | E2E 验收 A1 | ES 已启动 | 3 页站点图 crawl | visited>=3、搜索命中 E2E Home | 与预期一致 | Pass | TestCrawlEndToEnd |
