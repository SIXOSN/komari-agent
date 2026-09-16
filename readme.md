# SIXOSN Komari Agent

SIXOSN Komari Agent 是 [`SIXOSN/komari`](https://github.com/SIXOSN/komari) 的专用监控组件，由 SIXOSN 独立维护。

## 项目关系说明

本仓库基于 [`komari-monitor/komari-agent`](https://github.com/komari-monitor/komari-agent) 开发，不属于上游官方发行版，也不应默认获得上游支持或兼容性保证。许可证及依法需要保留的署名仍保存在 [LICENSE](./LICENSE) 与 Git 历史中。

## 安全边界

- 仅用于采集并上报服务器监控指标。
- 只与配套的 SIXOSN Komari 发行版连接。
- 不提供远程命令、Web SSH、网页终端或远程文件管理能力。
- 安装和更新来源限定为 SIXOSN 维护的发行渠道。

为减少不必要的信息暴露，本仓库不在公开简介中提供部署参数、内部协议或运维实现细节。请从配套 Komari 管理面板获取与具体节点对应的安装指引，并妥善保管节点凭据。
