# MCP（protocol/mcp）
> 模块职责：Rove 原生 MCP Server -- 以 stdio 协议向 Agent 客户端（Claude Desktop / Cursor / VS Code 等）暴露 search/fetch/browse/crawl 四个工具（PRD FR-MCP-001，P1）。
> 负责人：Charles
> 系统模块：core
> 关联服务：protocol/mcp、internal/app、protocol

## MCP Server
### 功能描述
`rove mcp` 启动 stdio 传输的 MCP Server（协议层用官方 github.com/modelcontextprotocol/go-sdk，MIT 许可）。Server 名 `rove`，版本随 CLI version。暴露四个工具：
- `rove_search`：从自有 ES 索引执行 Hybrid 检索，输出 Evidence（URL/标题/文本/分数/来源/耗时/trace）；
- `rove_fetch`：抓取单页并输出标准 Document（auto 模式 HTTP 内容不足时自动浏览器升级）；
- `rove_browse`：浏览器渲染页面，输出 Page State（标题/链接/交互元素与稳定 id）；
- `rove_crawl`：从 seed 抓取站点并建立索引，输出统计（visited/indexed/failed/duplicate）。

工具实现只依赖注入的 Application Service 接口（SearchService/FetchService/BrowseService/CrawlService），不直接访问 ES/浏览器（架构书 §11：Interfaces 不得实现内部逻辑）。无鉴权（stdio 本地场景）；工具调用出错返回 isError=true 的文本内容，不允许 silent failure（NFR-001）。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| NewServer | `NewServer(deps Deps) *Server` | 构造并注册四个工具的 MCP Server（Deps 为四个服务接口） |
| Serve | `(*Server).Serve() error` | 以 stdio 运行服务，阻塞至 stdin 关闭 |

### 入参要求
- Deps 四接口均非 nil，由 cmd/rove/mcp.go 注入 internal/app 服务；
- 工具入参：rove_search（query 必填，top_k/domain/language 可选）、rove_fetch（url 必填，mode 可选 auto/http/browser）、rove_browse（url 必填）、rove_crawl（seed 必填，max_pages/max_depth 可选，0 取配置默认）。

### 返回
- 每个工具返回 text 内容（JSON 序列化的 SearchResponse / Document / PageState / 抓取统计）；
- 错误返回 isError=true 的文本（机器可读错误摘要）；
- 工具不发布 outputSchema（输出为 JSON 文本，结构化输出 schema 留待后续）。

## 客户端配置示例（README 收录）
Claude Desktop / Cursor 的 mcpServers 配置：
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
