# fetch 模块单元用例
> 模块：core/fetch · 类型：unit · 对应代码：pkg/fetch/fetch_test.go

## 用例清单

| 用例名 | 场景 | 前置条件 | 断言 |
| --- | --- | --- | --- |
| TestFetchFollowsRedirect | 301 跳转后取最终 URL | httptest 双路径 | raw.URL 含 /end、StatusCode 200、Timings.Total > 0 |
| TestFetchMaxRedirects | 无限重定向循环 | MaxRedirects=5 | 返回含 "redirect" 的错误（fetch.redirect_loop，不可重试） |
| TestFetchMaxBody | 响应超过 body 上限 | MaxBodyBytes=1024，响应 2048 字节 | 返回 fetch.body_too_large |
| TestFetchSSRFBlockLoopback | 默认策略阻止 loopback | AllowPrivate=false | 返回含 "blocked" 的错误（fetch.ssrf_blocked） |
| TestFetchCharsetGBK | GBK 内容转 UTF-8 | charset=gbk，"你好" 的 GBK 字节 | raw.Body == "你好" |
| TestFetchAutoInsufficient | auto 模式正文不足 | 短 HTML + script | errors.Is(err, ErrEscalationRequired) |
| TestFetchModeBrowserNotImplemented | browser 模式未实现 | Mode=browser | 返回含 "M5" 的错误（browser.not_implemented） |
| TestFetchRetriesNetworkError | 网络错误自动重试 | 前 2 次连接被重置 | 第 3 次成功，calls >= 3 |
