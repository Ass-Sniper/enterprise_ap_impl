
# Data Plane

## Overview

数据面（Data Plane）运行在 AP（OpenWrt / ImmortalWrt）设备上，
负责 **实时流量控制、DNS 处理、认证前后访问决策**。

数据面是一个 **纯执行层**：
- 不保存长期策略
- 不进行策略决策
- 只根据 Control Plane 下发的 Runtime 执行

The data plane executes runtime policies on the AP and is responsible for
DNS handling, firewall enforcement, and authentication gating.

---

## Responsibilities

数据面的核心职责包括：

- DNS 重定向与旁路（Captive Portal Detection）
- 防火墙转发控制（iptables）
- 授权 / 旁路集合维护（ipset）
- 运行时幂等应用
- 本地自检与健康检查

---

## Directory Structure

```text
data-plane/tools/
├── portal-agent.sh                    # Runtime 拉取与调度入口
├── portal-fw.sh                       # 防火墙 / DNS / ipset 主逻辑
├── portal-sync.sh                     # 授权状态同步
├── portal-backup.sh                   # 规则备份
├── portal-restore.sh                  # 回滚恢复
└── portal-dnsmasq-ipset-selftest.sh   # DNS + ipset 自检
```

---

## Execution Model

### Entry Point: `portal-agent.sh`

`portal-agent.sh` 是数据面的 **唯一入口**：

* 从 Controller 拉取 Runtime
* 生成 `/tmp/portal-runtime.env`
* 调用 `portal-fw.sh` 应用规则
* 提供 `--check` 健康检查接口

```bash
portal-agent.sh
portal-agent.sh --check
```

---

## Runtime Environment

Runtime 以 **环境变量文件** 形式存在：

```text
/tmp/portal-runtime.env
```

示例：

```bash
export LAN_IF='br-lan'
export PORTAL_IP='192.168.16.118'
export IPSET_GUEST='portal_allow_guest'
export IPSET_STAFF='portal_allow_staff'
export BYPASS_DOMAINS='["msftconnecttest.com","captive.apple.com"]'
```

特点：

* 原子写入
* 每次执行覆盖
* 不持久化

---

## DNS Handling (dnsmasq + ipset)

### Design Goal

* 不劫持 HTTPS
* 不伪造 DNS 响应
* 仅在 **DNS 解析时** 标记目的 IP

---

### Implementation

1. dnsmasq 配置：

```text
/tmp/dnsmasq.d/portal-bypass-ipset.conf
```

内容示例：

```ini
ipset=/msftconnecttest.com/portal_bypass_dns
ipset=/captive.apple.com/portal_bypass_dns
```

2. dnsmasq 解析域名时：

   * 自动将 A / AAAA 记录写入 ipset
   * 数据面 **不执行 `ipset add`**

---

### IPSet Type

```text
portal_bypass_dns: hash:ip
```

> ⚠️ 域名 ipset **必须是 hash:ip**

---

## Firewall (iptables)

### NAT Table

* DNS 流量引导（可选）
* Portal HTTP 重定向

### FILTER Table

* 未认证流量拦截
* 已认证 / 旁路流量放行

---

### Flow Decision Order

1. BYPASS MAC
2. BYPASS IP
3. BYPASS DNS (ipset)
4. AUTHENTICATED
5. REDIRECT TO PORTAL

---

## IPSet Sets

| IPSet Name         | Type     | Purpose |
| ------------------ | -------- | ------- |
| portal_allow_guest | hash:mac | 已授权访客   |
| portal_allow_staff | hash:mac | 已授权员工   |
| portal_bypass_mac  | hash:mac | 设备旁路    |
| portal_bypass_ip   | hash:ip  | IP 旁路   |
| portal_bypass_dns  | hash:ip  | 域名旁路    |

---

## Backup & Restore

### Backup

```bash
portal-backup.sh
```

* 备份 iptables
* 备份 ipset

---

### Restore

```bash
portal-restore.sh
```

* 销毁现有规则
* 恢复备份状态

---

## Health Check & Self-Test

### portal-agent --check

触发以下检查：

* dnsmasq 进程
* dnsmasq ipset 支持
* dnsmasq conf-dir
* ipset 是否存在
* ipset 类型是否正确
* DNS 查询是否触发写入

---

### Self-Test Script

```bash
portal-dnsmasq-ipset-selftest.sh
```

支持：

```bash
portal-dnsmasq-ipset-selftest.sh \
  portal_bypass_dns \
  www.microsoft.com
```

---

## Idempotency & Safety

数据面所有脚本遵循：

* 重复执行安全
* 不依赖执行顺序
* 不假设初始状态
* 失败可回滚

---

## Failure Handling

* dnsmasq reload 失败 → restart
* 规则应用失败 → 不破坏已有流量
* 自检失败 → 标记 dataplane unhealthy

---

## Design Principles

* **AP 是无状态的**
* **DNS 是唯一“隐式信号”**
* **iptables 永远是最后裁决者**
* **Portal 永不直接改防火墙**

---

## Summary

数据面是整个系统中 **最敏感、最靠近流量的层**，
其设计目标是：

* 稳定
* 幂等
* 可回滚
* 可自检

通过 dnsmasq + ipset 的组合，
实现了对现代 OS Captive Portal 行为的 **正确、非侵入式支持**。


---
