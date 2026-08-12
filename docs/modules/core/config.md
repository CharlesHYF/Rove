# 配置（config）
> 模块职责：加载与校验 Rove 配置 -- YAML 文件 + ROVE_ 环境变量 + 默认值（规格书 §4.1）。
> 负责人：Charles
> 系统模块：core
> 关联服务：internal/config、全链路

## 配置加载
### 功能描述
按"默认值 -> YAML 文件 -> 环境变量"三级加载配置并校验；校验失败返回 config.invalid 错误。

### 接口信息
| 函数 | 签名 | 说明 |
| --- | --- | --- |
| Load | `Load(path string) (*Config, error)` | path 为空则跳过文件 |
| Default | `Default() *Config` | 返回全部默认值 |

### 入参要求
| 参数 | 校验规则 | 默认值 |
| --- | --- | --- |
| fetch.max_body_bytes | > 0 | 10MB |
| fetch.timeout | > 0（Duration，支持 "30s"） | 30s |
| fetch.max_redirects | > 0 | 10 |
| chunk.max_tokens | > 0 | 800 |
| chunk.overlap | >= 0 | 100 |
| fetch.allow_private | bool | false |
| embedder.type | pseudo/openai/custom | pseudo |

### 返回
*Config；错误为 config.read / config.parse / config.invalid。
