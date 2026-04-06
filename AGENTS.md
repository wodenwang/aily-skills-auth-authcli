# AGENTS

## Repo Role

本仓只实现 `authcli` 本地鉴权能力，负责：

- `check` 命令
- 身份采集与输入优先级解析
- IAM 鉴权与刷新调用
- token 缓存、失效和 fail-closed 行为
- Skill 侧输出协议
- `--help` 正式验收文案

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

若与主控仓冲突：

- 本仓执行边界以当前文件为准
- 跨仓公共接口与冻结语义以主控仓契约为准
- 先修文档，再修实现，不允许用实现反向定义契约

## Ownership

- Lead: 维护仓库边界、公开命令面和版本节奏
- Runtime: 维护命令行行为、缓存与重试
- API: 维护 IAM 接口客户端与错误映射
- Quality: 维护 CLI 行为测试与联调样例
- Reviewer: 只按冻结文档检查偏差，不定义新接口

## Frozen Interfaces

- `auth-cli check --skill <skill_id> --user-id <user_id>`
- `auth-cli check --help`
- `POST /api/v1/auth/check`
- `POST /api/v1/token/refresh`
- AuthCLI 输出协议：`json` / `env` / `exit-code`
- Token 缓存文件格式与失效规则

## 0.2.0 Hard Constraints

- 最小输入模型固定为 `user_id + skill_id`
- 输入优先级固定为 `显式参数 > 环境变量 > 运行时上下文 > 本地配置`
- 缓存键只包含 `user_id + skill_id`
- `agent_id`、`chat_id` 不得作为核心授权输入、缓存键或公开输出字段
- deny 与 upstream error 必须区分
- 所有错误场景必须 fail-closed
- `--help` 是正式验收面

## Review Mechanism

- reviewer 输入只能来自当前仓 `AGENTS.md`、本仓文档、主控仓冻结契约和当前阶段产物
- reviewer 输出格式固定为：问题描述 -> 影响范围 -> 依据文档 -> 建议动作
- reviewer 每个阶段至少检查一次：命令面、输入优先级、缓存键、输出协议、`--help`、测试口径
- 如果发现主控仓 `0.2.0` 冻结点与本仓实现冲突，必须先修本文档和仓内镜像文档，再继续编码
- sub-agent 之间若有交互，只能引用文档，不允许靠临时口头约定
- 所有 sub-agent 必须与主 agent 使用同一模型能力，不允许切到 mini、lite 或其他降配模型

## Beta Goals

`0.2.0-beta` 阶段目标：

- 把 `authcli` 收口成可安装、可升级、可联调的 CLI
- 固定 `check` 命令面、`--help` 文案、缓存与 refresh 行为
- 固定 README beta 快速开始、安装脚本、smoke test 和 integration test 入口
- 形成 beta release checklist，作为发布前 gate 的直接依据

beta 阶段不引入：

- 新的授权输入字段
- 新的输出格式
- 新的缓存键维度
- 为试点场景定制的非冻结行为

## Beta Review Gate

beta 进入发布前，reviewer 必须检查：

- 是否仍残留旧输入模型，包括 `agent_id`、`chat_id` 作为核心输入或缓存键
- README、`--help`、输出协议和安装文档是否一致
- 安装脚本的默认版本、升级入口和离线校验是否一致
- smoke test / integration test 命令是否存在、可执行、可回归
- beta release checklist 是否覆盖命令面、缓存、帮助文案、安装、测试和联调

若 reviewer 发现上述任一项不一致，beta gate 视为未通过。

## Beta Release Requirements

- 必须保留 `auth-cli check --skill <skill_id> --user-id <user_id>` 作为唯一公开命令面
- 必须保留 `json` / `env` / `exit-code` 三种冻结输出协议
- 安装脚本必须支持安装、覆盖升级和离线校验
- README 必须提供 beta 快速开始、安装、升级、smoke 和 integration test 入口
- 发布前必须至少通过：
  - `go test ./...`
  - `go run ./cmd/auth-cli check --help`
  - beta smoke script
  - 真实 IAM smoke 或明确记录阻塞原因

## Definition of Done

- `check` 命令可稳定运行
- allow、deny、refresh 三类主路径可回归
- fail-closed 行为稳定
- 输出协议和缓存协议无漂移
- `--help` 文案完整且可回归

## Current Phase Status

截至 `2026-04-06`，本仓进入 `0.2.0-beta` 收口阶段，重点是把主控仓 `0.2.0` 冻结契约落实为可安装、可升级、可联调的 CLI：

- `auth-cli check` 是唯一公开命令面
- 最小输入模型已收敛到 `user_id + skill_id`
- 本地 token 缓存、刷新窗口判断、过期重鉴权和 fail-closed 行为必须保持稳定
- `/api/v1/auth/check` 与 `/api/v1/token/refresh` 继续对齐真实 `iam-service`
- `json`、`env`、`exit-code` 三种输出协议必须对齐当前冻结契约
- deny 响应与 upstream error 必须持续明确区分
- 安装脚本、README beta 快速开始、smoke test 和 release checklist 必须可回归

当前仓明确仍不负责：

- `/api/v1/token/verify` 服务端实现
- `/api/v1/token/revoke` 服务端实现
- 生产级发布、监控和运维流程
