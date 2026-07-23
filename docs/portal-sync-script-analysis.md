# ImmortalWRT Portal Sync（MAC → ipset）脚本功能分析

## 1. 概述

该脚本是 **ImmortalWRT Portal 的数据平面同步器（Data Plane Sync Agent）**，负责同步 AP 上实际在线的终端状态与 Controller 的授权状态，并维护 Linux `ipset`，供 `iptables`/`nftables` 进行访问控制。

整体架构如下：

```text
        +-----------------------------+
        |      AP Controller          |
        |  Session / Portal Service   |
        +-------------+---------------+
                      ^
            batch_status / heartbeat
                      |
                HTTP REST API
                      |
+------------------------------------------------------+
|           ImmortalWRT AP (Data Plane)                |
|                                                      |
| portal-sync.sh                                       |
|                                                      |
| DHCP  ----+                                          |
| ARP   ----+--> 收集 MAC ---> Controller 查询 ---> ipset
| FDB   ----+                              |
|                                          |
|                            portal_allow_guest
|                            portal_allow_staff
|                            portal_bypass_mac
|                                          |
|                                  firewall/nftables
+------------------------------------------------------+
```

---

# 2. 初始化

脚本启动后首先读取运行环境：

```sh
RUNTIME_ENV=/tmp/portal-runtime.env
```

通常包含：

```text
CTRL_BASE
AP_ID
POLICY_VERSION
LAN_IF
```

例如：

```text
CTRL_BASE=http://192.168.16.118:8443
AP_ID=AP-001
POLICY_VERSION=8
```

随后确保以下 ipset 存在：

```text
portal_allow_guest
portal_allow_staff
portal_bypass_mac
```

如果不存在则自动创建。

---

# 3. AP 自身旁路（Bypass）

脚本读取 AP 自身 LAN MAC：

```sh
SELF_MAC="$(cat /sys/class/net/${LAN_IF}/address)"
```

然后加入：

```text
portal_bypass_mac
```

作用：

避免 Portal 规则误将 AP 自己阻断，导致：

```text
AP
 ↓
iptables
 ↓
Portal
 ↓
Controller 无法访问
```

因此 AP 自身必须永久旁路。

---

# 4. 在线终端发现

脚本同时从三个来源收集客户端 MAC。

## 4.1 DHCP Lease

读取：

```text
/tmp/dhcp.leases
```

示例：

```text
1723456789 aa:bb:cc:11:22:33 192.168.16.101 phone *
```

得到：

```text
aa:bb:cc:11:22:33
source = dhcp
```

---

## 4.2 ARP 邻居表

执行：

```sh
ip neigh show dev br-lan
```

示例：

```text
192.168.16.101 dev br-lan lladdr aa:bb:cc:11:22:33 REACHABLE
```

得到：

```text
aa:bb:cc:11:22:33
source = arp
```

过滤：

```text
FAILED
```

状态，因为表示设备已不可达。

---

## 4.3 Bridge FDB

执行：

```sh
brctl showmacs br-lan
```

例如：

```text
port no mac addr is local ageing timer

1 aa:bb:cc:11:22:33 no
```

得到：

```text
aa:bb:cc:11:22:33
source = fdb
```

忽略：

```text
is_local = yes
```

因为属于桥自身 MAC。

---

# 5. 为什么同时使用 DHCP / ARP / FDB？

三者互补，各有优缺点。

| 来源   | 优点      | 缺点           |
| ---- | ------- | ------------ |
| DHCP | 最可靠     | 静态 IP 无法发现   |
| ARP  | 能发现通信设备 | ARP 会过期      |
| FDB  | 二层学习最快  | 只能说明 MAC 出现过 |

例如：

静态 IP：

```text
DHCP ×
ARP √
FDB √
```

刚接入 WiFi：

```text
DHCP ×
ARP ×
FDB √
```

因此三者共同收集，提高在线检测准确率。

---

# 6. 来源优先级

脚本定义：

```text
DHCP > ARP > FDB
```

原因：

DHCP 最可信。

例如同一 MAC：

```text
FDB
ARP
DHCP
```

最终保留：

```text
MAC -> dhcp
```

避免同一终端重复统计。

---

# 7. Miss Threshold（离线判定）

这是脚本的重要设计之一。

维护状态文件：

```text
/tmp/portal-sync.state
```

格式：

```text
MAC miss_count
```

例如：

```text
AA 0
BB 1
CC 2
```

每轮同步：

发现客户端：

```text
miss = 0
```

未发现：

```text
miss++
```

达到：

```text
MISS_THRESHOLD = 3
```

才真正认为离线。

例如：

```text
第一次扫描
AA 不存在
↓

AA miss=1

第二次扫描
↓

AA miss=2

第三次扫描
↓

AA miss=3

↓

真正删除授权
```

目的：

