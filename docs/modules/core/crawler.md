# 爬取（crawler + scheduler）
> 模块职责：seed -> frontier（SQLite）-> 调度（host 令牌/robots/退避）-> fetch -> pipeline -> index -> 链接回填（规格书 §4 / A1）。
> 负责人：Charles
> 系统模块：core
> 关联服务：pkg/crawler、internal/scheduler、pkg/content、pkg/index

## 爬取编排
### 功能描述
有界并发工作池消费 frontier；每页按 robots 策略放行，抓取后经内容管线去重索引，同 host 链接回填（深度内）。调度器执行 crawl.delay（默认 500ms，环境变量 ROVE_CRAWL_DELAY 覆盖）最小抓取间隔，礼貌爬取避免触发站点限速。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| New | `crawler.New(frontier, sched, robots, policy, fetcher, registry, pipeline, indexer, workers, maxPages) *Crawler` | - |
| Run | `(*Crawler).Run(ctx, seeds []string) (*Stats, error)` | 一轮抓取 |
| NewFrontier | `scheduler.NewFrontier(path string) (*Frontier, error)` | SQLite 持久化（:memory: 需单连接） |
| NewScheduler | `scheduler.NewScheduler(frontier, Options) *Scheduler` | 调度（原子 UPDATE-RETURNING 认领） |

### 入参要求
seed 合法 URL；maxPages > 0；workers > 0；索引须 EnsureIndexes。

### 返回
Stats{Visited, Indexed, Failed, Duplicate}；单页失败不中断整轮。
