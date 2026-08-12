/*
 * 文件作用：HTML Parser -- 提取元数据（标题/描述/canonical/链接/语言/发布时间）与正文（readability）+ Markdown 视图。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"bytes"
	"context"
	"net/url"
	"strings"
	"time"

	mdconv "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/go-shiori/go-readability"
	"golang.org/x/net/html"

	"rove/pkg/document"
	"rove/pkg/rove"
)

// HTMLParser 解析 HTML 文档。
type HTMLParser struct{}

// NewHTMLParser 构造 HTMLParser。
func NewHTMLParser() *HTMLParser {

	return &HTMLParser{}
}

// Parse 提取元数据与正文：元数据走 DOM 遍历，正文走 readability，失败时降级为全文文本。
func (p *HTMLParser) Parse(ctx context.Context, raw *document.RawDocument) (*ParsedContent, error) {

	doc, err := html.Parse(bytes.NewReader(raw.Body))
	if err != nil {
		return nil, rove.NewError("content.html_parse", rove.CategoryContent, false, "html parse: %v", err)
	}

	pc := &ParsedContent{Metadata: map[string]any{}}
	walkMeta(doc, pc)

	// 正文提取：readability 失败时降级为全文文本
	baseURL, urlErr := url.Parse(raw.URL)
	if urlErr != nil {
		pc.Content = textFromHTML(doc)
		pc.Markdown = pc.Content
		return pc, nil
	}
	article, err := readability.FromReader(bytes.NewReader(raw.Body), baseURL)
	if err != nil {
		pc.Content = textFromHTML(doc)
		pc.Markdown = pc.Content
		return pc, nil
	}

	if pc.Title == "" {
		pc.Title = article.Title
	}
	pc.Content = strings.TrimSpace(article.TextContent)
	if markdown, merr := mdconv.NewConverter("", true, nil).ConvertString(article.Content); merr == nil {
		pc.Markdown = markdown
	} else {
		pc.Markdown = pc.Content
	}
	if pc.PublishedAt == nil && article.PublishedTime != nil {
		pc.PublishedAt = article.PublishedTime
	}
	return pc, nil
}

// walkMeta 遍历 DOM 收集元数据：lang/title/meta/link rel=canonical/链接。
func walkMeta(node *html.Node, pc *ParsedContent) {

	if node.Type == html.ElementNode {
		switch node.Data {
		case "html":
			for _, attr := range node.Attr {
				if attr.Key == "lang" {
					pc.Language = attr.Val
				}
			}
		case "title":
			if node.FirstChild != nil {
				pc.Title = strings.TrimSpace(node.FirstChild.Data)
			}
		case "meta":
			collectMeta(node, pc)
		case "link":
			collectCanonical(node, pc)
		case "a":
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					pc.Links = append(pc.Links, attr.Val)
				}
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walkMeta(child, pc)
	}
}

// collectMeta 提取 meta 标签中的描述/OG 信息/发布时间。
func collectMeta(node *html.Node, pc *ParsedContent) {

	var name, property, content string
	for _, attr := range node.Attr {
		switch attr.Key {
		case "name":
			name = attr.Val
		case "property":
			property = attr.Val
		case "content":
			content = attr.Val
		}
	}

	switch {
	case strings.EqualFold(property, "og:title") && pc.Title == "":
		pc.Title = content
	case strings.EqualFold(name, "description") && pc.Description == "":
		pc.Description = content
	case strings.EqualFold(property, "og:description") && pc.Description == "":
		pc.Description = content
	case strings.EqualFold(property, "article:published_time"):
		if parsed, perr := time.Parse(time.RFC3339, content); perr == nil {
			pc.PublishedAt = &parsed
		}
	}
}

// collectCanonical 提取 <link rel="canonical">。
func collectCanonical(node *html.Node, pc *ParsedContent) {

	var rel, href string
	for _, attr := range node.Attr {
		switch attr.Key {
		case "rel":
			rel = attr.Val
		case "href":
			href = attr.Val
		}
	}
	if pc.CanonicalURL == "" && strings.Contains(strings.ToLower(rel), "canonical") {
		pc.CanonicalURL = href
	}
}

// textFromHTML 提取全部文本节点（readability 失败时的降级路径）。
func textFromHTML(node *html.Node) string {

	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(strings.TrimSpace(n.Data))
			sb.WriteString(" ")
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.TrimSpace(sb.String())
}
