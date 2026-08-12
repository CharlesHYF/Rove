/*
 * 文件作用：robots.txt 客户端 -- 每 host 缓存解析结果，默认遵守（规格书 §7）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package crawler 提供抓取策略与爬取编排。
package crawler

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/temoto/robotstxt"

	"rove/pkg/fetch"
)

const userAgent = "rove"

// robotsResult 是单个 host 的解析结果。
type robotsResult struct {
	data       *robotstxt.RobotsData
	crawlDelay time.Duration
}

// RobotsClient 缓存并查询 robots.txt。
type RobotsClient struct {
	fetcher *fetch.HTTPFetcher
	cache   map[string]robotsResult
}

// NewRobotsClient 构造 RobotsClient。
func NewRobotsClient(fetcher *fetch.HTTPFetcher) *RobotsClient {

	return &RobotsClient{fetcher: fetcher, cache: map[string]robotsResult{}}
}

// IsAllowed 判断 URL 是否被 robots.txt 允许；无 robots.txt 或拉取失败时 fail-open。
func (c *RobotsClient) IsAllowed(ctx context.Context, urlStr string) (bool, error) {

	u, err := url.Parse(urlStr)
	if err != nil || u.Host == "" {
		return false, nil
	}
	result, err := c.load(ctx, robotsKey(u))
	if err != nil || result.data == nil {
		return true, nil // fail-open
	}
	return result.data.TestAgent(u.Path, userAgent), nil
}

// CrawlDelay 返回页面所在 host 的 crawl-delay；未配置或拉取失败返回 0。
func (c *RobotsClient) CrawlDelay(ctx context.Context, pageURL string) (time.Duration, error) {

	u, err := url.Parse(pageURL)
	if err != nil || u.Host == "" {
		return 0, nil
	}
	result, err := c.load(ctx, robotsKey(u))
	if err != nil {
		return 0, nil
	}
	return result.crawlDelay, nil
}

// load 拉取并缓存 robots.txt（缓存键 = scheme://host，与页面同 scheme）。
func (c *RobotsClient) load(ctx context.Context, key string) (robotsResult, error) {

	if result, ok := c.cache[key]; ok {
		return result, nil
	}
	result := robotsResult{}
	robotsURL := key + "/robots.txt"
	raw, err := c.fetcher.Fetch(ctx, &fetch.Request{URL: robotsURL, Mode: fetch.ModeHTTP})
	if err != nil {
		c.cache[key] = result
		return result, nil
	}
	data, err := robotstxt.FromBytes(raw.Body)
	if err != nil {
		c.cache[key] = result
		return result, nil
	}
	result.data = data
	if group := data.FindGroup(userAgent); group != nil {
		result.crawlDelay = group.CrawlDelay
	}
	c.cache[key] = result
	return result, nil
}

// robotsKey 生成缓存键：scheme://host（小写）。
func robotsKey(u *url.URL) string {

	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Hostname())
}

// hostOf 提取 URL 的 hostname（小写）。
func hostOf(urlStr string) string {

	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}
