
## 一、整体拓扑（高度抽象版）

```
                    ┌────────────────────────────┐
                    │        Internet / ISP       │
                    └──────────────┬─────────────┘
                                   │
                            (DHCP / NAT)
                                   │
                            ImmortalWRT
                            WAN: eth1
                         192.168.124.0/24
                                   │
                ┌──────────────────┴──────────────────┐
                │           ImmortalWRT                │
                │                                      │
                │   br-lan (VLAN filtering ON)         │
                │                                      │
                │   VLAN 1  (管理网)  192.168.16.0/24  │
                │   VLAN 10 (业务网)  172.16.10.0/24   │
                │                                      │
                └───────┬───────────────┬─────────────┘
                        │               │
                  lan1 (untagged)   lan3 (untagged)
                  VLAN 1            VLAN 10
                        │               │
                        │               └─────────(备用/测试)
                        │
          Intel I219-V 物理网卡（Win10）
                        │
┌───────────────────────┴──────────────────────────────┐
│                       Windows 10                      │
│                                                        │
│  VMware 虚拟交换体系（关键！）                         │
│                                                        │
│  VMnet1  Host-Only   192.168.100.0/24                 │
│  VMnet8  NAT         192.168.200.0/24                 │
│                                                        │
│  同一台物理网卡被 VMware Bridge/NAT/HostOnly 复用     │
└───────────────┬───────────────────────────┬──────────┘
                │                           │
        Ubuntu 20.04 VM              Ubuntu 16.04 VM
        (Portal / RADIUS)            (备用/测试)

```

---

## 二、逐层拆解（这是最关键的）

### 1️⃣ **ImmortalWRT：纯“交换 + 三层网关”角色**

你在 ImmortalWRT 上做的事情非常标准，而且是**正确的**：

#### VLAN & Bridge 关系

| 组件          | 角色                                     |
| ----------- | -------------------------------------- |
| `br-lan`    | VLAN-aware bridge（交换层）                 |
| `br-lan.1`  | VLAN 1 的三层接口（管理网）                      |
| `br-lan.10` | VLAN 10 的三层接口（业务网）                     |
| `lan1`      | access port → VLAN 1                   |
| `lan3`      | access port → VLAN 10                  |
| `lan4`      | hybrid（VLAN1 untagged + VLAN10 tagged） |
| `rax0`      | Wi-Fi 接口 → VLAN 10                     |

👉 **结论**：
ImmortalWRT 在这里**不是 NAT，不是 Portal**，它只是：

* VLAN 切分
* 三层网关
* 把 VLAN 1 / VLAN 10 流量送上游

---

### 2️⃣ **Win10：这是“最复杂、最容易出问题的一层”**

Windows 在这里 **同时扮演了三种角色**：

1. **真实物理二层设备**
2. **VMware 虚拟交换机宿主**
3. **NAT / Host-only / Bridge 的调度者**

#### 实际效果等价于：

```
ImmortalWRT VLAN 1
   ↓
Win10 物理 NIC
   ↓
VMware 虚拟交换层
   ↓
┌───────────────┬─────────────────┐
│ VMnet1        │ VMnet8          │
│ Host-Only     │ NAT             │
│ 192.168.100.0 │ 192.168.200.0   │
└───────────────┴─────────────────┘
```

⚠️ **关键认知**：

* VMnet1 / VMnet8 **不是 VLAN**
* 它们是 **VMware 私有的 L2 网络**
* 和 ImmortalWRT 的 VLAN **完全不同层次**

---

### 3️⃣ **Ubuntu 20.04 VM：你真正的“控制平面节点”**

你这台 VM 非常典型，是**“多腿控制节点”**：

#### 接口职责表

| 接口             | 网络               | 作用                     |
| -------------- | ---------------- | ---------------------- |
| ens38          | 192.168.16.0/24  | **直连 ImmortalWRT 管理网** |
| ens33          | 192.168.100.0/24 | Host-only（运维/调试）       |
| ens37          | 192.168.200.0/24 | NAT（出公网）               |
| docker0 / br-* | 172.17/18/19     | **容器内部网络**             |

#### Docker 网络再抽象一层

```
Ubuntu ens38 (192.168.16.118)
        │
   iptables DNAT
        │
Docker bridge (172.19.0.0/16)
        │
┌───────────┬───────────┬───────────┐
│ FreeRADIUS│ MySQL     │ Redis     │
│ 172.19.0.5│ 172.19.0.2│ 172.19.0.3│
└───────────┴───────────┴───────────┘
```

👉 **非常重要的事实**：

* ImmortalWRT **永远看不到** 172.19.0.0/16
* 它只看到 `192.168.16.118`
* Docker 只是 **Ubuntu 内部私网**

---

## 三、真实数据流（以 Portal / RADIUS 为例）

### 🔁 RADIUS 认证流

```
Client (VLAN10)
   ↓
ImmortalWRT (br-lan.10)
   ↓
Win10 NIC
   ↓
Ubuntu ens38 (192.168.16.118)
   ↓
iptables DNAT
   ↓
FreeRADIUS (172.19.0.5:1812)
```

### 🔁 Portal HTTP 流

```
Client (VLAN10)
   ↓
ImmortalWRT (redirect / policy)
   ↓
Ubuntu ens38
   ↓
Docker captive-portal-server
```

---

## 四、你这套拓扑的“本质一句话总结”

> **这是一个典型的「企业 AP 控制面实验环境」：
> ImmortalWRT 只负责交换和三层，
> Win10 是虚拟交换中枢，
> Ubuntu 是真正的控制平面与业务承载节点。**

---

