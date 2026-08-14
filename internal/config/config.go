/*
 * 文件作用：加载与校验 Rove 配置 -- YAML 文件 + ROVE_ 环境变量 + 默认值三级覆盖（规格书 §4.1）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package config 负责加载与校验 Rove 配置。
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"

	"rove/pkg/rove"
)

// Duration 支持 YAML 字符串形式的 time.Duration（如 "30s"、"500ms"）。
type Duration struct {
	time.Duration
}

// UnmarshalYAML 将 "30s" 解析为 time.Duration。
func (d *Duration) UnmarshalYAML(node *yaml.Node) error {

	parsed, err := time.ParseDuration(node.Value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", node.Value, err)
	}
	d.Duration = parsed
	return nil
}

// Config 是 Rove 全部配置的单一结构（规格书 §4.1）。
type Config struct {
	Fetch    FetchConfig    `yaml:"fetch"`
	Chunk    ChunkConfig    `yaml:"chunk"`
	Crawl    CrawlConfig    `yaml:"crawl"`
	ES       ESConfig       `yaml:"es"`
	Robots   RobotsConfig   `yaml:"robots"`
	Embedder EmbedderConfig `yaml:"embedder"`
	Search   SearchConfig   `yaml:"search"`
	Index    IndexConfig    `yaml:"index"`
	Browser  BrowserConfig  `yaml:"browser"`
}

// FetchConfig 抓取相关配置。
type FetchConfig struct {
	MaxBodyBytes int64    `yaml:"max_body_bytes"`
	Timeout      Duration `yaml:"timeout"`
	MaxRedirects int      `yaml:"max_redirects"`
	UserAgent    string   `yaml:"user_agent"`
	AllowPrivate bool     `yaml:"allow_private"`
}

// ChunkConfig 分块相关配置。
type ChunkConfig struct {
	MaxTokens int `yaml:"max_tokens"`
	Overlap   int `yaml:"overlap"`
}

// CrawlConfig 爬虫相关配置。
type CrawlConfig struct {
	MaxPages        int      `yaml:"max_pages"`
	MaxDepth        int      `yaml:"max_depth"`
	DomainAllowlist []string `yaml:"domain_allowlist"`
	Workers         int      `yaml:"workers"`
	StateDB         string   `yaml:"state_db"`
	Delay           Duration `yaml:"delay"` // 每次抓取的最小间隔（礼貌爬取，避免触发站点限速）
}

// ESConfig Elasticsearch 配置（M2 使用）。
type ESConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// RobotsConfig robots.txt 策略配置（M4 使用）。
type RobotsConfig struct {
	Enabled bool `yaml:"enabled"`
}

// EmbedderConfig 向量嵌入配置（M3 使用）。
type EmbedderConfig struct {
	Type     string `yaml:"type"` // pseudo | openai | custom
	Endpoint string `yaml:"endpoint"`
}

// SearchConfig 检索相关配置（M2 起生效）。
type SearchConfig struct {
	TopK int `yaml:"top_k"`
}

// IndexConfig 索引相关配置。
type IndexConfig struct {
	EmbeddingDim int    `yaml:"embedding_dim"`
	Prefix       string `yaml:"prefix"`
}

// BrowserConfig 浏览器配置（M5 起生效）。
type BrowserConfig struct {
	Enabled        bool     `yaml:"enabled"`
	ExecutablePath string   `yaml:"executable_path"`
	Timeout        Duration `yaml:"timeout"`
}

// Default 返回全部默认值。
func Default() *Config {

	return &Config{
		Fetch: FetchConfig{
			MaxBodyBytes: 10 << 20,
			Timeout:      Duration{Duration: 30 * time.Second},
			MaxRedirects: 10,
			UserAgent:    "rove/0.1",
			AllowPrivate: false,
		},
		Chunk:    ChunkConfig{MaxTokens: 800, Overlap: 100},
		Crawl:    CrawlConfig{MaxPages: 1000, MaxDepth: 3, Workers: 2, StateDB: "rove.db", Delay: Duration{Duration: 500 * time.Millisecond}},
		ES:       ESConfig{URL: "http://localhost:9200"},
		Robots:   RobotsConfig{Enabled: true},
		Embedder: EmbedderConfig{Type: "pseudo"},
		Search:   SearchConfig{TopK: 10},
		Index:    IndexConfig{EmbeddingDim: 256, Prefix: "rove"},
		Browser:  BrowserConfig{Enabled: true, Timeout: Duration{Duration: 30 * time.Second}},
	}
}

// Load 加载配置：默认值 -> YAML 文件（path 非空时）-> ROVE_ 环境变量。
func Load(path string) (*Config, error) {

	c := Default()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, rove.NewError("config.read", rove.CategoryContent, false, "read config %s: %v", path, err)
		}
		if err := yaml.Unmarshal(data, c); err != nil {
			return nil, rove.NewError("config.parse", rove.CategoryContent, false, "parse config %s: %v", path, err)
		}
	}
	applyEnv(c)
	if err := validate(c); err != nil {
		return nil, err
	}
	return c, nil
}

// applyEnv 应用 ROVE_ 前缀环境变量（规格书 §4.1 全量映射）。
func applyEnv(c *Config) {

	applyBool(&c.Fetch.AllowPrivate, "ROVE_ALLOW_PRIVATE")
	applyInt(&c.Fetch.MaxRedirects, "ROVE_FETCH_MAX_REDIRECTS")
	applyInt(&c.Chunk.MaxTokens, "ROVE_CHUNK_MAX_TOKENS")
	applyInt(&c.Chunk.Overlap, "ROVE_CHUNK_OVERLAP")
	applyInt(&c.Crawl.MaxPages, "ROVE_CRAWL_MAX_PAGES")
	applyInt(&c.Crawl.MaxDepth, "ROVE_CRAWL_MAX_DEPTH")
	applyInt(&c.Crawl.Workers, "ROVE_CRAWL_WORKERS")
	applyString(&c.Crawl.StateDB, "ROVE_CRAWL_STATE_DB")
	applyDuration(&c.Crawl.Delay.Duration, "ROVE_CRAWL_DELAY")
	applyInt(&c.Search.TopK, "ROVE_SEARCH_TOP_K")
	applyInt(&c.Index.EmbeddingDim, "ROVE_INDEX_EMBEDDING_DIM")
	applyString(&c.Index.Prefix, "ROVE_INDEX_PREFIX")
	applyString(&c.ES.URL, "ROVE_ES_URL")
	applyString(&c.ES.Username, "ROVE_ES_USERNAME")
	applyString(&c.ES.Password, "ROVE_ES_PASSWORD")
	applyString(&c.Embedder.Type, "ROVE_EMBEDDER_TYPE")
	applyString(&c.Embedder.Endpoint, "ROVE_EMBEDDER_ENDPOINT")
	applyBool(&c.Robots.Enabled, "ROVE_ROBOTS_ENABLED")
	applyBool(&c.Browser.Enabled, "ROVE_BROWSER_ENABLED")
	applyString(&c.Browser.ExecutablePath, "ROVE_BROWSER_EXECUTABLE")
}

// applyBool 应用布尔环境变量。
func applyBool(target *bool, key string) {

	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			*target = parsed
		}
	}
}

// applyInt 应用整数环境变量。
func applyInt(target *int, key string) {

	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			*target = parsed
		}
	}
}

// applyString 应用字符串环境变量。
func applyString(target *string, key string) {

	if value := os.Getenv(key); value != "" {
		*target = value
	}
}

// applyDuration 应用时长环境变量（如 "500ms"、"2s"），解析失败保持默认值。
func applyDuration(target *time.Duration, key string) {

	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			*target = parsed
		}
	}
}

// validate 校验关键配置项，违规返回 config.invalid。
func validate(c *Config) error {

	if c.Fetch.MaxBodyBytes <= 0 {
		return rove.NewError("config.invalid", rove.CategoryContent, false, "fetch.max_body_bytes must be > 0")
	}
	if c.Fetch.Timeout.Duration <= 0 {
		return rove.NewError("config.invalid", rove.CategoryContent, false, "fetch.timeout must be > 0")
	}
	if c.Fetch.MaxRedirects <= 0 {
		return rove.NewError("config.invalid", rove.CategoryContent, false, "fetch.max_redirects must be > 0")
	}
	if c.Chunk.MaxTokens <= 0 {
		return rove.NewError("config.invalid", rove.CategoryContent, false, "chunk.max_tokens must be > 0")
	}
	if c.Chunk.Overlap < 0 {
		return rove.NewError("config.invalid", rove.CategoryContent, false, "chunk.overlap must be >= 0")
	}
	if c.Crawl.MaxPages <= 0 {
		return rove.NewError("config.invalid", rove.CategoryContent, false, "crawl.max_pages must be > 0")
	}
	if c.Crawl.MaxDepth < 0 {
		return rove.NewError("config.invalid", rove.CategoryContent, false, "crawl.max_depth must be >= 0")
	}
	if c.Crawl.Workers <= 0 {
		return rove.NewError("config.invalid", rove.CategoryContent, false, "crawl.workers must be > 0")
	}
	if c.Crawl.Delay.Duration < 0 {
		return rove.NewError("config.invalid", rove.CategoryContent, false, "crawl.delay must be >= 0")
	}
	if c.Search.TopK <= 0 {
		return rove.NewError("config.invalid", rove.CategoryContent, false, "search.top_k must be > 0")
	}
	if c.Index.EmbeddingDim <= 0 {
		return rove.NewError("config.invalid", rove.CategoryContent, false, "index.embedding_dim must be > 0")
	}
	return nil
}
