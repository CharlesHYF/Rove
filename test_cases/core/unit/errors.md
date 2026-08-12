<!--
	文件作用：errors 模块单元测试用例，覆盖错误构造与 machine-readable JSON 契约
	创建日期：2026-08-12
-->
# errors 模块（core/errors）-- 单元测试用例

> 模块简介：统一 machine-readable 错误模型（rove.Error / NewError / WithDetails / JSON 序列化）
> 测试环境：开发环境（dev）
> 测试框架：Go testing + testify

## 功能一：错误构造（NewError）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-ERR-001 | P0 | 字段与格式化 | - | NewError("fetch.invalid_url", content, false, "invalid url: %s", "://bad") | Code/Category/Retryable 正确，Error() 返回格式化 message | 与预期一致 | Pass | 对应 TestNewErrorMessage |
| TC-ERR-002 | P0 | JSON 序列化契约 | - | NewError + WithDetails 后 Marshal | 输出含 code/category/retryable/message/details 五个字段 | 与预期一致 | Pass | 对应 TestErrorJSONShape |
| TC-ERR-003 | P0 | 可重试性语义 | - | network=true / policy=false | Retryable 标志正确 | 与预期一致 | Pass | 对应 TestRetryableCategories |
