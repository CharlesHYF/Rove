/*
 * 文件作用：语义分块 -- 以 Markdown 为输入，按标题/段落单元切分并维护 heading_path 与 position（规格书 §4.1）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"context"
	"strings"

	"rove/pkg/document"
)

// Chunker 按语义结构切分 Markdown 为 Chunk。
type Chunker struct {
	MaxTokens int
	Overlap   int
}

// NewChunker 构造 Chunker。
func NewChunker(maxTokens, overlap int) *Chunker {

	return &Chunker{MaxTokens: maxTokens, Overlap: overlap}
}

// block 是 Markdown 的一个语义单元。
type block struct {
	headingLevel int // 0 = 非标题
	text         string
}

// Chunk 将 Markdown 按标题/段落/列表单元切分，维护 heading_path 与 position。
// 标题是硬边界（换标题即换 chunk）；超长块按 MaxTokens 词窗口切分；相邻 chunk 之间保留 Overlap。
func (c *Chunker) Chunk(ctx context.Context, doc *document.Document) ([]document.Chunk, error) {

	blocks := splitBlocks(doc.Markdown)
	var chunks []document.Chunk
	var current []string
	var currentTokens int
	var stack []string // heading 栈
	var levels []int   // heading 层级栈

	flush := func() {
		if len(current) == 0 {
			return
		}
		position := len(chunks)
		content := strings.Join(current, "\n\n")
		chunks = append(chunks, document.Chunk{
			ChunkID:     document.ChunkID(doc.ID, position),
			DocumentID:  doc.ID,
			Position:    position,
			HeadingPath: strings.Join(stack, " > "),
			Content:     cleanBlock(content),
			TokenCount:  approxTokens(content),
			Language:    doc.Language,
		})
	}

	// seedOverlap 以刚 flush 内容的尾部 Overlap 个词作为下一 chunk 的开头种子。
	seedOverlap := func() {
		if c.Overlap <= 0 {
			current = nil
			currentTokens = 0
			return
		}
		tail := tailTokens(strings.Join(current, "\n\n"), c.Overlap)
		current = nil
		currentTokens = 0
		if tail != "" {
			current = []string{tail}
			currentTokens = approxTokens(tail)
		}
	}

	for _, b := range blocks {
		if b.headingLevel > 0 {
			flush() // 标题是硬边界
			// 更新 heading 栈：弹出层级 >= 当前级别的标题
			for len(levels) > 0 && levels[len(levels)-1] >= b.headingLevel {
				levels = levels[:len(levels)-1]
				stack = stack[:len(stack)-1]
			}
			levels = append(levels, b.headingLevel)
			stack = append(stack, strings.TrimSpace(strings.TrimLeft(b.text, "# ")))
			continue
		}

		for _, part := range splitByTokens(b.text, c.MaxTokens) {
			partTokens := approxTokens(part)
			if currentTokens+partTokens > c.MaxTokens && len(current) > 0 {
				flush()
				seedOverlap()
			}
			current = append(current, part)
			currentTokens += partTokens
		}
	}
	flush()
	return chunks, nil
}

// splitBlocks 按空行分割 Markdown，识别 # 标题行。
func splitBlocks(markdown string) []block {

	var blocks []block
	for _, para := range strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		lines := strings.Split(para, "\n")
		if len(lines) > 0 {
			first := lines[0]
			if level := headingLevel(first); level > 0 {
				blocks = append(blocks, block{headingLevel: level, text: first})
				if len(lines) > 1 {
					blocks = append(blocks, block{text: strings.Join(lines[1:], "\n")})
				}
				continue
			}
		}
		blocks = append(blocks, block{text: para})
	}
	return blocks
}

// headingLevel 返回行的标题级别（1-6），非标题返回 0。
func headingLevel(line string) int {

	if !strings.HasPrefix(line, "#") {
		return 0
	}
	level := 0
	for _, r := range line {
		if r != '#' {
			break
		}
		level++
	}
	if level > 6 || (level < len(line) && line[level] != ' ') {
		return 0
	}
	return level
}

// splitByTokens 将超长文本按 maxTokens 切为词窗口。
func splitByTokens(text string, maxTokens int) []string {

	if approxTokens(text) <= maxTokens {
		return []string{text}
	}
	fields := strings.Fields(text)
	var parts []string
	for len(fields) > 0 {
		size := maxTokens
		if size > len(fields) {
			size = len(fields)
		}
		parts = append(parts, strings.Join(fields[:size], " "))
		fields = fields[size:]
	}
	return parts
}

// tailTokens 返回字符串末尾 n 个词；不足 n 个时返回原串。
func tailTokens(text string, count int) string {

	fields := strings.Fields(text)
	if len(fields) <= count {
		return text
	}
	return strings.Join(fields[len(fields)-count:], " ")
}

// cleanBlock 去除行首 Markdown 标记（#、-、*、>、数字列表、代码围栏）。
func cleanBlock(text string) string {

	lines := strings.Split(text, "\n")
	for index, line := range lines {
		line = strings.TrimSpace(line)
		if line == "```" || strings.HasPrefix(line, "```") {
			line = ""
		} else {
			line = strings.TrimLeft(line, "#>*- ")
			line = strings.TrimSpace(line)
		}
		lines[index] = line
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// approxTokens 近似 token 数：单词数 + 非 ASCII 字符数/2。
func approxTokens(text string) int {

	words := len(strings.Fields(text))
	nonASCII := 0
	for _, r := range text {
		if r > 127 {
			nonASCII++
		}
	}
	return words + nonASCII/2
}
