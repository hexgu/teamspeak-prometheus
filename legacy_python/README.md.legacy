# TeamSpeak Prometheus Exporter

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)
[![Python](https://img.shields.io/badge/python-3.12+-blue.svg)](https://www.python.org/)

这是一个用于监控 TeamSpeak 3 服务器的 Prometheus 导出器。它通过 TS3 Server Query 接口获取服务器指标和在线玩家详细信息，并提供美观的 Grafana 展示面板。

## 🌟 功能特性

- **多服务器支持**：可同时监控多个 TeamSpeak 服务器。
- **详尽的指标**：
  - 服务器状态：在线人数、带宽、丢包率、运行时间等。
  - 玩家详情：昵称、IP、国家、客户端版本、静音状态、在线时长等。
- **智能过滤**：自动排除 ServerQuery 等非真实玩家连接。
- **开箱即用**：提供预配置的 Grafana 面板和 Docker 部署方案。

## 🚀 快速开始

### 使用 Docker 部署 (推荐)

1. **克隆仓库**
   ```bash
   git clone https://github.com/hexgu/teamspeak-prometheus/
   cd teamspeak-prometheus
   ```

2. **配置服务器**
   修改 `config.yaml`，添加你的 TeamSpeak 服务器信息：
   ```yaml
   servers:
     - name: "我的服务器"
       host: "127.0.0.1"
       port: 10011
       username: "serveradmin"
       password: "你的密码"
   ```

3. **启动服务**
   ```bash
   docker compose up -d
   ```

### 本地开发环境

项目使用 [uv](https://github.com/astral-sh/uv) 进行包管理：
```bash
uv sync
uv run python app.py
```

## 📊 指标与监控

- **导出器地址**: `http://localhost:8001/metrics`
- **Prometheus 地址**: `http://localhost:9090`
- **Grafana 导入**:
  1. 打开 Grafana，进入 **Dashboards -> Import**。
  2. 上传仓库中的 `grafana-dashboard.json`。
  3. 关联你的 Prometheus 数据源。

## ⚙️ 配置说明

| 配置项 | 说明 | 默认值 |
| :--- | :--- | :--- |
| `metrics_port` | 导出器监听端口 | 8000 |
| `read_interval` | 抓取频率（秒） | 60 |
| `server_name` | 自定义服务器显示名称 | - |
| `port` | Server Query 端口 | 10011 |

## 🛠️ 开发规范

- 使用 `ThreadPoolExecutor` 实现多服务器并行采集。
- 代码适配了 `python-ts3` 库的底层 API 调用，确保兼容性。
- 指标统一以 `teamspeak_` 为前缀。

## 👤 作者

- **仙姑本咕**

## 🙏 鸣谢

本项目参考或衍生自以下优秀开源项目：
- [xkdvSrPD/teamspeak-prometheus](https://github.com/xkdvSrPD/teamspeak-prometheus)
- [TilmannF/teamspeak-prometheus](https://github.com/TilmannF/teamspeak-prometheus)

## 📄 开源协议

本项目采用 MIT 协议开源，详情请参阅 [LICENSE.md](LICENSE.md)。
