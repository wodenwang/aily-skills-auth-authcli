# Token Cache Contract

本仓镜像上游冻结协议，作为缓存读写、测试和后续兼容性检查依据。

上游来源：

- [token-cache.md](/Users/wenzhewang/workspace/codex/aily-skills-auth/docs/contracts/token-cache.md)

## File Location

- 默认路径：`~/.aily-skills-auth/cache/tokens.json`
- 支持通过配置覆盖，但文件内容结构必须保持一致

## Cache Key

缓存键固定由以下字段拼接：

- `user_id`
- `skill_id`

`0.2.0` 不再把 `agent_id`、`chat_id` 作为缓存键组成部分。

## Invalidation Rules

- 当前时间早于 `refresh_before_at` 时，允许直接命中缓存
- 当前时间大于等于 `refresh_before_at` 且早于 `expires_at` 时，必须先调用 `/api/v1/token/refresh`
- 当前时间大于等于 `expires_at` 时，禁止刷新旧 token，必须重新调用 `/api/v1/auth/check`
- 上游返回 `TOKEN_REVOKED`、`TOKEN_INVALID`、`TOKEN_EXPIRED` 时，必须立即删除对应缓存项
- `user_id` 或 `skill_id` 任一变化时，禁止复用旧缓存项

## Write Rules

- 使用原子写入，避免并发写坏文件
- 缓存文件只保存短期 access token，不保存长期用户凭据
- 不允许在缓存中补充本地推导的权限结果
