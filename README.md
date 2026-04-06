# aily-skills-auth-authcli

`aily-skills-auth-authcli` 是企业 Agent Skill 鉴权平台的本地鉴权 CLI 仓，固定实现 `0.2.0` 的最小输入模型、缓存协议和 Skill 输出协议。

## Purpose

为 Agent 宿主环境提供统一的本地鉴权入口，固定接入 `aily-skills-auth-iam-service` 的最小鉴权闭环。

## Scope

- In scope:
  - `auth-cli check --skill <skill_id> --user-id <user_id>`
  - 最小输入模型：`user_id + skill_id`
  - 输入优先级：显式参数 > 环境变量 > 运行时上下文 > 本地配置
  - 本地 token 缓存与刷新
  - `/api/v1/auth/check`、`/api/v1/token/refresh` 客户端
  - `json` / `env` / `exit-code` 输出
  - `--help` 正式验收文案
- Out of scope:
  - 本地策略计算
  - JWT 服务端验证
  - 长期凭据托管
  - 管理控制台页面
  - `agent_id`、`chat_id` 作为核心授权输入

## Interfaces

- Inputs:
  - `--skill`
  - `--user-id`
  - `--format`
  - `--context-file`
  - `AUTHCLI_USER_ID`
  - `AUTHCLI_FORMAT`
  - `AUTHCLI_IAM_BASE_URL`
  - `AUTHCLI_TIMEOUT`
  - `AUTHCLI_CACHE_PATH`
  - `AUTHCLI_CONFIG_FILE`
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
- `examples/`: 固定联调样例配置与上下文
- `docs/`: 子仓内冻结文档与测试计划

## Runtime Semantics

`check` 命令固定执行以下流程：

1. 解析 `--skill`、`--user-id`、`--format`、`--context-file`
2. 按显式参数、环境变量、运行时上下文、本地配置的顺序组装输入
3. 命中未进入刷新窗口的缓存时直接返回
4. 命中刷新窗口但未过期的 token 时先调用 `/api/v1/token/refresh`
5. token 已过期或 refresh 返回重置错误码时删除缓存并重新调用 `/api/v1/auth/check`
6. deny、上游异常、上下文缺失都 fail-closed

## Local Development

1. 构建：`go build ./...`
2. 测试：`go test ./...`
3. 查看帮助：`go run ./cmd/auth-cli check --help`
4. 本地运行：

```bash
go run ./cmd/auth-cli check \
  --skill sales-analysis \
  --user-id ou_abc123 \
  --format json \
  --context-file ./examples/context-private.json
```

固定联调命令见 [minimal-integration.md](/Users/wenzhewang/workspace/codex/aily-skills-auth-authcli/docs/minimal-integration.md)。

## Beta Quick Start

1. 查看帮助：

```bash
go run ./cmd/auth-cli check --help
```

2. 执行 beta smoke：

```bash
./scripts/beta-smoke.sh
```

3. 执行最小本地调用：

```bash
go run ./cmd/auth-cli check \
  --skill sales-analysis \
  --user-id ou_abc123 \
  --format json \
  --context-file ./examples/context-private.json
```

4. 如需真实 IAM 联调：

```bash
AUTHCLI_REAL_IAM_BASE_URL=http://127.0.0.1:8000 ./scripts/real-iam-smoke.sh
```

## Distribution

`0.2.0` 的构建产物、分发方式和宿主机部署要求见 [release-and-distribution.md](/Users/wenzhewang/workspace/codex/aily-skills-auth-authcli/docs/release-and-distribution.md)。

官方宿主机安装说明见 [host-installation.md](/Users/wenzhewang/workspace/codex/aily-skills-auth-authcli/docs/host-installation.md)。

推荐安装命令：

```bash
curl -fsSL https://github.com/wodenwang/aily-skills-auth-authcli/releases/download/v0.2.0/install-authcli.sh \
  | sh -s -- --version v0.2.0 --install-dir /usr/local/bin
```

安装后最小离线校验：

```bash
auth-cli check
```

期望：

- 退出码 `20`
- stderr 包含 `AUTHCLI_INVALID_INPUT: missing required flag: --skill`

升级入口：

```bash
curl -fsSL https://github.com/wodenwang/aily-skills-auth-authcli/releases/download/v0.2.0/install-authcli.sh \
  | sh -s -- --version v0.2.0 --install-dir /usr/local/bin
```

beta 发布前检查见 [beta-release-checklist.md](/Users/wenzhewang/workspace/codex/aily-skills-auth-authcli/docs/beta-release-checklist.md)。

稳定 stderr 前缀：

- `AUTHCLI_INVALID_INPUT:`
- `AUTHCLI_CACHE_FAILURE:`
- `AUTHCLI_UPSTREAM_FAILURE:`
- `AUTHCLI_INTERNAL_ERROR:`

## Upstream Contracts

本仓字段命名和行为以上游冻结契约为唯一准绳：

- [authcli-output.md](/Users/wenzhewang/workspace/codex/aily-skills-auth/docs/contracts/authcli-output.md)
- [token-cache.md](/Users/wenzhewang/workspace/codex/aily-skills-auth/docs/contracts/token-cache.md)
- [authcli.md](/Users/wenzhewang/workspace/codex/aily-skills-auth/docs/modules/authcli.md)
