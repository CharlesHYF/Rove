/*
 * 文件作用：rove search 子命令 -- 从自有 ES 索引执行 BM25 检索（规格书 §9 A3 验收）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package main

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"rove/internal/app"
	"rove/internal/config"
	"rove/pkg/elastic"
	"rove/pkg/retrieval"
	"rove/protocol"
)

var searchFlags struct {
	topK     int
	domain   string
	language string
	vertical string
	json     bool
}

// searchCmd 实现 rove search <query>。
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "从自有 ES 索引检索（BM25）",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

// init 注册 search 命令参数。
func init() {

	searchCmd.Flags().IntVar(&searchFlags.topK, "top-k", 0, "返回条数（默认配置 search.top_k）")
	searchCmd.Flags().StringVar(&searchFlags.domain, "domain", "", "按域名过滤")
	searchCmd.Flags().StringVar(&searchFlags.language, "language", "", "按语言过滤（如 en/zh）")
	searchCmd.Flags().StringVar(&searchFlags.vertical, "vertical", "auto", "垂类路由: auto|web|docs|code|academic")
	searchCmd.Flags().BoolVar(&searchFlags.json, "json", false, "以 JSON 输出")
}

// runSearch 执行检索。
func runSearch(cmd *cobra.Command, args []string) error {

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	client, err := elastic.New(cfg.ES.URL, cfg.ES.Username, cfg.ES.Password, cfg.Index.Prefix)
	if err != nil {
		return err
	}

	topK := searchFlags.topK
	if topK <= 0 {
		topK = cfg.Search.TopK
	}
	vertical, err := retrieval.ParseVertical(searchFlags.vertical)
	if err != nil {
		return err
	}
	query := &retrieval.Query{
		Text: args[0],
		TopK: topK,
		Filters: retrieval.Filters{
			Domain:   searchFlags.domain,
			Language: searchFlags.language,
		},
		Vertical: vertical,
	}
	searchService := app.NewSearch(retrieval.New(client, retrieval.NewPseudoEmbedder(cfg.Index.EmbeddingDim)))
	result, err := searchService.Search(cmd.Context(), query)
	if err != nil {
		return err
	}
	items, err := searchService.Evidence(cmd.Context(), result)
	if err != nil {
		return err
	}

	if searchFlags.json {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(protocol.FromSearchResult(result, items))
	}

	cmd.Printf("query: %s (trace %s, vertical %s)\n", result.Query, result.TraceID, result.Vertical)
	for rank, item := range items {
		cmd.Printf("%d. %s  [%.3f]\n", rank+1, item.Title, item.Score)
		cmd.Printf("   %s\n", item.URL)
		cmd.Printf("   %s\n", item.Text)
	}
	return nil
}
