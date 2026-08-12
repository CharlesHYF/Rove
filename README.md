# Rove

Open web infrastructure for AI agents.

Crawl. Browse. Index. Retrieve. Rank. Evidence.

自己的索引、自己的证据、自己的账单（Self-owned Retrieval）。

## 里程碑状态

- [x] M1 骨架 + Fetch + Content（当前）
- [ ] M2 ES 索引 + BM25
- [ ] M3 Vector + Hybrid + Evidence
- [ ] M4 Crawler
- [ ] M5 Browser Escalation
- [ ] M6 TUI + 收尾验收

## 快速开始（M1）

```bash
go build -o bin/rove ./cmd/rove
./bin/rove fetch https://example.com --json
```

> auto 模式下，HTTP 内容不足（SPA/需交互）会提示需要浏览器渲染，浏览器能力在 M5 里程碑提供。

## 文档

- `docs/Rove_PRD_产品说明书.txt` -- 产品需求
- `docs/Rove_技术架构书.txt` -- 技术架构
- `docs/modules/core/` -- 模块文档（编码前闸门）

## 许可

Apache-2.0
