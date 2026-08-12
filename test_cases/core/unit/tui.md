<!--
	文件作用：tui 模块单元测试用例
	创建日期：2026-08-12
-->
# tui 模块（core/tui）-- 单元测试用例

> 模块简介：五视图模型纯逻辑
> 测试环境：dev · 测试框架：Go testing + testify

## 功能一：模型逻辑

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-TUI-001 | P0 | Tab 切换视图 | - | KeyTab | tab=1，视图含 SEARCH 头 | 与预期一致 | Pass | TestModelTabSwitch |
| TC-TUI-002 | P0 | 查询输入 | - | 输入 "browser" | searchQuery=browser | 与预期一致 | Pass | TestModelSearchInput |
| TC-TUI-003 | P0 | 五视图渲染 | - | View() | 含 SEARCH/CRAWL/BROWSE/INDEX/RUNTIME | 与预期一致 | Pass | TestModelViewSections |
| TC-TUI-004 | P0 | 结果渲染 | - | 注入 searchResult | 含 query 与标题 | 与预期一致 | Pass | TestModelSearchResultRendering |

# eval 模块（core/eval）-- 单元测试用例

> 模块简介：检索评估指标纯函数
> 测试环境：dev · 测试框架：Go testing + testify

## 功能一：评估指标

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-EVAL-001 | P0 | Recall@K | - | 3 相关文档 + 各类排序 | recall@2=2/3、无命中=0 | 与预期一致 | Pass | TestRecallAtK |
| TC-EVAL-002 | P0 | NDCG@K | - | 完美/部分排序 | 完美=1、部分<1 | 与预期一致 | Pass | TestNDCGAtK |
| TC-EVAL-003 | P0 | MRR | - | 首相关在第 2 位/无相关 | 0.5 / 0 | 与预期一致 | Pass | TestMRR |
