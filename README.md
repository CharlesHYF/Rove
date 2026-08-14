<h1 align="center">Rove</h1>
<div align="center">

![Go](https://img.shields.io/badge/Go-1.26-blue) ![Elasticsearch](https://img.shields.io/badge/Elasticsearch-8.13-yellow) ![MCP](https://img.shields.io/badge/MCP-stdio-blueviolet) ![License](https://img.shields.io/badge/license-Apache--2.0-blue) ![Version](https://img.shields.io/badge/version-0.1.0-lightgrey) ![CI](https://img.shields.io/github/actions/workflow/status/CharlesHYF/Rove/verify.yml)

</div>

## 目录
- [项目简介](#项目简介)
- [技术栈](#技术栈)
- [项目亮点](#项目亮点)
- [快速开始](#快速开始)
- [使用说明](#使用说明)
- [MCP 接入](#mcp-接入)
- [配置项](#配置项)
- [项目结构](#项目结构)
- [常用命令](#常用命令)
- [文档](#文档)
- [里程碑](#里程碑)
- [维护者](#维护者)
- [许可](#许可)

## 项目简介
Rove 是面向 AI Agent 的开源（Apache-2.0）自托管 Web 基础设施。它把 Crawl → Fetch → Parse → Index → Retrieve → Rank → Evidence 的完整链路装进一个 Go 单二进制，让任意 Agent 不依赖外部 Search SaaS 就能可靠地发现、访问、索引 Web 内容，并以可追溯的 Evidence 形式获得检索结果。核心主张：自己的索引、自己的证据、自己的账单。

## 技术栈
- 语言：Go（单二进制，go.mod 声明 1.26.5）
- 检索引擎：Elasticsearch 8.13（Docker Compose 本地化）
- 浏览器：Chromium + Chrome DevTools Protocol（chromedp）
- 本地状态：SQLite（纯 Go 驱动 modernc.org/sqlite，仅承担 Runtime 状态）
- 界面：Cobra CLI + Bubble Tea TUI + MCP Server（官方 modelcontextprotocol/go-sdk）
- 测试：Go testing + testify；交付闸门 `make verify`（规范校验 + 全量测试）

## 项目亮点
- 无 key 全闭环：内置确定性 PseudoEmbedder，没有任何 LLM API Key 也能完成 抓取 → 索引 → BM25+向量 Hybrid → Evidence 全流程
- 可追溯证据：每条结果携带 URL、来源、相关片段、分数分解（lexical/vector/fusion/rank）与阶段耗时，可直接被 Agent 引用
- HTTP First + Browser Escalation：静态页走轻量 HTTP，动态页自动升级浏览器渲染，进入同一条内容管线
- 对话式 TUI：启动即对话查询，回答卡片带来源与分数；Tab 切换六视图运行时调试器（对话/搜索/抓取/浏览/索引/运行时）
- 原生 MCP：stdio 一键接入 Claude Desktop / Cursor，暴露 search/fetch/browse/crawl 四个工具
- 确定性优先：Query Router（垂类推断）、启发式 Reranker、四层去重均无 LLM 依赖，可测试、可复现
- 索引零停机演进：`rove index migrate` 创建新版本索引、重嵌入向量、计数校验后原子切换 alias

## 快速开始
### 前置要求
- Go >= 1.26（与 go.mod 声明一致）
- Docker（运行 Elasticsearch）
- Chrome / Edge（浏览器渲染功能，可选但推荐）

### 本地运行
```bash
# 1) 构建
go build -o bin/rove ./cmd/rove

# 2) 启动 Elasticsearch（首次拉取镜像约 1.3GB，之后秒起）
docker compose up -d elasticsearch

# 3) 初始化索引
./bin/rove index init
```

### 最小演示
```bash
docker compose up -d elasticsearch
./bin/rove index init
./bin/rove crawl https://example.com --max-pages 5
./bin/rove search "example domain"
./bin/rove search "example domain" --json   # 查看 lexical/vector/fusion/rank 分数与 Evidence
./bin/rove                                  # 进入 TUI 对话查询
```

## 使用说明
| 命令 | 说明 |
| --- | --- |
| `rove` | 进入 TUI：默认对话式查询（回车检索、Ctrl+E 展开分数、Ctrl+L 清空），Tab 切换六视图调试器 |
| `rove fetch <url> [--mode auto\|http\|browser] [--json]` | 抓取单页，输出标准 Document（auto 模式动态页自动浏览器升级） |
| `rove crawl <seed> [--max-pages N] [--max-depth D]` | 抓取整个站点：发现链接、去重、索引（含向量） |
| `rove search <query> [--json] [--vertical auto\|web\|docs\|code\|academic] [--rerank]` | BM25 + 向量 Hybrid 检索，输出 Evidence 与分数分解（垂类路由默认自动推断，--rerank 启用启发式重排序） |
| `rove browse <url> [--json]` | 浏览器渲染页面状态（文本/链接/交互元素 + 稳定元素 id） |
| `rove map <url> [--json]` | 发现页面链接（站内/站外，仅 URL Discovery，不建索引） |
| `rove mcp` | 以 stdio 启动 MCP Server，供 Claude Desktop / Cursor 等 Agent 接入 |
| `rove index init\|status\|stats\|rebuild\|migrate [--delete-old]` | 索引生命周期管理（migrate 为 schema 演进：新版本索引 + 原子 alias 切换） |

## MCP 接入
`rove mcp` 以 stdio 启动 MCP Server，暴露 rove_search / rove_fetch / rove_browse / rove_crawl 四个工具。Claude Desktop / Cursor 的 mcpServers 配置示例：
```json
{
  "mcpServers": {
    "rove": {
      "command": "rove",
      "args": ["mcp"]
    }
  }
}
```

## 配置项
全部配置可用环境变量覆盖（`ROVE_` 前缀）：

| 变量 | 说明 | 默认 |
| --- | --- | --- |
| `ROVE_ES_URL` | Elasticsearch 地址 | http://localhost:9200 |
| `ROVE_INDEX_PREFIX` | 索引前缀（多实例隔离） | rove |
| `ROVE_INDEX_EMBEDDING_DIM` | 向量维度 | 256 |
| `ROVE_ALLOW_PRIVATE` | 允许抓取私网/loopback（开发用） | false |
| `ROVE_CRAWL_STATE_DB` | frontier SQLite 路径 | rove.db |
| `ROVE_CRAWL_MAX_PAGES` / `ROVE_CRAWL_MAX_DEPTH` | 抓取上限 | 1000 / 3 |
| `ROVE_CRAWL_WORKERS` | 抓取并发 | 2 |
| `ROVE_CRAWL_DELAY` | 每次抓取最小间隔（礼貌爬取，避免触发站点限速） | 500ms |
| `ROVE_BROWSER_EXECUTABLE` | 浏览器可执行文件路径（自动发现失败时指定） | 自动发现 |
| `ROVE_BROWSER_ENABLED` | 浏览器功能开关 | true |
| `ROVE_SEARCH_TOP_K` | 检索默认返回条数 | 10 |

## 项目结构
```
rove/
├─ cmd/rove/            # Cobra 入口（命令薄壳，只做装配）
├─ internal/
│  ├─ app/              # Application Service（Search/Fetch/Browse/Crawl/Map/Index）
│  ├─ config/           # YAML + env + flags 配置加载与校验
│  ├─ runtime/          # 事件总线与 Runtime Telemetry
│  └─ scheduler/        # Frontier 调度内部实现
├─ pkg/
│  ├─ crawler/          # 站点抓取、链接发现、robots 策略
│  ├─ fetch/            # HTTP Fetcher
│  ├─ browser/          # CDP Browser Runtime 与 Page State
│  ├─ content/          # Parser 注册表、抽取、规范化、去重、分块
│  ├─ document/         # 核心数据模型（Document/Chunk/RawDocument）
│  ├─ elastic/          # ES 客户端、mapping、alias 生命周期与迁移操作
│  ├─ index/            # 索引写入、幂等 bulk 与 schema 迁移
│  ├─ retrieval/        # BM25/向量/Hybrid Fusion、Query Router、Reranker
│  ├─ ranking/          # 启发式 Ranker（freshness/domain_quality/language）
│  ├─ evidence/         # Evidence Engine
│  └─ rove/             # machine-readable 错误模型
├─ protocol/            # JSON 契约与 MCP Server（stdio）
├─ tui/                 # Bubble Tea 六视图（对话首页 + 五调试视图）
├─ tests/               # e2e 与验收测试（A1-A7）
└─ docs/                # 产品文档与模块文档
```

## 常用命令
| 命令 | 说明 |
| --- | --- |
| `make verify` | 交付总闸门：规范校验 + 全量测试，全绿才算完成 |
| `make lint` | 仅规范校验（禁用字符/文件头/必需文件/命名） |
| `make check` | 仅跑测试 |
| `make build` | 编译 CLI 到 bin/ |
| `go test ./...` | 运行全部 Go 测试 |

## 文档
- `docs/Rove_PRD_产品说明书.txt` -- 产品需求
- `docs/Rove_技术架构书.txt` -- 技术架构
- `docs/modules/core/` -- 模块文档（编码前闸门）
- `test_cases/core/unit/` -- 测试用例文档

## 里程碑
- [x] M1 骨架 + Fetch + Content
- [x] M2 ES 索引 + BM25
- [x] M3 Vector + Hybrid + Evidence
- [x] M4 Crawler
- [x] M5 Browser Escalation
- [x] M6 TUI + 收尾
- [x] P1 MCP / Map / Query Router / Reranker / Index Migrate

## 维护者
Charles <w1400214654@outlook.com>

## 许可
Apache-2.0，详见 [LICENSE](./LICENSE)。
