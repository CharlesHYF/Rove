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

// CrawlConfig 爬虫相关配置（M4 使用）。
type CrawlConfig struct {
	MaxPages        int      `yaml:"max_pages"`
	MaxDepth        int      `yaml:"max_depth"`
	DomainAllowlist []string `yaml:"domain_allowlist"`
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
		Crawl:    CrawlConfig{MaxPages: 1000, MaxDepth: 3},
		ES:       ESConfig{URL: "http://localhost:9200"},
		Robots:   RobotsConfig{Enabled: true},
		Embedder: EmbedderConfig{Type: "pseudo"},
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

// applyEnv 应用 ROVE_ 前缀环境变量（M1 支持子集，M2 补齐全量映射）。
func applyEnv(c *Config) {

	if value := os.Getenv("ROVE_ALLOW_PRIVATE"); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			c.Fetch.AllowPrivate = parsed
		}
	}
	if value := os.Getenv("ROVE_ES_URL"); value != "" {
		c.ES.URL = value
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
	return nil
}
