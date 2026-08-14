<!--
	文件作用：tui 模块单元测试用例
	创建日期：2026-08-12
	修改日期：2026-08-15
-->
# tui 模块（core/tui）-- 单元测试用例

> 模块简介：对话视图 + 五视图模型纯逻辑
> 测试环境：dev · 测试框架：Go testing + testify

## 功能一：对话视图

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-TUI-010 | P0 | 默认进入对话视图 | - | New() | tab=0，视图含"对话" | 与预期一致 | Pass | TestModelDefaultTabIsChat |
| TC-TUI-011 | P0 | 对话输入 | - | 输入"浏览器" | chatInput=浏览器 | 与预期一致 | Pass | TestModelChatInputTyping |
| TC-TUI-012 | P0 | 输入上限 | chatInput 已满 200 字符 | 继续输入 | 输入被忽略 | 与预期一致 | Pass | TestModelChatInputLimit |
| TC-TUI-013 | P0 | 回车提交问题 | 已输入问题 | KeyEnter | 追加用户消息、清空输入、进入检索中 | 与预期一致 | Pass | TestModelChatEnterSubmits |
| TC-TUI-014 | P0 | 空输入回车忽略 | 输入为空 | KeyEnter | 无消息、无异步命令 | 与预期一致 | Pass | TestModelChatEnterEmptyIgnored |
| TC-TUI-015 | P0 | 清空会话 | 已有两条消息 | KeyCtrlL | 消息清空、翻页归零 | 与预期一致 | Pass | TestModelChatCtrlLClears |
| TC-TUI-016 | P0 | 历史输入翻阅 | 两条历史 | KeyUp/KeyDown 组合 | 依次显示新旧输入，越界回空白 | 与预期一致 | Pass | TestModelChatHistoryNavigation |
| TC-TUI-017 | P1 | 会话翻页 | - | KeyPgUp/KeyPgDown | chatScroll 增减，顶部不越界 | 与预期一致 | Pass | TestModelChatScrollKeys |
| TC-TUI-018 | P1 | 展开分数分解 | 最新为回答消息 | KeyCtrlE 两次 | 展开后折叠 | 与预期一致 | Pass | TestModelChatExpandToggle |
| TC-TUI-019 | P0 | 异步回答追加 | chatLoading=true | chatAnswerMsg | 回答入列、loading 复位 | 与预期一致 | Pass | TestModelChatAnswerAppends |
| TC-TUI-020 | P0 | 回答卡片渲染 | 注入消息与来源 | View() | 含问题/回答/来源/相关度 | 与预期一致 | Pass | TestModelChatRendering |
| TC-TUI-021 | P0 | 空结果回答 | 无 Evidence | buildChatAnswer | 提示"没有找到相关内容" | 与预期一致 | Pass | TestBuildChatAnswerEmpty |
| TC-TUI-022 | P0 | 来源与分数构造 | 4 条 Evidence + 分数 | buildChatAnswer | 来源截 3 条、分数/耗时正确 | 与预期一致 | Pass | TestBuildChatAnswerSourcesAndScores |
| TC-TUI-023 | P1 | 长回答截断 | 12 行 / 400 字符文本 | buildChatAnswer | 行数与字符数受限并带省略号 | 与预期一致 | Pass | TestBuildChatAnswerTruncatesLongAnswer / TruncatesLongChars |
| TC-TUI-024 | P0 | 错误兜底 | 构造错误 | buildChatError | 友好提示 + 原始错误 | 与预期一致 | Pass | TestBuildChatError |

## 功能二：模型逻辑

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-TUI-001 | P0 | Tab 切换视图 | - | KeyTab | tab=1，视图含"搜索"头 | 与预期一致 | Pass | TestModelTabSwitch |
| TC-TUI-002 | P0 | 查询输入 | tab=搜索视图 | 输入 "browser" | searchQuery=browser | 与预期一致 | Pass | TestModelSearchInput |
| TC-TUI-003 | P0 | 六视图渲染 | - | View() | 含对话/搜索/抓取/浏览/索引/运行时 | 与预期一致 | Pass | TestModelViewSections |
| TC-TUI-004 | P0 | 结果渲染 | tab=搜索视图，注入 searchResult | View() | 含 query 与标题 | 与预期一致 | Pass | TestModelSearchResultRendering |

# eval 模块（core/eval）-- 单元测试用例

> 模块简介：检索评估指标纯函数
> 测试环境：dev · 测试框架：Go testing + testify

## 功能一：评估指标

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-EVAL-001 | P0 | Recall@K | - | 3 相关文档 + 各类排序 | recall@2=2/3、无命中=0 | 与预期一致 | Pass | TestRecallAtK |
| TC-EVAL-002 | P0 | NDCG@K | - | 完美/部分排序 | 完美=1、部分<1 | 与预期一致 | Pass | TestNDCGAtK |
| TC-EVAL-003 | P0 | MRR | - | 首相关在第 2 位/无相关 | 0.5 / 0 | 与预期一致 | Pass | TestMRR |
