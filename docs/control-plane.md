
# Control Plane

## Overview

控制面（Control Plane）负责 **集中管理 AP 的运行时策略（Runtime Policy）**，
并通过 HTTP API 下发给各个 AP。  
控制面 **不直接操作数据面（iptables / dnsmasq / ipset）**，而是通过
“声明式运行时配置”驱动 AP 行为。

The control plane centrally manages **runtime policies** for APs and exposes
them via HTTP APIs.  
It never directly manipulates dataplane components; instead, it delivers
declarative runtime state consumed by AP agents.

---

## Responsibilities

控制面的核心职责包括：

- 提供 AP Runtime API
- 管理策略版本（Policy Version）
- 统一下发 Portal / Bypass / Dataplane 配置
- 审计与可追溯（Audit Log）
- 支持多 AP / 多 Site

---

## Directory Structure

```text
control-plane/
├── ap-controller/        # Python reference implementation
├── ap-controller-go/     # Go production-oriented implementation
├── captive-portal/       # Portal UI & Auth server
├── config/
│   └── controller.yaml   # Controller configuration
└── docker-compose.yml    # Control plane bootstrap
```

## Controller Implementations

### 1️⃣ Python Controller (Reference)

**Path**

```text
control-plane/ap-controller/
```

**Purpose**

* 快速验证控制面模型
* 适合作为 PoC / Demo / 行为参考

**Key Files**

| File        | Description     |
| ----------- | --------------- |
| `main.py`   | HTTP API 入口     |
| `config.py` | Controller 配置加载 |
| `store.py`  | 策略与状态存储         |
| `audit.py`  | 审计日志            |

**Characteristics**

* 简单易读
* 单进程模型
* 不强调高并发

---

### 2️⃣ Go Controller (Production-Oriented)

**Path**

```text
control-plane/ap-controller-go/
```

**Purpose**

* 面向生产部署
* 强类型模型
* 更好的并发与扩展性

**Structure**

```text
internal/
├── audit/        # 审计模型与记录
├── config/       # 配置加载
├── http/         # API handlers
├── policy/       # Runtime policy & versioning
├── roles/        # 角色 / 决策模型
└── store/        # Runtime store abstraction
```

**Characteristics**

* 清晰的 domain model
* 易于扩展 RBAC / 多租户
* 适合长期维护

---

## Runtime Policy API

### Endpoint

```http
GET /api/v1/policy/runtime
```

### Query Parameters

| Name    | Description |
| ------- | ----------- |
| `ap_id` | AP 唯一标识     |
| `site`  | 站点 ID       |
| `radio` | 无线接口        |

---

### Response (Example)

```json
{
  "policy_version": 0,
  "ap_id": "ImmortalWrt",
  "site": "default",
  "radio": "radio0",

  "dataplane": {
    "lan_if": "br-lan",
    "portal_ip": "192.168.16.118",
    "dns_port": 53
  },

  "ipsets": {
    "guest": "portal_allow_guest",
    "staff": "portal_allow_staff"
  },

  "bypass": {
    "enabled": true,
    "macs": ["70:4d:7b:64:3b:da"],
    "ips": ["192.168.16.1", "192.168.16.118"],
    "domains": [
      "msftconnecttest.com",
      "captive.apple.com",
      "connectivitycheck.hicloud.com"
    ]
  }
}
```

---

## Policy Versioning

* 每次策略变更都会增加 `policy_version`
* AP 会在运行时日志中记录当前版本
* 支持未来扩展：

  * 条件下发
  * 灰度发布
  * 回滚策略

---

## Audit & Observability

控制面支持审计能力：

* Runtime 下发记录
* AP 拉取时间戳
* 策略版本变化

Go 实现中：

```text
internal/audit/
```

可扩展对接：

* 文件日志
* 数据库
* OpenTelemetry / Prometheus

---

## Captive Portal Integration

控制面同时托管 Portal Server：

```text
control-plane/captive-portal/
```

职责：

* 提供 Portal UI
* 处理用户认证 / 授权
* 与数据面通过 **IPSet / MAC / Session** 间接联动

> 控制面不直接修改 iptables，避免强耦合。

---

## Deployment Model

### Docker Compose

```bash
cd control-plane
docker-compose up -d
```

启动组件包括：

* Controller API
* Portal Server
* （可选）后端存储

---

## Design Principles

* **Stateless AP**：AP 不保存长期策略
* **Declarative Runtime**：只下发“应该是什么”
* **Loose Coupling**：控制面与数据面解耦
* **Audit First**：每次变更可追溯

---

## Summary

控制面是整个系统的 **策略中枢与单一事实源（Single Source of Truth）**。
通过清晰的 Runtime API 和版本化模型，实现：

* 多 AP 统一管理
* 数据面极简、稳定
* 系统可扩展、可演进

