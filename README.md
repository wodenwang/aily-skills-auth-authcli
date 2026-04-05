# aily-skills-auth-authcli

`aily-skills-auth-authcli` 是企业 Agent Skill 鉴权平台的本地鉴权 CLI 仓，负责身份采集、鉴权请求、token 缓存与刷新，以及向 Skill 输出冻结协议。

## Purpose

为 Agent 宿主环境提供统一的本地鉴权入口，固定接入 `aily-skills-auth-iam-service` 的最小鉴权闭环。

## Scope

- In scope:
  - `auth-cli check`
  - 四元组身份采集
  - 本地 token 缓存与刷新
  - `/api/v1/auth/check`、`/api/v1/token/refresh` 客户端
  - `json` / `env` / `exit-code` 输出
- Out of scope:
  - 本地策略计算
  - JWT 服务端验证
  - 长期凭据托管
  - 管理控制台页面

## Interfaces

- Inputs:
  - 显式参数
  - 环境变量
  - Agent 运行时上下文
  - 本地配置
- Outputs:
  - `json`
  - `env`
  - `exit-code`
- Dependencies:
  - `aily-skills-auth-iam-service`
  - 上游规范仓 `/Users/wenzhewang/workspace/codex/aily-skills-auth`

## Project Layout

- `cmd/auth-cli/`: CLI 入口
- `internal/app/`: 运行编排
- `internal/auth/`: IAM 客户端与契约类型
- `internal/cache/`: token 缓存文件读写和失效判断
- `internal/cli/`: 参数解析与输入优先级处理
- `internal/output/`: `json` / `env` / `exit-code` 输出
- `docs/`: 子仓内冻结文档与测试计划

## Runtime Semantics

`check` 命令固定执行以下流程：

1. 解析 `--skill`、`--user-id`、`--agent-id`、`--chat-id`、`--format`、`--context-file`
2. 按显式参数、环境变量、运行时上下文、本地配置的顺序组装输入
3. 命中未进入刷新窗口的缓存时直接返回
4. 命中刷新窗口但未过期的 token 时先调用 `/api/v1/token/refresh`
5. 缓存不可用、token 已过期或刷新失败时重新调用 `/api/v1/auth/check`
6. deny、上游异常、上下文缺失都 fail-closed

## Local Development

1. 构建：`go build ./...`
2. 测试：`go test ./...`
3. 本地运行：

```bash
go run ./cmd/auth-cli check \
  --skill sales-analysis \
  --user-id ou_abc123 \
  --agent-id host-vm-a1b2c3d4 \
  --format json
```

默认环境变量：

- `AUTHCLI_IAM_BASE_URL`: IAM 服务地址，默认 `http://127.0.0.1:8000`
- `AUTHCLI_TIMEOUT`: 请求超时，默认 `5s`
- `AUTHCLI_CACHE_PATH`: token 缓存路径，默认 `~/.aily-skills-auth/cache/tokens.json`
- `AUTHCLI_CONFIG_FILE`: 本地配置文件路径，默认 `~/.aily-skills-auth/config.json`
- `AUTHCLI_USER_ID`
- `AUTHCLI_AGENT_ID`
- `AUTHCLI_CHAT_ID`
- `AUTHCLI_FORMAT`

## Upstream Contracts

本仓字段命名和行为以上游冻结契约为唯一准绳：

- [auth-check.md](/Users/wenzhewang/workspace/codex/aily-skills-auth/docs/contracts/auth-check.md)
- [token-refresh.md](/Users/wenzhewang/workspace/codex/aily-skills-auth/docs/contracts/token-refresh.md)
- [authcli-output.md](/Users/wenzhewang/workspace/codex/aily-skills-auth/docs/contracts/authcli-output.md)
- [token-cache.md](/Users/wenzhewang/workspace/codex/aily-skills-auth/docs/contracts/token-cache.md)
- [domain-model.md](/Users/wenzhewang/workspace/codex/aily-skills-auth/docs/contracts/domain-model.md)
