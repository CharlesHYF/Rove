# 核心数据模型（document）
> 模块职责：定义 Rove 全链路的统一数据模型 -- RawDocument / Document / Chunk / Source / FetchTimings，并生成幂等 ID。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/document、pkg/fetch、pkg/content、pkg/index(M2)

## RawDocument（抓取原始输出）
### 功能描述
Fetcher 的输出结构，保留响应原始信息（状态码、Headers、Content-Type、正文、抓取方式与阶段耗时），不做业务级正文解释。

### 接口信息
| 项 | 值 |
| --- | --- |
| 类型 | Go struct |
| 路径 | pkg/document/document.go |

### 入参要求
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| ID | string | uuid（临时，不落索引） |
| URL | string | 最终 URL（含 redirect 后） |
| StatusCode | int | HTTP 状态码 |
| Headers | http.Header | 原始响应头 |
| ContentType | string | 响应 Content-Type |
| Body | []byte | 响应正文（UTF-8，已做 charset 转换） |
| FetchMethod | FetchMethod | http \| browser |
| Timings | FetchTimings | DNS/Connect/TLS/TTFB/Total 耗时 |
| FetchedAt | time.Time | 抓取时间（UTC） |

### 返回
无（纯数据载体）。

## Document（标准文档）
### 功能描述
内容管线输出的标准文档，是索引（M2）与检索（M3）的输入单元。

### 入参要求
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| ID | string | 确定性：sha256(canonicalURL) 前 16 字节 hex，幂等重抓按此覆盖 |
| CanonicalURL | string | rel=canonical 或规范化后的 URL |
| Content / Markdown | string | 正文文本 / 可读文本视图（chunk 切分的语义来源） |
| Source | Source | Domain + Type（web/docs/code/academic） |
| ContentHash | string | sha256(Content) |
| Chunks | []Chunk | 切分结果 |

### 返回
无（纯数据载体）。

## Chunk（检索与 Evidence 最小单元）
### 功能描述
检索召回与 Evidence 输出的最小单元，携带 heading_path 与 position 供回溯。

### 入参要求
| 字段 | 类型 | 说明 |
| --- | --- | --- |
| ChunkID | string | `<DocumentID>-<Position>`，作为 ES _id 保证幂等 |
| Position | int | 文档内序号 |
| HeadingPath | string | 如 "Installation > Quick Start" |
| TokenCount | int | 近似估算（单词数 + 非 ASCII 字符数/2） |
| Embedding | []float32 | 可选；M3 起由 Embedder 填充 |

### 返回
无（纯数据载体）。

## ID 生成函数
### 功能描述
为 Document 与 Chunk 生成确定性 ID，支撑 NFR-007 幂等写入。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| NewDocumentID | `NewDocumentID(canonicalURL string) string` | sha256 前 16 字节 hex |
| ChunkID | `ChunkID(documentID string, position int) string` | `<docID>-<position>` |

### 入参要求
canonicalURL 非空；position >= 0。

### 返回
16 位 hex 字符串 / 拼接字符串。
