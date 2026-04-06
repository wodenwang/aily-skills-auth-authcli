# Host Installation

## Goal

冻结 `authcli` 在 Skill 宿主机上的安装、校验、升级和回滚方式。

本文件面向：

- 安装 Skill 的宿主机维护者
- 试点环境的部署执行者
- 需要把 `auth-cli` 放入固定运行环境的运维人员

## Official Installation Path

唯一推荐路径：

- 通过 GitHub Release 分发二进制压缩包
- 通过随 Release 发布的 `install-authcli.sh` 安装脚本完成下载和落盘

不作为官方安装入口：

- Homebrew
- 第三方包管理平台
- 容器镜像
- 手写 `curl | sh` 指向非 tag 的分支脚本

## Supported Platforms

当前固定支持：

- `darwin-arm64`
- `linux-amd64`

对应 Release 附件名称：

- `auth-cli-darwin-arm64.tar.gz`
- `auth-cli-linux-amd64.tar.gz`
- `install-authcli.sh`

## Quick Install

安装指定版本到默认目录 `~/.local/bin`：

```bash
curl -fsSL https://github.com/wodenwang/aily-skills-auth-authcli/releases/download/v0.1.0-alpha/install-authcli.sh \
  | sh -s -- --version v0.1.0-alpha
```

安装到固定系统目录：

```bash
curl -fsSL https://github.com/wodenwang/aily-skills-auth-authcli/releases/download/v0.1.0-alpha/install-authcli.sh \
  | sh -s -- --version v0.1.0-alpha --install-dir /usr/local/bin
```

如果宿主机不允许管道执行，使用两步式安装：

```bash
curl -fsSL -o /tmp/install-authcli.sh \
  https://github.com/wodenwang/aily-skills-auth-authcli/releases/download/v0.1.0-alpha/install-authcli.sh

sh /tmp/install-authcli.sh --version v0.1.0-alpha --install-dir /usr/local/bin
```

## Verification

安装完成后至少验证以下内容：

1. 二进制在固定路径存在并可执行
2. 命令 `auth-cli check` 返回冻结错误前缀
3. 宿主机环境变量可正确指向 IAM

最小验证命令：

```bash
command -v auth-cli
auth-cli check
```

期望结果：

- 退出码：`20`
- stderr 包含：`AUTHCLI_INVALID_INPUT: missing required flag: --skill`

这是安装后的离线校验，不依赖 IAM 在线。

## Runtime Preparation

推荐在宿主机固定：

- `PATH` 中包含 `auth-cli`
- `AUTHCLI_IAM_BASE_URL`
- `AUTHCLI_CACHE_PATH`
- `AUTHCLI_CONFIG_FILE`

示例：

```bash
export AUTHCLI_IAM_BASE_URL=http://127.0.0.1:8000
export AUTHCLI_CACHE_PATH="$HOME/.aily-skills-auth/cache/tokens.json"
export AUTHCLI_CONFIG_FILE="$HOME/.aily-skills-auth/config.json"
```

## Upgrade

升级规则固定为：

1. 用新版本的 `install-authcli.sh` 指向目标 tag
2. 覆盖原二进制
3. 重新执行离线校验
4. 再执行一条真实 `check --skill ...` 验证

示例：

```bash
curl -fsSL https://github.com/wodenwang/aily-skills-auth-authcli/releases/download/v0.1.0-alpha/install-authcli.sh \
  | sh -s -- --version v0.1.0-alpha --install-dir /usr/local/bin
```

升级不修改：

- `AUTHCLI_CACHE_PATH`
- `AUTHCLI_CONFIG_FILE`
- Skill 调用命令入口

## Rollback

回滚方式固定为重新安装旧 tag：

```bash
curl -fsSL https://github.com/wodenwang/aily-skills-auth-authcli/releases/download/v0.1.0-alpha/install-authcli.sh \
  | sh -s -- --version v0.1.0-alpha --install-dir /usr/local/bin
```

回滚后必须重新执行 Verification 小节中的离线校验。

## Manual Fallback

若宿主机不能执行安装脚本，可手动下载并落盘：

```bash
mkdir -p /tmp/authcli-release /usr/local/bin
curl -fsSL -o /tmp/authcli-release/auth-cli-darwin-arm64.tar.gz \
  https://github.com/wodenwang/aily-skills-auth-authcli/releases/download/v0.1.0-alpha/auth-cli-darwin-arm64.tar.gz
tar -xzf /tmp/authcli-release/auth-cli-darwin-arm64.tar.gz -C /tmp/authcli-release
install /tmp/authcli-release/auth-cli-darwin-arm64 /usr/local/bin/auth-cli
```

手动安装后仍然必须执行 Verification。
