/*
 * 文件作用：定义 Rove 统一的 machine-readable 错误 -- 类别、可重试性与 JSON 序列化契约（NFR-001）。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
// Package rove 是 Rove Core 的根 API 门面，定义对外统一的 machine-readable 错误模型。
package rove

import "fmt"

// ErrorCategory 描述错误所属子系统类别（规格书 §5.6）。
type ErrorCategory string

const (
	CategoryNetwork   ErrorCategory = "network"
	CategoryTimeout   ErrorCategory = "timeout"
	CategoryPolicy    ErrorCategory = "policy"
	CategoryContent   ErrorCategory = "content"
	CategoryBrowser   ErrorCategory = "browser"
	CategoryIndex     ErrorCategory = "index"
	CategoryEmbedding ErrorCategory = "embedding"
)

// Error 是 Rove 统一的 machine-readable 错误。JSON 序列化字段与 CLI/Agent 消费方契约一致。
type Error struct {
	Code      string         `json:"code"`
	Category  ErrorCategory  `json:"category"`
	Retryable bool           `json:"retryable"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
}

// Error 实现 error 接口。
func (e *Error) Error() string {

	return e.Message
}

// WithDetails 附加结构化细节（如 host、attempt、url 脱敏值）。
func (e *Error) WithDetails(details map[string]any) *Error {

	e.Details = details
	return e
}

// NewError 构造标准错误。format/args 用于生成面向日志与 Agent 的 message。
func NewError(code string, category ErrorCategory, retryable bool, format string, args ...any) *Error {

	return &Error{
		Code:      code,
		Category:  category,
		Retryable: retryable,
		Message:   fmt.Sprintf(format, args...),
	}
}
