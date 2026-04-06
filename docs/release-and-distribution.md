# Release And Distribution

## Goal

冻结 `aily-skills-auth-authcli` 在 `0.2.0` 阶段的发布物、分发方式、宿主机部署要求和运行约束。

本文件只定义分发方案，不直接执行构建、打包或发布。

## Release Target

- 版本：`0.2.0`
- Git tag：`v0.2.0`
- Release 入口：GitHub Release

## Artifacts

- 源码 tag：`v0.2.0`
- GitHub Release
- Darwin 二进制压缩包
- Linux 二进制压缩包
- Release 安装脚本 `install-authcli.sh`

固定约束：

- 使用 `go build` 产出单文件二进制
- 不做 Homebrew
- 不做第三方平台分发

## Distribution Profile

- 官方入口是 GitHub Release
- 官方安装方式是 Release 附带的 `install-authcli.sh`
- 手动下载二进制压缩包是兜底方式，不是首选路径
- 运行时通过环境变量或配置文件指向 IAM 地址

## Host Deployment

`authcli` 不以容器形式独立部署，而是作为 Skill 宿主机上的本地可执行文件安装。

宿主机至少需要准备：

- 固定可执行路径
- 可写缓存目录
- 可读配置文件路径
- 指向 `iam-service` 的 `AUTHCLI_IAM_BASE_URL`

标准安装、校验和升级流程见 [host-installation.md](/Users/wenzhewang/workspace/codex/aily-skills-auth-authcli/docs/host-installation.md)。

## Runtime Requirements

- 默认失败关闭
- 缓存路径必须可写
- 输出协议冻结，不因试点 skill 定制而变更
- Skill 不得绕过 `auth-cli check`

## Build And Release Flow

1. 运行 `go test ./...`
2. 运行 `./scripts/beta-smoke.sh`
3. 使用 `go build` 为 Darwin / Linux 构建二进制
4. 打包压缩产物
5. 准备 `scripts/install-authcli.sh`
6. 创建 Git tag `v0.2.0`
7. 创建 GitHub Release 并上传压缩包和安装脚本

Release 说明至少应包含：

- 支持的平台
- 官方安装命令
- 固定命令入口 `auth-cli check --skill <skill_id> --user-id <user_id>`
- 必填环境变量
- 已冻结输出协议
- 失败关闭约束

## Installation Notes

推荐在宿主机上固定：

- 二进制安装目录
- `AUTHCLI_CACHE_PATH`
- `AUTHCLI_CONFIG_FILE`
- `AUTHCLI_IAM_BASE_URL`

官方安装命令：

```bash
curl -fsSL https://github.com/wodenwang/aily-skills-auth-authcli/releases/download/v0.2.0/install-authcli.sh \
  | sh -s -- --version v0.2.0 --install-dir /usr/local/bin
```

升级方式：

- 重新执行安装脚本并指定目标 tag
- 覆盖本地 `auth-cli`
- 重新执行离线校验和一条真实 `check`

## Verification

- `go build` 可为 Darwin / Linux 产出单文件二进制
- `auth-cli check --skill <skill_id> --user-id <user_id>` 命令可执行
- `auth-cli check` 在离线场景下返回冻结错误前缀和退出码 `20`
- 缓存路径与配置路径在宿主机可用
- 与 `iam-service` 的联调路径保持可用

## Frozen Interface

- `auth-cli check --skill <skill_id> --user-id <user_id>`

## Beta Gate

发布前必须通过 [beta-release-checklist.md](/Users/wenzhewang/workspace/codex/aily-skills-auth-authcli/docs/beta-release-checklist.md)。

## Out Of Scope

- Homebrew tap
- 第三方平台安装器
- Windows 发行物
