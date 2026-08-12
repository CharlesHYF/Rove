# 内容管线（content）
> 模块职责：把 RawDocument 转换为标准 Document -- 解析（Parser 注册表）、规范化（Canonicalizer）、去重（Deduper）、分块（Chunker）与管线编排。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/content、pkg/fetch、pkg/document、pkg/index(M2)

## Parser（解析）
### 功能描述
按 Content-Type 路由解析：text/html（元数据 + readability 正文 + Markdown 视图）、text/plain、application/json（保留结构 + 可读视图）、text/markdown；未知类型返回 content.unsupported_type。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| DefaultRegistry | `DefaultRegistry() *Registry` | 注册 html/text/json/markdown（pdf 在 M4 注册） |
| For | `(*Registry).For(mimeType string) (Parser, error)` | 忽略 charset 参数 |
| Parse | `Parse(ctx, raw) (*ParsedContent, error)` | 各 Parser 实现 |

### 入参要求
raw.ContentType 必填；raw.Body 为 UTF-8。

### 返回
*ParsedContent（Title/Description/Content/Markdown/CanonicalURL/Links/Language/PublishedAt/Metadata）；错误 content.html_parse / content.json_parse / content.unsupported_type。

## Canonicalize（URL 规范化）
### 功能描述
归并 URL：scheme/host 小写、去 default port、去 fragment、去 tracking 参数（utm_*/fbclid/gclid 等）、trailing slash 归并；rel=canonical 相对时按 rawURL 解析。

### 接口信息
| 函数 | 签名 |
| --- | --- |
| Canonicalize | `(*Canonicalizer).Canonicalize(ctx, rawURL, relCanonical string) (string, error)` |

### 入参要求
rawURL 合法；relCanonical 可为空。

### 返回
规范化 URL；错误 content.invalid_url。

## Deduper（四层去重）
### 功能描述
canonical -> exact content hash -> near duplicate（SimHash，Hamming <= 3）三层检查；URL 层由 frontier 负责（M4）。非重复文档登记进状态。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| NewDeduper | `NewDeduper() *Deduper` | 内存态（M1） |
| IsDuplicate | `(*Deduper).IsDuplicate(ctx, doc) (bool, string, error)` | reason: canonical_duplicate / exact_content_duplicate / near_duplicate |

### 入参要求
doc.CanonicalURL、doc.ContentHash、doc.Content 非空。

### 返回
(bool, reason)；重复时调用方跳过索引（M2/M4 使用）。

## Chunker（分块）
### 功能描述
以 Markdown 为输入按语义结构切分：标题为硬边界，维护 heading stack 生成 heading_path 与 position；超长块按 MaxTokens 词窗口切分；相邻 chunk 保留 Overlap。

### 接口信息
| 函数 | 签名 |
| --- | --- |
| NewChunker | `NewChunker(maxTokens, overlap int) *Chunker` |
| Chunk | `(*Chunker).Chunk(ctx, doc) ([]document.Chunk, error)` |

### 入参要求
doc.Markdown 非空；maxTokens > 0；overlap >= 0。

### 返回
[]document.Chunk（ChunkID = `<docID>-<position>`）。

## Pipeline（管线编排）
### 功能描述
解析 -> 规范化 -> 去重 -> 分块；去重命中时返回 IsDuplicate=true 且 Chunks 为空。

### 接口信息
| 函数 | 签名 |
| --- | --- |
| NewPipeline | `NewPipeline(registry, canonicalizer, deduper, chunker) *Pipeline` |
| Process | `(*Pipeline).Process(ctx, raw, sourceType) (*document.Document, *DedupResult, error)` |
| HashContent | `HashContent(s string) string`（sha256 hex） |
| Hostname | `Hostname(rawURL string) string`（失败返回 "unknown"） |

### 入参要求
raw.ContentType 可路由；sourceType 默认 "web"。

### 返回
(*document.Document, *DedupResult)；错误透传 Parser/Canonicalizer。
