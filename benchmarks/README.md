# 检索基准（benchmarks）

评估骨架：`pkg/eval` 提供 Recall@K / NDCG@K / MRR 纯函数。

## 使用（bench_search.sh）

```bash
bash benchmarks/bench_search.sh <seed-url> "query1" "query2"
```

流程：seed 抓取索引 -> 对每个查询执行 `rove search --json` -> 输出命中标题与耗时。
完整 benchmark 语料与相关判定（query -> 期望文档）属 P2，评估脚本按需扩展。

> 真实质量调优（权重/阈值）依赖 benchmark 语料，见架构书 §15。
