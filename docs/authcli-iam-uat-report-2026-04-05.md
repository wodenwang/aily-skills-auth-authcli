# AuthCLI And IAM Service UAT Report

## Summary

- Report time: `2026-04-05 21:11:53 CST`
- Target IAM base URL: `http://127.0.0.1:8000`
- AuthCLI repo: `aily-skills-auth-authcli`
- IAM repo: `aily-skills-auth-iam-service`
- UAT mode: real service joint integration

本次联调基于正在运行的真实 `iam-service`，按主项目 Phase 1 / UAT 目标重新执行 `authcli` 完整主路径验证。

## Environment

- Health check time: `2026-04-05 21:11:25 CST`
- Health result: `GET /healthz -> HTTP 200`, body `{"status":"ok"}`
- Real IAM UAT execution window:
  - started before `2026-04-05 21:11:25 CST`
  - completed at `2026-04-05 21:11:53 CST`
- Service-side observer log:
  - `2026-04-05 20:56:xx observed inbound localhost traffic: multiple POST /api/v1/auth/check and POST /api/v1/token/refresh requests returned 200 OK; no errors in current tail.`

## Case Results

| Case ID | Scenario | Result | Notes |
|---------|----------|--------|-------|
| UAT-01 | 服务健康检查 | PASS | `GET /healthz` 返回 `200 OK` |
| UAT-02 | 私聊 allow | PASS | `user_id=ou_abc123`, `chat_id=null`，返回 `allowed=true` 和有效 token |
| UAT-03 | 允许群 allow | PASS | `chat_id=oc_sales_weekly`，返回 `allowed=true` 和有效 token |
| UAT-04 | 私聊 `env` 输出 | PASS | 输出包含 `AUTH_OK=true`、`AUTH_ALLOWED=true`、`AUTH_TOKEN_TYPE=Bearer`、身份字段 |
| UAT-05 | 随机群 deny | PASS | `chat_id=oc_random_group`，返回 `deny_code=CHAT_SKILL_DENIED`，退出码 `10` |
| UAT-06 | deny JSON 协议 | PASS | deny 响应仅含 `request_id`、`deny_code`、`deny_message`，不含空 `auth_context` |
| UAT-07 | 离职用户 deny | PASS | `user_id=ou_left999` 时被拒绝，没有误判为 upstream error |
| UAT-08 | 缓存复用 | PASS | 首次 `auth/check` 后，未进入刷新窗口再次调用命中缓存，`cache_hit=true` |
| UAT-09 | 进入刷新窗口自动 refresh | PASS | 进入 refresh window 后调用 `/api/v1/token/refresh`，返回新 token |
| UAT-10 | refresh 后缓存更新 | PASS | 缓存项 `source` 更新为 `token_refresh` |
| UAT-11 | token 过期后重新鉴权 | PASS | 过期缓存不再 refresh，改走新的 `/api/v1/auth/check` |
| UAT-12 | 过期后缓存回写 | PASS | 重新鉴权后缓存项 `source=auth_check` |
| UAT-13 | 私聊/群聊缓存隔离 | PASS | 私聊和群聊 token 不同，缓存项独立存在 |
| UAT-14 | upstream fail-closed | PASS | 指向不可达 IAM 地址时退出码 `40`，stderr 前缀为 `AUTHCLI_UPSTREAM_FAILURE:` |
| UAT-15 | 全仓基础回归 | PASS | `go test ./...` 通过 |

## Executed Commands

真实 IAM UAT：

```bash
AUTHCLI_REAL_IAM_BASE_URL=http://127.0.0.1:8000 go test ./internal/app -run 'TestRealIAM' -v
```

基础回归：

```bash
go test ./...
```

健康检查：

```bash
curl -sS -i http://127.0.0.1:8000/healthz
```

## Detailed Test Output Snapshot

```text
=== RUN   TestRealIAMPrivateAllow
--- PASS: TestRealIAMPrivateAllow (0.12s)
=== RUN   TestRealIAMAllowedGroupJSON
--- PASS: TestRealIAMAllowedGroupJSON (0.07s)
=== RUN   TestRealIAMPrivateAllowEnvOutput
--- PASS: TestRealIAMPrivateAllowEnvOutput (0.06s)
=== RUN   TestRealIAMGroupDeny
--- PASS: TestRealIAMGroupDeny (0.02s)
=== RUN   TestRealIAMGroupDenyJSON
--- PASS: TestRealIAMGroupDenyJSON (0.01s)
=== RUN   TestRealIAMLeftUserDenied
--- PASS: TestRealIAMLeftUserDenied (0.01s)
=== RUN   TestRealIAMRefreshPath
--- PASS: TestRealIAMRefreshPath (0.13s)
=== RUN   TestRealIAMCacheReuseBeforeRefreshWindow
--- PASS: TestRealIAMCacheReuseBeforeRefreshWindow (0.06s)
=== RUN   TestRealIAMRefreshWindowRefreshesAndUpdatesCache
--- PASS: TestRealIAMRefreshWindowRefreshesAndUpdatesCache (0.13s)
=== RUN   TestRealIAMExpiredTokenReauths
--- PASS: TestRealIAMExpiredTokenReauths (0.17s)
=== RUN   TestRealIAMPrivateAndGroupCachesStayIsolated
--- PASS: TestRealIAMPrivateAndGroupCachesStayIsolated (0.13s)
=== RUN   TestRealIAMUpstreamFailClosed
--- PASS: TestRealIAMUpstreamFailClosed (0.00s)
PASS
```

## Findings

- 本轮 UAT 中，`authcli` 与真实 `iam-service` 的 Phase 1 主路径已可稳定联调。
- 本轮未观察到服务端 5xx、超时或协议不匹配问题。
- 服务端观察日志当前只记录到 access log 级别结论，没有新增需要客户端跟进的异常项。

## Not Covered In This Round

- `token/verify` 驱动的跨 chat 复用拒绝
- `token/revoke` 后客户端缓存即时失效
- `verify-sdk` 侧的 `IDENTITY_MISMATCH`、`CHAT_CONTEXT_MISMATCH`

这些仍属于后续 `verify-sdk` / 更完整 E2E 链路范围，不构成当前 Phase 1 联调阻塞。
