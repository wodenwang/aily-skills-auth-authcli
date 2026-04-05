# AGENTS

## Repo Role

本仓只实现 `authcli` 本地鉴权能力，负责：

- `check` 命令
- 身份采集与输入优先级解析
- IAM 鉴权与刷新调用
- token 缓存、失效和 fail-closed 行为
- Skill 侧输出协议

本仓不负责：

- 服务端策略计算
- JWT 服务端验证链路
- 管理控制台
- 多语言 SDK

## Document Priority

1. `AGENTS.md`
2. `README.md`
3. `docs/contracts/**`
4. `docs/**`
5. 上游冻结契约 `/Users/wenzhewang/workspace/codex/aily-skills-auth/docs/**`

## Ownership

- Lead: 维护仓库边界与版本节奏
- Runtime: 维护命令行行为、缓存与重试
- API: 维护 IAM 接口客户端与错误映射
- Quality: 维护 CLI 行为测试与联调样例

## Frozen Interfaces

- `auth-cli check --skill <skill_id>`
- `POST /api/v1/auth/check`
- `POST /api/v1/token/refresh`
- AuthCLI 输出协议
- Token 缓存文件格式与失效规则

## Definition of Done

- `check` 命令可稳定运行
- allow、deny、refresh 三类主路径可回归
- fail-closed 行为稳定
- 输出协议和缓存协议无漂移

## Current Phase Status

截至 `2026-04-05`，本仓 Phase 1 已完成，并已作为完整 E2E 链路的稳定上游输入：

- `auth-cli check` 已实现并固定为唯一公开命令面
- 输入优先级已按 `显式参数 > 环境变量 > 运行时上下文 > 本地配置` 落地
- 本地 token 缓存、刷新窗口判断、过期重鉴权和 fail-closed 行为已实现
- `/api/v1/auth/check` 与 `/api/v1/token/refresh` 已完成真实 `iam-service` 联调
- `json`、`env`、`exit-code` 三种输出协议已对齐当前冻结契约
- deny 响应与 upstream error 已明确区分
- 真实 IAM UAT 已通过，报告见 [docs/authcli-iam-uat-report-2026-04-05.md](/Users/wenzhewang/workspace/codex/aily-skills-auth-authcli/docs/authcli-iam-uat-report-2026-04-05.md)

当前仓明确仍不负责：

- `/api/v1/token/verify` 服务端实现
- `/api/v1/token/revoke` 服务端实现
- 生产级发布、监控和运维流程

当前阶段后续重点：

- 保持输出协议和缓存协议稳定，不为试点接入引入破坏性变更
- 配合试点 Skill 和完整 E2E 回归
