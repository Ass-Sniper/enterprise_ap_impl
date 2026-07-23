# Portal Agent Runtime Bootstrap 流程分析

## 概述

`portal-agent.sh` 是 Portal 数据面的运行时引导（Runtime Bootstrap）组件。

运行位置：

- ImmortalWRT（Data Plane）

主要职责：

1. 从 Controller 获取最新 Portal Runtime 配置
2. 解析 Runtime JSON
3. 生成本地运行环境文件 `/tmp/portal-runtime.env`
4. （可选）调用 `portal-fw.sh` 同步 iptables/ipset 数据面规则

整个脚本不直接操作防火墙，而是负责**Runtime 同步**，因此属于 **Control Plane → Data Plane** 的桥梁。

---

# 整体流程

```text
          Controller
               │
               │ HTTP GET Runtime
               ▼
        portal-agent.sh
               │
               │
      Parse Runtime JSON
               │
               ▼
 Generate /tmp/portal-runtime.env
               │
               ▼
     Execute portal-fw.sh
               │
               ▼
   Configure iptables/ipset/dnsmasq
```

---

# 功能模块

整个脚本可以划分为六个阶段。

```
初始化
    │
    ▼
Health Check（可选）
    │
    ▼
获取 Runtime
    │
    ▼
解析 Runtime
    │
    ▼
生成 Runtime Env
    │
    ▼
调用 portal-fw.sh
```

---

# 第一阶段：初始化

## 默认配置

初始化默认参数：

- Controller 地址
- Runtime API
- AP 身份
- Runtime 文件位置
- 是否自动 Apply Firewall

例如：

```text
CTRL_HOST
CTRL_PORT
CTRL_BASE

AP_ID
SITE_ID
RADIO_ID

RUNTIME_ENV
PORTAL_FW
```

这些参数均支持环境变量覆盖。

例如：

```
CTRL_HOST=10.0.0.10 portal-agent.sh
```

即可替换默认 Controller。

---

# 第二阶段：Health Check

支持：

```
portal-agent.sh --check
```

执行流程：

```
portal-agent
      │
      ▼
portal-dnsmasq-ipset-selftest.sh
      │
      ▼
Pass / Fail
```

主要用于：

- dnsmasq
- ipset
- dataplane

是否正常工作。

返回值：

```
0   healthy

1   unhealthy

2   selftest不存在
```

---

# 第三阶段：获取 Runtime

首先请求：

```
GET

/api/v1/policy/runtime
```

如果失败：

自动回退：

```
/policy/runtime
```

保证兼容旧 Controller。

请求参数：

```
site

ap_id

radio_id
```

最终得到：

```
Runtime JSON
```

例如：

```json
{
  "dataplane": {
    "portal_ip": "...",
    "lan_if": "...",
    "policy_version": 12
  },
  "bypass": {
    ...
  }
}
```

---

# 第四阶段：解析 Runtime

主要解析四类数据。

---

## 1）Policy Version

读取：

```
dataplane.policy_version
```

特点：

- 必须是数字
- 非数字自动回退 0

例如：

```
policy_version=15
```

或者：

```
policy_version=0
```

表示未知版本。

---

## 2）Dataplane 参数

读取：

```
LAN_IF

PORTAL_IP

DNS_PORT
```

例如：

```
br-lan

192.168.16.118

53
```

---

## 3）IPSet 名称

读取：

```
allow.guest

allow.staff
```

默认：

```
portal_allow_guest

portal_allow_staff
```

Controller 可以统一下发名称。

---

## 4）Bypass 配置

解析：

```
MAC 白名单

IP 白名单

Domain 白名单

Enabled
```

即：

```
BYPASS_MACS

BYPASS_IPS

BYPASS_DOMAINS

BYPASS_ENABLED
```

这些内容保持 JSON 字符串格式。

后续由：

```
portal-fw.sh
```

负责进一步解析。

---

# 第五阶段：生成 Runtime Environment

最终生成：

```
/tmp/portal-runtime.env
```

采用：

```
write_env_atomic()
```

流程：

```
stdin
   │
   ▼
tmp file
   │
   ▼
mv(rename)
   │
   ▼
portal-runtime.env
```

具有以下特点：

- 原子更新
- 不会产生半写文件
- 多进程安全
- 使用 rename()

最终生成：

```
export LAN_IF=...

export PORTAL_IP=...

export POLICY_VERSION=...

export BYPASS_MACS=...

...
```

随后：

```
portal-fw.sh
```

直接 source 即可。

---

# 第六阶段：Apply Dataplane

如果：

```
APPLY_FW=1
```

则自动执行：

```
portal-fw.sh
```

流程：

```
Runtime更新
      │
      ▼
portal-runtime.env
      │
      ▼
portal-fw.sh
      │
      ▼
iptables

ipset

dnsmasq
```

因此：

portal-agent 不关心规则如何安装。

只负责：

> Runtime 同步 + Apply。

---

# Policy Version 逻辑

读取：

```
policy_version
```

流程：

```
读取

↓

是否合法

↓

合法？

↓

Yes
↓

使用版本

No
↓

0
```

随后：

```
0

↓

FORCE_APPLY

↓

立即刷新 Firewall
```

这样即使 Controller 没实现版本控制，

仍然保证：

```
Runtime 必定同步。
```

---

# Runtime 文件作用

portal-agent 与 portal-fw 之间没有共享内存。

二者通过：

```
/tmp/portal-runtime.env
```

通信。

关系如下：

```
Controller
     │
     ▼
portal-agent
     │
     ▼
portal-runtime.env
     │
     ▼
portal-fw
```

因此：

```
portal-runtime.env
```

就是：

整个 Data Plane 的 Runtime Configuration。

---

# 模块职责划分

## portal-agent

负责：

- 获取 Runtime
- Runtime 校验
- Runtime 落盘
- 调用 portal-fw

不负责：

- iptables
- ipset
- dnsmasq

---

## portal-fw

负责：

- 创建 ipset
- 创建 iptables
- DNS 劫持
- Portal Redirect
- Bypass
- Forward Control

不负责：

- HTTP 获取 Runtime

---

# 设计特点

整个架构遵循以下原则：

1. **控制面与数据面解耦**

Controller 不直接操作 iptables，仅下发 Runtime。

---

2. **Runtime 文件作为唯一数据源**

所有数据面规则均来源于：

```
/tmp/portal-runtime.env
```

避免多个组件维护不同状态。

---

3. **原子更新**

使用：

```
tmp → rename()
```

确保 Runtime 更新过程中不会出现部分写入。

---

4. **向后兼容**

Runtime API 支持：

- 新接口
- Legacy 接口

降低升级风险。

---

5. **职责单一**

portal-agent 仅负责 Runtime 生命周期管理。

portal-fw 专注于数据面规则构建，职责清晰，便于独立维护和扩展。