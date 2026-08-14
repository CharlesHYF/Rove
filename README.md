# Rove

Open web infrastructure for AI agents.

Crawl. Browse. Index. Retrieve. Rank. Evidence.

自己的索引、自己的证据、自己的账单（Self-owned Retrieval）。

## 环境要求

- Go >= 1.22
- Docker（运行 Elasticsearch）
- Chrome / Edge（浏览器渲染功能，可选但推荐）

## 快速开始

```bash
# 1) 构建
go build -o bin/rove ./cmd/rove

# 2) 启动 Elasticsearch（首次拉取镜像约 1.3GB，之后秒起）
docker compose up -d elasticsearch

# 3) 初始化索引
./bin/rove index init
```

## 使用

| 命令 | 说明 |
| --- | --- |
| `rove` | 进入 TUI：默认对话式查询（回车检索、Ctrl+E 展开分数、Ctrl+L 清空），Tab 切换五视图调试器 |
| `rove fetch <url> [--mode auto\|http\|browser]` | 抓取单页，输出标准 Document（auto 模式动态页自动浏览器升级） |
| `rove crawl <seed> [--max-pages N] [--max-depth D]` | 抓取整个站点：发现链接、去重、索引（含向量） |
| `rove search <query> [--json]` | BM25 + 向量 Hybrid 检索，输出 Evidence 与分数分解 |
| `rove browse <url>` | 浏览器渲染页面状态（文本/链接/交互元素 + 稳定元素 id） |
| `rove map <url> [--json]` | 发现页面链接（站内/站外，仅 URL Discovery，不建索引） |
| `rove mcp` | 以 stdio 启动 MCP Server，供 Claude Desktop / Cursor 等 Agent 接入（rove_search/fetch/browse/crawl） |
| `rove index init\|status\|stats\|rebuild` | 索引生命周期管理 |

## 最小演示

```bash
docker compose up -d elasticsearch
./bin/rove index init
./bin/rove crawl https://example.com --max-pages 5
./bin/rove search "example domain"
./bin/rove search "example domain" --json   # 查看 lexical/vector/fusion/rank 分数与 Evidence
./bin/rove                                  # TUI 里实时观察
```

## 里程碑状态

- [x] M1 骨架 + Fetch + Content
- [x] M2 ES 索引 + BM25
- [x] M3 Vector + Hybrid + Evidence
- [x] M4 Crawler
- [x] M5 Browser Escalation
- [x] M6 TUI + 收尾

## MCP 接入

`rove mcp` 以 stdio 启动 MCP Server。Claude Desktop / Cursor 的 mcpServers 配置示例：

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

## 配置

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
| `ROVE_BROWSER_EXECUTABLE` | 浏览器可执行文件路径（自动发现失败时指定） | 自动发现 |
| `ROVE_BROWSER_ENABLED` | 浏览器功能开关 | true |
| `ROVE_SEARCH_TOP_K` | 检索默认返回条数 | 10 |

## 文档

- `docs/Rove_PRD_产品说明书.txt` -- 产品需求
- `docs/Rove_技术架构书.txt` -- 技术架构
- `docs/modules/core/` -- 模块文档

## 许可

Apache-2.0
