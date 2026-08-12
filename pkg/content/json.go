/*
 * 文件作用：JSON Parser -- 处理 application/json，保留结构并生成可读文本视图。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"context"
	"encoding/json"
	"strings"

	"rove/pkg/document"
	"rove/pkg/rove"
)

// JSONParser 处理 application/json。
type JSONParser struct{}

// NewJSONParser 构造 JSONParser。
func NewJSONParser() *JSONParser {

	return &JSONParser{}
}

// Parse 将 JSON 美化输出为 Content，并尽量提取 title 字段。
func (p *JSONParser) Parse(ctx context.Context, raw *document.RawDocument) (*ParsedContent, error) {

	var value any
	if err := json.Unmarshal(raw.Body, &value); err != nil {
		return nil, rove.NewError("content.json_parse", rove.CategoryContent, false, "json parse: %v", err)
	}

	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, rove.NewError("content.json_parse", rove.CategoryContent, false, "json marshal: %v", err)
	}

	pc := &ParsedContent{
		Content:  string(pretty),
		Markdown: "```json\n" + string(pretty) + "\n```",
		Metadata: map[string]any{},
	}
	if object, ok := value.(map[string]any); ok {
		if title, ok := object["title"].(string); ok {
			pc.Title = strings.TrimSpace(title)
		}
	}
	return pc, nil
}
