
# Deployment Guide

## Overview

本文档描述 **Enterprise AP Captive Portal System** 的完整部署流程，
包括：

- 控制面（Controller）
- Captive Portal 服务
- AP 数据面（OpenWrt / ImmortalWrt）

目标是实现：

- 控制面容器化运行
- AP 侧零持久化、可回滚
- 快速部署、快速验证

---

## Deployment Topology

```text
+----------------------+        HTTP        +----------------------+
|  AP (OpenWrt)       |  <-------------->  |  Controller (Docker) |
|  Data Plane         |                    |  + Portal Server     |
+----------------------+                    +----------------------+
```

---

## Prerequisites

### Control Plane Host

* Linux x86_64
* Docker ≥ 20.x
* docker-compose ≥ v2
* 可被 AP 访问的 IP 地址

---

### AP (Data Plane)

* OpenWrt / ImmortalWrt
* dnsmasq（启用 ipset）
* ipset
* iptables
* curl
* jq / jsonfilter

---

## Control Plane Deployment

### 1️⃣ Clone Repository

```bash
git clone <repo-url>
cd enterprise_ap_impl/control-plane
```

---

### 2️⃣ Configuration

编辑控制面配置文件：

```text
control-plane/config/controller.yaml
```

常见配置项：

```yaml
server:
  listen: 0.0.0.0:8443

portal:
  base_url: http://<controller-ip>:8080
```

---

### 3️⃣ Start Controller & Portal

```bash
docker-compose up -d
```

启动后将包含：

* ap-controller
* captive-portal

---

### 4️⃣ Verify Controller

```bash
curl http://<controller-ip>:8443/api/v1/policy/runtime
```

期望返回 JSON Runtime。

---

## AP Data Plane Deployment

### 1️⃣ Copy Scripts to AP

在宿主机执行：

```bash
scp data-plane/tools/*.sh root@<AP-IP>:/mnt/
```

---

### 2️⃣ Prepare Scripts

在 AP 上执行：

```bash
ssh root@<AP-IP>
chmod +x /mnt/portal-*.sh
```

---

### 3️⃣ Initial Run

```bash
/mnt/portal-agent.sh
```

执行内容：

* 拉取 Runtime
* 写入 `/tmp/portal-runtime.env`
* 应用防火墙 / DNS / ipset 规则

---

### 4️⃣ Verify Health

```bash
/mnt/portal-agent.sh --check
```

成功输出示例：

```text
portal-agent: dataplane healthy
```

---

## Captive Portal Deployment

### Docker Image

Portal Server 通常由 `docker-compose` 自动启动，
如需单独构建：

```bash
cd control-plane/captive-portal/portal-server
docker build -t portal-server .
```

---

### Access Portal

浏览器访问：

```text
http://<controller-ip>:8080
```

---

## Rollback & Recovery

### Backup Dataplane State

```bash
/mnt/portal-backup.sh
```

---

### Restore Dataplane State

```bash
/mnt/portal-restore.sh
```

用途：

* 升级失败回滚
* 配置错误恢复
* 应急处理

---

## Restart & Reapply

数据面 **可安全重复执行**：

```bash
/mnt/portal-agent.sh
```

幂等保证：

* 不重复创建规则
* 不破坏现有授权状态
* 不依赖执行顺序

---

## Logging & Debugging

### AP Logs

```bash
logread | grep portal
```

---

### Controller Logs

```bash
docker-compose logs -f
```

---

## Common Deployment Pitfalls

### dnsmasq 未启用 ipset

检查：

```bash
dnsmasq --version | grep ipset
```

---

### ipset 类型错误

```bash
ipset list portal_bypass_dns
```

必须为：

```text
Type: hash:ip
```

---

### Portal 页面不弹出

* 检查 DNS bypass 是否生效
* 确认未劫持 HTTPS
* 查看 NAT 表规则

---

## Upgrade Strategy

推荐流程：

1. 更新 Controller
2. 验证 Runtime API
3. AP 执行 `portal-agent.sh`
4. 执行 `--check`
5. 如失败，立即 `portal-restore.sh`

---

## Production Recommendations

* Controller 建议部署在内网
* AP 与 Controller 间使用 HTTPS（可选）
* 定期执行 dataplane healthcheck
* 日志集中收集

---

## Summary

通过本部署方案，可以实现：

* 快速搭建 Captive Portal 系统
* 清晰的职责分离
* 可回滚、可自检的数据面
* 适合生产与实验环境

该系统既可作为 **生产实现**，也可作为 **企业 AP 架构参考**。
