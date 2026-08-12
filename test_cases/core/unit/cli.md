# CLI 模块单元用例
> 模块：core/cli · 类型：unit · 对应代码：cmd/rove/fetch_test.go

## 用例清单

| 用例名 | 场景 | 前置条件 | 断言 |
| --- | --- | --- | --- |
| TestFetchCommandJSON | JSON 输出抓取结果 | httptest HTML 页面、ROVE_ALLOW_PRIVATE=true | 输出含 `"Title": "CLI Test Page"` 与 `"Chunks":` |
| TestFetchCommandHuman | 人类可读摘要输出 | 同上 | 输出含 "title: CLI Test Page" 与 "chunks: " |
