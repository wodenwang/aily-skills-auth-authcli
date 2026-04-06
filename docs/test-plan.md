# Test Plan

## Scope

- Module: `auth-cli check`
- Risk: 输入优先级漂移、缓存失效规则漂移、上游错误被误映射为 deny、输出协议不稳定

## Test Types

- Unit
- Contract
- Integration
- E2E

## Cases

| ID | Scenario | Expected |
|----|----------|----------|
| T1 | 显式参数、环境变量、上下文文件、本地配置同时存在 | 显式参数优先 |
| T2 | 命中未到刷新窗口的缓存 | 直接返回缓存 token |
| T3 | 命中刷新窗口内 token | 调用 `/api/v1/token/refresh` 并覆盖缓存 |
| T4 | refresh 返回 `TOKEN_REVOKED` | 删除缓存并重新 `/api/v1/auth/check` |
| T5 | IAM 返回 deny | 退出码为 `10`，不输出 token |
| T6 | 上游超时或不可用 | 退出码为 `40`，fail-closed |
| T7 | `json` 输出 allow | 字段名与冻结契约一致 |
| T8 | `env` 输出 deny | 只输出 deny 字段，不输出 token |
| T9 | 同一用户切换 Skill | 不复用旧缓存 |

## Tooling

- Framework: Go `testing` + `httptest`
- Fixtures: 临时目录缓存文件、模拟 IAM 服务
- CI: `go test ./...` + `./scripts/beta-smoke.sh`