* 避免 WiFi 漫游
* 避免 ARP 短暂失效
* 避免 DHCP 刷新
* 避免 Bridge FDB 更新造成误踢用户

---

# 8. Active / Offline 分类

根据 miss_count：

```text
miss < threshold
```

生成：

```text
active.txt
```

表示：

仍然在线。

达到阈值：

```text
offline.txt
```

立即执行：

```text
ipset del
```

删除所有授权。

---

# 9. 批量查询 Controller

对于 Active MAC，构造：

```json
{
  "policy_version":"8",
  "ap_id":"AP01",
  "entries":[
    {
      "mac":"aa:bb:cc:11:22:33",
      "source":"dhcp"
    }
  ]
}
```

发送：

```text
POST /api/v1/session/batch_status
```

失败后自动回退：

```text
POST /portal/batch_status
```

Controller 返回：

```json
{
  "results":[
    {
      "mac":"aa:bb:cc:11:22:33",
      "authorized":true,
      "ttl":280,
      "role":"guest",
      "ipset":"portal_allow_guest"
    }
  ]
}
```

---

# 10. TTL 策略

Controller 返回：

```text
ttl = 600
```

脚本不会完全采用，而是根据来源限制：

| 来源   | 最大 TTL            |
| ---- | ----------------- |
| DHCP | 使用 Controller TTL |
| ARP  | 300 秒             |
| FDB  | 120 秒             |

最终：

```text
实际 TTL = min(Controller TTL, Source Cap)
```

例如：

```text
Controller = 600

ARP Cap = 300

最终 TTL = 300
```

这样即使 ARP 很快消失，也不会长期保留授权。

---

# 11. 更新 ipset

根据 Controller 返回：

```text
role
```

决定目标 ipset：

```text
guest
↓

portal_allow_guest
```

```text
staff
↓

portal_allow_staff
```

执行：

```sh
ipset add timeout TTL
```

若角色发生变化：

```text
guest
↓

staff
```

则自动：

```text
portal_allow_guest 删除

portal_allow_staff 添加
```

保证客户端只存在于一个授权集合中。

---

# 12. Heartbeat

同步完成后，发送：

```text
POST /api/v1/session/batch_heartbeat
```

目的不是重新授权，而是告诉 Controller：

```text
这些客户端仍然在线
```

Controller 可用于：

* 延长 Session
* 更新最后在线时间
* 维持 Portal 登录状态

如果 Batch Heartbeat 不存在，则直接跳过，不进行逐 MAC 请求，避免大量 HTTP 请求。

---

# 13. 日志

同步结束输出：

```text
event=sync_done

policy_version=8

seen=20
active=18
offline=2

scanned=18
allowed=17
refreshed=17
removed=2
heartbeated=18
```

可通过：

```sh
logread -e portal-sync
```

查看同步情况。

---

# 14. 整体流程

```text
          DHCP
            │
          ARP
            │
          FDB
            │
            ▼
      收集所有 MAC
            │
            ▼
   去重（DHCP > ARP > FDB）
            │
            ▼
      更新 miss_count
            │
            ├───────────────► offline
            │                    │
            │                    ▼
            │               删除 ipset
            │
            ▼
         active MAC
            │
            ▼
POST /session/batch_status
            │
            ▼
     Controller 返回授权状态
            │
            ▼
   根据 role + ttl 更新 ipset
            │
            ▼
POST /session/batch_heartbeat
            │
            ▼
         输出统计日志
```

---

# 15. 设计特点总结

## 优点

### 多来源在线检测

结合 DHCP、ARP、FDB 三种来源，能够覆盖动态 IP、静态 IP、刚接入 WiFi 等多种场景。

### 数据平面与控制平面解耦

AP 不保存授权逻辑，仅负责采集在线终端和执行策略；Controller 负责授权决策、Session 管理和策略下发。

### 批量接口设计

采用 Batch Status 和 Batch Heartbeat，避免大量单客户端 HTTP 请求，提高同步效率。

### TTL 自动控制

结合 Controller TTL 与本地来源 TTL 上限，避免缓存过久导致授权状态失真。

### Miss Threshold

采用连续多次未发现才判定离线的策略，有效降低 WiFi 漫游、ARP 更新等瞬时变化带来的误踢风险。

### 自动角色切换

当用户角色发生变化时，自动迁移至对应 ipset，确保授权状态始终一致。

---

# 16. 总结

该脚本实现了一个典型的 **Controller + AP（Control Plane + Data Plane）Portal 架构**。

AP 专注于在线终端发现和本地访问控制，Controller 专注于授权和 Session 管理，两者通过批量 REST API 协同工作。脚本利用 **DHCP + ARP + FDB 多源检测**、**批量同步**、**TTL 控制**、**Miss Threshold 离线判定** 等机制，实现了较高的可靠性、可扩展性和网络抖动容忍能力，适合作为企业级 Portal 或集中式无线控制系统的数据平面同步组件。
