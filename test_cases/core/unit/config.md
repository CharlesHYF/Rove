<!--
	文件作用：config 模块单元测试用例，覆盖默认值 / YAML 文件 / 环境变量三级加载与校验
	创建日期：2026-08-12
-->
# config 模块（core/config）-- 单元测试用例

> 模块简介：配置加载与校验（YAML + ROVE_ 环境变量 + 默认值）
> 测试环境：开发环境（dev）
> 测试框架：Go testing + testify

## 功能一：配置加载（Load）

| 测试编号 | 优先级 | 测试用例 | 前置条件 | 输入内容 | 预计结果 | 实际结果 | 是否通过 | 备注 |
| -------- | ------ | -------- | -------- | -------- | -------- | -------- | -------- | ---- |
| TC-CFG-001 | P0 | 默认值 | 无配置文件 | Load("") | max_body=10MB、timeout=30s、redirects=10、chunk=800/100、embedder=pseudo | 与预期一致 | Pass | 对应 TestDefaults |
| TC-CFG-002 | P0 | YAML 覆盖默认值 | testdata/config.yaml | Load(testdata/config.yaml) | max_body=1MB、timeout=10s、max_tokens=400 生效 | 与预期一致 | Pass | 对应 TestLoadFile |
| TC-CFG-003 | P1 | 环境变量覆盖 | - | ROVE_ALLOW_PRIVATE=true、ROVE_ES_URL=http://127.0.0.1:9200 | AllowPrivate=true、ES.URL 生效 | 与预期一致 | Pass | 对应 TestEnvOverride |
| TC-CFG-004 | P1 | 非法时长报错 | - | YAML 中 timeout: nope | 返回 config.parse 错误 | 与预期一致 | Pass | 对应 TestLoadInvalidDuration |
