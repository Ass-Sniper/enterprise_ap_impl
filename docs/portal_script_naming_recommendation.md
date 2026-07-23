# Portal 脚本命名建议

## 背景

当前 Portal 数据面包含两个核心脚本：

```text
portal-agent.sh
portal-fw.sh
```

随着 Portal 功能逐步完善，脚本已不仅限于最初的职责，因此现有命名已经不能完全反映实际功能。

本文对两个脚本的职责进行分析，并给出更符合职责的命名建议。

---

# 当前命名分析

## portal-agent.sh

### 当前职责

负责：

- 从 Controller 获取 Runtime
- 解析 Runtime JSON
- 生成 `/tmp/portal-runtime.env`
- 调用数据面脚本同步配置
- Health Check

整个流程如下：

```text
Controller
      │
      ▼
Fetch Runtime
      │
      ▼
Parse Runtime
      │
      ▼
Generate Runtime Env
      │
      ▼
Apply Dataplane
```

实际上，该脚本更像是 **Runtime Bootstrapper** 或 **Runtime Synchronizer**。

"Agent" 虽然能够表达它是后台组件，但不能准确体现其核心职责。

---

## portal-fw.sh

### 当前职责

脚本已经不仅仅负责 Firewall。

实际完成：

- iptables
- ipset
- dnsmasq
- DNS Hijack
- HTTP Redirect
- Portal Bypass
- Runtime Reconcile
- Forward Policy

整体流程：

```text
Runtime Env
      │
      ▼
Configure iptables
      │
      ▼
Configure ipset
      │
      ▼
Configure dnsmasq
      │
      ▼
Portal Dataplane Ready
```

因此，"fw"（Firewall）已经不足以概括其职责。

---

# 推荐命名

## 推荐方案

### Runtime

```text
portal-runtime.sh
```

职责：

- Runtime Bootstrap
- Runtime Synchronization
- Runtime Environment Generation

---

### Dataplane

```text
portal-dataplane.sh
```

职责：

- iptables
- ipset
- dnsmasq
- Portal Redirect
- DNS Hijack
- Forward Control
- Bypass Policy

整体关系更加清晰：

```text
Controller
      │
      ▼
portal-runtime.sh
      │
      ▼
portal-runtime.env
      │
      ▼
portal-dataplane.sh
```

---

# 可选命名方案

## 方案一（推荐）

```text
portal-runtime.sh
portal-dataplane.sh
```

优点：

- 简洁
- 职责明确
- 与实现解耦
- 易于扩展

---

## 方案二

```text
portal-runtime-sync.sh
portal-dataplane.sh
```

突出 Runtime Synchronization。

适用于后续支持：

- 周期同步
- 增量同步
- Policy Version 管理

---

## 方案三

```text
portal-bootstrap.sh
portal-dataplane.sh
```

突出启动阶段的 Runtime 初始化。

但当脚本后续承担持续同步职责时，"bootstrap" 含义会逐渐弱化，因此不作为首选。

---

# 为什么不建议继续使用 portal-fw.sh

从实现来看，该脚本已经远超 Firewall 的范畴。

当前管理内容包括：

- iptables
- ipset
- dnsmasq
- DNS 劫持
- HTTP Portal Redirect
- Bypass 策略
- Runtime Reconcile
- Forward Policy

如果继续命名为：

```text
portal-fw.sh
```

容易让维护者误认为：

> 该脚本仅负责 Firewall 规则。

而实际上，它承担了整个 Portal 数据面的配置工作。

因此：

```text
portal-dataplane.sh
```

能够更准确表达其职责。

---

# 推荐目录结构

```text
data-plane/tools/

portal-runtime.sh
portal-dataplane.sh
portal-selftest.sh
```

职责对应如下：

## portal-runtime.sh

负责：

- 获取 Runtime
- Runtime 校验
- Runtime 落盘
- 调用 Dataplane

---

## portal-dataplane.sh

负责：

- iptables
- ipset
- dnsmasq
- DNS Hijack
- HTTP Redirect
- Portal Bypass
- Forward Policy

---

## portal-selftest.sh

负责：

- Dataplane Health Check
- ipset 检查
- dnsmasq 检查
- Runtime 检查

---

# 最终建议

推荐采用如下命名：

| 当前名称 | 建议名称 | 原因 |
|----------|----------|------|
| portal-agent.sh | **portal-runtime.sh** | 更准确体现 Runtime 获取、解析与同步职责 |
| portal-fw.sh | **portal-dataplane.sh** | 覆盖整个数据面配置，而不仅限于 Firewall |

采用该命名后，整体职责关系如下：

```text
Controller
      │
      ▼
portal-runtime.sh
      │
      ▼
portal-runtime.env
      │
      ▼
portal-dataplane.sh
      │
      ▼
iptables
ipset
dnsmasq
```

新的命名更加符合组件职责，也更利于后续扩展和长期维护。