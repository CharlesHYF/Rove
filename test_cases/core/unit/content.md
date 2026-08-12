<!--
	文件作用：content 模块单元测试用例，覆盖解析 / 规范化 / 去重 / 分块 / 管线编排全部功能
	创建日期：2026-08-12
-->
# content 模块（core/content）-- 单元测试用例

> 模块简介：内容管线（Parser 注册表、HTML/Text/JSON/Markdown 解析、Canonicalizer、Jaccard 去重、Chunker、Pipeline）
> 测试环境：开发环境（dev）
> 测试框架：Go testing + testify

## 功能一：Parser 注册表（Registry）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-CNT-001 | P0 | MIME 路由 | DefaultRegistry | text/html、text/html;charset、text/plain、application/json、text/markdown | 均能解析出 Parser | 与预期一致 | Pass | 对应 TestRegistryFor |
| TC-CNT-002 | P0 | 未知类型报错 | - | application/unknown | 返回 content.unsupported_type | 与预期一致 | Pass | 同上 |

## 功能二：解析器（Parser）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-CNT-003 | P0 | HTML 元数据与正文 | testdata/page.html | 完整 HTML 页面 | Title/Description/CanonicalURL/Language/PublishedAt/Links 全部提取，正文与 Markdown 非空 | 与预期一致 | Pass | 对应 TestHTMLParser |
| TC-CNT-004 | P0 | 纯文本解析 | - | "  hello\nworld  " | 首尾裁剪、换行归一为 "hello\nworld" | 与预期一致 | Pass | 对应 TestTextParser |
| TC-CNT-005 | P0 | JSON 解析 | - | {"title":"api doc","a":1} | Title 提取、Content 为美化 JSON | 与预期一致 | Pass | 对应 TestJSONParser |
| TC-CNT-006 | P0 | Markdown 解析 | - | "# Title\n\nsome body" | Title 提取、Content 含正文 | 与预期一致 | Pass | 对应 TestMarkdownParser |

## 功能三：URL 规范化（Canonicalizer）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-CNT-007 | P0 | 规范化归并（表驱动 8 场景） | - | scheme 大小写 / 默认端口 / fragment / trailing slash / root slash / tracking 参数 / rel=canonical 绝对与相对 | 全部输出预期 URL | 与预期一致 | Pass | 对应 TestCanonicalize 各子用例 |
| TC-CNT-008 | P1 | 非法 URL 报错 | - | "://bad" | 返回 content.invalid_url | 与预期一致 | Pass | 对应 TestCanonicalizeInvalid |

## 功能四：近重复与去重（Jaccard / Deduper）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-CNT-009 | P0 | Jaccard 阈值判定 | - | 1 词变异文本 / 无关文本 / 空集合 | 近重复 >= 0.8、无关 < 0.8、空集合为 0 | 与预期一致 | Pass | 对应 TestJaccardSimilarity |
| TC-CNT-010 | P0 | 三层去重层序 | - | 同 URL / 同内容不同 URL / 1 词变异 / 无关文本 | 依次命中 canonical / exact / near_duplicate，无关文本通过 | 与预期一致 | Pass | 对应 TestDeduperLayers |

## 功能五：分块（Chunker）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-CNT-011 | P0 | heading_path 与 position | - | 三级标题 Markdown | 每个标题段生成独立 chunk，HeadingPath/Position/ChunkID 正确 | 与预期一致 | Pass | 对应 TestChunkerHeadingPathAndPosition |
| TC-CNT-012 | P0 | 词窗口切分 | MaxTokens=300 | 500 词段落 x2 | 切为 >= 4 chunk，每块 <= 300 token | 与预期一致 | Pass | 对应 TestChunkerMaxTokens |
| TC-CNT-013 | P1 | overlap 语义 | MaxTokens=300, Overlap=80 | 1000 词文本 | 相邻 chunk 内容重叠且不同 | 与预期一致 | Pass | 对应 TestChunkerOverlap |

## 功能六：管线编排（Pipeline）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-CNT-014 | P0 | 完整转换 | page.html fixture | RawDocument(text/html) | 输出标准 Document（title/canonical/source/contenthash/chunks），chunk ID 正确 | 与预期一致 | Pass | 对应 TestPipelineProducesDocument |
| TC-CNT-015 | P0 | 去重透传 | - | 同 URL / 同内容不同 URL 二次处理 | 返回 canonical_duplicate / exact_content_duplicate | 与预期一致 | Pass | 对应 TestPipelineDuplicate |
| TC-CNT-016 | P1 | 不支持类型报错 | - | application/unknown | 返回 content.unsupported_type | 与预期一致 | Pass | 对应 TestPipelineUnsupportedContentType |
