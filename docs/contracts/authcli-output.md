# AuthCLI Output Contract

本仓镜像上游冻结协议，作为本地实现与测试的直接依据。

上游来源：

- [authcli-output.md](/Users/wenzhewang/workspace/codex/aily-skills-auth/docs/contracts/authcli-output.md)

## Commands

- `auth-cli check`
- `auth-cli check --help`

## Input Fields

- `user_id`
- `skill_id`
- `context_file`

输入优先级：

1. 显式参数
2. 环境变量
3. Agent 运行时上下文
4. 本地配置

## Formats

- `json`
- `env`
- `exit-code`

未显式指定时默认输出 `json`。

## Exit Codes

- `0`: allowed
- `10`: denied by policy
- `20`: invalid input
- `30`: cache read/write failure
- `40`: upstream unavailable or timeout
- `50`: unexpected internal error

## Rules

- deny 和 error 必须区分，禁止把上游异常伪装成 deny
- 任一错误场景必须 fail-closed，不输出伪造 token
- `env` 模式仅在 allow 场景输出 token 字段
- `json` 和 `env` 都必须保留 `request_id` 以便审计串联
- `agent_id`、`chat_id` 不属于 `0.2.0` 核心输入，也不得出现在公开输出字段中
