# SIXOSN Komari Agent

SIXOSN Komari Agent is the dedicated monitoring agent for [`SIXOSN/komari`](https://github.com/SIXOSN/komari). It reports host metrics but does not include remote command execution, Web SSH, browser-terminal, or remote file-management functionality.

## 项目关系说明

本仓库基于 [`komari-monitor/komari-agent`](https://github.com/komari-monitor/komari-agent) 开发，由 SIXOSN 独立维护，不属于上游官方发行版。许可证及依法需要保留的署名仍保存在 [LICENSE](./LICENSE) 与 Git 历史中。

本 Agent 与 SIXOSN Komari 服务端进行双向发行版身份校验：它会拒绝非 SIXOSN 服务端，SIXOSN 服务端也会拒绝其他发行版 Agent。自动更新与安装脚本只使用 `SIXOSN/komari-agent`。

## Linux 安装

从管理面板复制节点地址与 Token，然后执行：

```bash
wget -qO- 'https://raw.githubusercontent.com/SIXOSN/komari-agent/refs/heads/main/install.sh' \
  | sudo bash -s -- \
  --install-version snapshot \
  -e 'https://你的面板地址' \
  -t '节点TOKEN'
```

安装后检查：

```bash
systemctl status komari-agent --no-pager
journalctl -u komari-agent -n 50 --no-pager
```

## Docker

```bash
docker run -d \
  --name komari-agent \
  --restart always \
  ghcr.io/sixosn/komari-agent:snapshot \
  -e 'https://你的面板地址' \
  -t '节点TOKEN'
```

原生 systemd 部署更适合完整采集宿主机指标。

## 配置方式

Agent 支持命令行参数、环境变量和 JSON 配置文件。配置优先级从低到高依次为默认值、命令行参数、环境变量、JSON 配置文件。

```json
{
  "endpoint": "https://example.com",
  "token": "your-token",
  "interval": 3,
  "disable_auto_update": false,
  "ignore_unsafe_cert": false
}
```

常用参数：

| JSON 字段 | 环境变量 | 命令行参数 | 说明 |
| --- | --- | --- | --- |
| `endpoint` | `AGENT_ENDPOINT` | `--endpoint`, `-e` | SIXOSN Komari 面板地址 |
| `token` | `AGENT_TOKEN` | `--token`, `-t` | 节点私密 Token |
| `interval` | `AGENT_INTERVAL` | `--interval`, `-i` | 采集间隔，单位秒 |
| `disable_auto_update` | `AGENT_DISABLE_AUTO_UPDATE` | `--disable-auto-update` | 禁用自动更新 |
| `ignore_unsafe_cert` | `AGENT_IGNORE_UNSAFE_CERT` | `--ignore-unsafe-cert`, `-u` | 忽略不安全证书 |
| `include_nics` | `AGENT_INCLUDE_NICS` | `--include-nics` | 仅统计指定网卡 |
| `exclude_nics` | `AGENT_EXCLUDE_NICS` | `--exclude-nics` | 排除指定网卡 |
| `include_mountpoints` | `AGENT_INCLUDE_MOUNTPOINTS` | `--include-mountpoint` | 仅统计指定挂载点 |
| `auto_discovery_key` | `AGENT_AUTO_DISCOVERY_KEY` | `--auto-discovery` | 自动发现密钥 |
| `custom_dns` | `AGENT_CUSTOM_DNS` | `--custom-dns` | 自定义 DNS |
| `enable_gpu` | `AGENT_ENABLE_GPU` | `--gpu` | 启用详细 GPU 监控 |
| `prefer_ip_version` | `AGENT_PREFER_IP_VERSION` | `--prefer-ip-version` | 优先使用 IPv4 或 IPv6 |

完整参数：

```bash
/opt/komari/agent --help
```

请勿公开节点 Token，也不要在未获授权的服务器上部署本 Agent。
