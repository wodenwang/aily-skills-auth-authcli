# Beta Release Checklist

## Goal

作为 `0.2.0-beta` 发布前的最终 gate，确保 `authcli` 已达到可安装、可升级、可联调状态。

## Interface Freeze

- [ ] 公开命令面仍为 `auth-cli check --skill <skill_id> --user-id <user_id>`
- [ ] `--help` 文案与 README、主控仓冻结契约一致
- [ ] 输出协议仍为 `json` / `env` / `exit-code`
- [ ] 仓内无把 `agent_id`、`chat_id` 作为核心输入、缓存键或公开输出字段的残留实现

## Cache And Refresh

- [ ] 缓存键只包含 `user_id + skill_id`
- [ ] 缓存文件版本为 `2`
- [ ] 命中缓存、进入 refresh 窗口、refresh reset code、过期重鉴权四类行为已回归
- [ ] 所有错误场景保持 fail-closed

## Install And Upgrade

- [ ] `scripts/install-authcli.sh` 默认版本、安装命令和 README 一致
- [ ] 安装脚本支持覆盖升级
- [ ] 安装脚本离线校验命令可执行
- [ ] `docs/host-installation.md` 和 `docs/release-and-distribution.md` 的安装升级口径一致

## README And Docs

- [ ] README 包含 beta quick start
- [ ] README 包含安装、升级、smoke、integration test 入口
- [ ] `docs/minimal-integration.md` 提供 allow、deny、refresh 联调路径
- [ ] `docs/test-plan.md` 覆盖输入优先级、缓存、输出、fail-closed

## Verification Commands

- [ ] `go test ./...`
- [ ] `go run ./cmd/auth-cli check --help`
- [ ] `./scripts/beta-smoke.sh`
- [ ] `./scripts/real-iam-smoke.sh` 或明确记录阻塞原因

## Reviewer Gate

reviewer 必须按以下格式输出一次 gate 结果：

- 问题描述
- 影响范围
- 依据文档
- 建议动作
