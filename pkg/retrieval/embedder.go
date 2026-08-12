/*
 * 文件作用：PseudoEmbedder -- 确定性 feature hashing 词袋向量（规格书 §5.2，无 API Key 的 Hybrid 底座）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package retrieval

import (
	"context"
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

// pseudoDim 是 PseudoEmbedder 的固定向量维度（与索引 mapping 的 embedding_dim 默认值一致）。
const pseudoDim = 256

// Embedder 是可替换的向量嵌入接口（规格书 §4.1）。
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Dim() int
}

// PseudoEmbedder 基于 feature hashing 生成确定性词袋向量，零依赖、无模型。
type PseudoEmbedder struct{}

// NewPseudoEmbedder 构造 PseudoEmbedder。
func NewPseudoEmbedder() *PseudoEmbedder {

	return &PseudoEmbedder{}
}

// Dim 返回向量维度。
func (p *PseudoEmbedder) Dim() int {

	return pseudoDim
}

// Embed 将文本转为 256 维 L2 归一化计数向量。
func (p *PseudoEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {

	vectors := make([][]float32, 0, len(texts))
	for _, text := range texts {
		vectors = append(vectors, pseudoVector(text))
	}
	return vectors, nil
}

// pseudoVector 计算单条文本的伪向量：token -> fnv64 -> dim 取模计数 -> L2 归一化。
func pseudoVector(text string) []float32 {

	vector := make([]float32, pseudoDim)
	hasher := fnv.New64a()
	for _, token := range tokenizePseudo(text) {
		hasher.Reset()
		_, _ = hasher.Write([]byte(token))
		index := hasher.Sum64() % pseudoDim
		vector[index]++
	}
	normalize(vector)
	return vector
}

// tokenizePseudo 将文本切成小写 token（按非字母数字分隔）。
func tokenizePseudo(text string) []string {

	var tokens []string
	var current strings.Builder
	flush := func() {
		if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return tokens
}

// normalize 原地 L2 归一化；零向量保持全零。
func normalize(vector []float32) {

	var sumSquares float64
	for _, value := range vector {
		sumSquares += float64(value) * float64(value)
	}
	if sumSquares == 0 {
		return
	}
	norm := float32(math.Sqrt(sumSquares))
	for index := range vector {
		vector[index] /= norm
	}
}

// Cosine 计算两个向量的余弦相似度（维度不一致或零向量返回 0）。
func Cosine(a, b []float32) float64 {

	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for index := range a {
		dot += float64(a[index]) * float64(b[index])
		normA += float64(a[index]) * float64(a[index])
		normB += float64(b[index]) * float64(b[index])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
