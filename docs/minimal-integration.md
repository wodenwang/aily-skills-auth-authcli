# Minimal Integration Guide

## Goal

提供 `authcli` 与本地 `iam-service` 的最小可复现联调路径，覆盖 allow、deny、refresh 三条主链路。

## Prerequisites

- `aily-skills-auth-iam-service` 已可本地运行
- `go test ./...` 与 `go build ./...` 在本仓通过
- 本地可访问 `http://127.0.0.1:8000`

## Local Setup

1. 启动 IAM Service：

```bash
cd /Users/wenzhewang/workspace/codex/aily-skills-auth-iam-service
docker compose up -d postgres redis
uv run alembic upgrade head
uv run uvicorn aily_auth_iam_service.main:app --reload
```

2. 在本仓准备样例配置：

```bash
cp examples/config.example.json ~/.aily-skills-auth/config.json
```

3. 可选：清理旧缓存：

```bash
rm -f ~/.aily-skills-auth/cache/tokens.json
```

## Fixed Commands

私聊 allow：

```bash
go run ./cmd/auth-cli check \
  --skill sales-analysis \
  --user-id ou_abc123 \
  --agent-id host-vm-a1b2c3d4 \
  --format json \
  --context-file ./examples/context-private.json
```

群聊 allow：

```bash
go run ./cmd/auth-cli check \
  --skill sales-analysis \
  --user-id ou_abc123 \
  --agent-id host-vm-a1b2c3d4 \
  --chat-id oc_sales_weekly \
  --format json \
  --context-file ./examples/context-group.json
```

群聊 deny：

```bash
go run ./cmd/auth-cli check \
  --skill sales-analysis \
  --user-id ou_abc123 \
  --agent-id host-vm-a1b2c3d4 \
  --chat-id oc_random_group \
  --format json \
  --context-file ./examples/context-group.json
```

刷新验证：

1. 先执行一次群聊 allow，生成缓存。
2. 等待 token 进入 `refresh_before` 窗口后重复执行同一命令。
3. 期望返回新的 token，且缓存文件中 `source` 更新为 `token_refresh`。

## Demo Skill Baseline

`demo-skill` 集成时只允许依赖：

- `auth-cli check --format json`
- `AUTH_*` 环境变量输出
- 退出码 `0/10/20/30/40/50`

禁止依赖：

- `authcli` 内部 Go 包
- 缓存文件内部结构
- 非冻结 stderr 文本之外的临时调试输出

## Stable stderr Prefixes

- `AUTHCLI_INVALID_INPUT:`
- `AUTHCLI_CACHE_FAILURE:`
- `AUTHCLI_UPSTREAM_FAILURE:`
- `AUTHCLI_INTERNAL_ERROR:`

## Real IAM Smoke

在 `iam-service` 已启动并带有 demo fixture 时，可直接运行：

```bash
AUTHCLI_REAL_IAM_BASE_URL=http://127.0.0.1:8000 \
go test ./... -run TestRealIAM
```

或使用仓库脚本：

```bash
./scripts/real-iam-smoke.sh
```
