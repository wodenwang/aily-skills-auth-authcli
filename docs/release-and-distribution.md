# Release And Distribution

## Goal

冻结 `aily-skills-auth-authcli` 在 `0.1.0-alpha` 阶段的发布物、分发方式、宿主机部署要求和运行约束。

本文件只定义 alpha 分发方案，不直接执行构建、打包或发布。

## Release Target

- 版本：`0.1.0-alpha`
- Git tag：`v0.1.0-alpha`
- Release 入口：GitHub Release

## Artifacts

- 源码 tag：`v0.1.0-alpha`
- GitHub Release
- Darwin 二进制压缩包
- Linux 二进制压缩包

alpha 阶段固定约束：

- 使用 `go build` 产出单文件二进制
- 不做 Homebrew
- 不做平台安装器

## Distribution Profile

- 通过 GitHub Release 附件下载二进制压缩包
- 在 Agent 宿主机手动安装到固定路径
- 运行时通过环境变量或配置文件指向 IAM 地址

## Host Deployment

`authcli` 不以容器形式独立部署，而是作为 Skill 宿主机上的本地可执行文件安装。

宿主机至少需要准备：

- 固定可执行路径
- 可写缓存目录
- 可读配置文件路径
- 指向 `iam-service` 的 `AUTHCLI_IAM_BASE_URL`

## Runtime Requirements

- 默认失败关闭
- 缓存路径必须可写
- 输出协议冻结，不因试点 skill 定制而变更
- Skill 不得绕过 `auth-cli check`

## Build And Release Flow

1. 运行 `go test ./...`
2. 使用 `go build` 为 Darwin / Linux 构建二进制
3. 打包压缩产物
4. 创建 Git tag `v0.1.0-alpha`
5. 创建 GitHub Release 并上传压缩包

Release 说明至少应包含：

- 支持的平台
- 固定命令入口 `auth-cli check --skill <skill_id>`
- 必填环境变量
- 已冻结输出协议
- 失败关闭约束

## Installation Notes

推荐在宿主机上固定：

- 二进制安装目录
- `AUTHCLI_CACHE_PATH`
- `AUTHCLI_CONFIG_FILE`
- `AUTHCLI_IAM_BASE_URL`

alpha 阶段不提供自动升级器，升级依赖重新下载并替换二进制。

## Verification

- `go build` 可为 Darwin / Linux 产出单文件二进制
- `auth-cli check --skill <skill_id>` 命令可执行
- 缓存路径与配置路径在宿主机可用
- 与 `iam-service` 的联调路径保持可用

## Frozen Interface

- `auth-cli check --skill <skill_id>`

## Out Of Scope

- Homebrew tap
- 自动安装器
- Windows 发行物
