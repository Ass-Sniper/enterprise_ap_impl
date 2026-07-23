# Portal 数据面初始化流程（portal-fw.sh）

## 概述

`portal-fw.sh` 是 Portal 数据面的初始化脚本，主要负责完成以下工作：

1. 加载运行时配置。
2. 创建并校验 ipset。
3. 同步（Reconcile）Bypass 白名单。
4. 配置 DNS 豁免。
5. 配置 HTTP Portal 重定向。
6. 配置 DNS 劫持。
7. 配置数据转发控制。
8. 完成 Portal 数据面初始化。

整个脚本采用 **Runtime Authoritative（运行时状态为唯一真相）** 的设计，每次执行都会重新同步运行时配置，而不是在已有状态上增量修改。

---

# 初始化流程

```text
启动 portal-fw.sh
        │
        ▼
加载 runtime.env
        │
        ▼
创建/校验 ipset
        │
        ▼
同步 Bypass 白名单
(MAC/IP/Domain)
        │
        ▼
重新生成 dnsmasq ipset 配置
        │
        ▼
Reload dnsmasq
        │
        ▼
配置 HTTP Portal
        │
        ▼
配置 DNS Portal
        │
        ▼
配置 Forward 访问控制
        │
        ▼
Portal 初始化完成
```

---

# 1. 加载运行时配置

脚本首先加载：

```text
/tmp/portal-runtime.env
```

该文件由：

```text
portal-agent.sh
```

动态生成。

主要包括：

- Portal IP
- Portal Port
- LAN 接口
- Bypass 开关
- Bypass MAC
- Bypass IP
- Bypass Domain

如果某项不存在，则使用脚本默认值。

---

# 2. 创建 ipset

初始化阶段保证所有 ipset 已存在。

包括：

```text
portal_allow_guest
portal_allow_staff
portal_bypass_mac
portal_bypass_ip
portal_bypass_dns
```

随后校验类型：

```text
hash:mac
hash:ip
```

如果类型错误立即退出，避免后续规则异常。

---

# 3. Reconcile Bypass 白名单

采用 **Flush + Rebuild** 模式。

流程：

```text
runtime
    │
    ▼
Flush ipset
    │
    ▼
解析 JSON
    │
    ▼
重新写入 ipset
```

同步对象包括：

## 3.1 MAC

同步：

```text
portal_bypass_mac
```

来源：

```text
BYPASS_MACS
```

同时自动加入：

```text
LAN 接口自身 MAC
```

避免设备自己被 Portal 拦截。

---

## 3.2 IP

同步：

```text
portal_bypass_ip
```

来源：

```text
BYPASS_IPS
```

---

## 3.3 Domain

Domain 使用两层模型。

### 第一层

dnsmasq 配置：

```text
portal-bypass-ipset.conf
```

内容例如：

```text
ipset=/example.com/portal_bypass_dns
```

dnsmasq 在解析域名时自动把解析出的 IP 放入：

```text
portal_bypass_dns
```

---

### 第二层

真正参与 iptables 匹配的是：

```text
portal_bypass_dns
```

类型：

```text
hash:ip
```

因此：

```text
Domain
    │
dnsmasq
    │
解析
    │
IP
    │
ipset(hash:ip)
```

---

# 4. Reload dnsmasq

重新加载：

```text
/etc/init.d/dnsmasq reload
```

若失败：

```text
restart
```

保证：

```text
portal-bypass-ipset.conf
```

立即生效。

---

# 5. HTTP Portal

创建：

```text
PORTAL_HTTP
```

然后挂载：

```text
PREROUTING
        │
        ▼
PORTAL_HTTP
```

位置：

```text
PREROUTING Rule #2
```

因为：

```text
Rule #1
```

预留给 DNS。

---

## HTTP 链规则

顺序如下：

```text
Guest
        │
RETURN

Staff
        │
RETURN

Bypass MAC
        │
RETURN

Bypass IP
        │
RETURN

其它 HTTP
        │
DNAT
        │
Portal
```

即：

未认证客户端访问：

```text
TCP/80
```

全部重定向：

```text
Portal_IP:Portal_Port
```

---

# 6. DNS Portal

创建：

```text
PORTAL_DNS
```

挂载：

```text
PREROUTING Rule #1
```

因此：

```text
PREROUTING

Rule1
PORTAL_DNS

Rule2
PORTAL_HTTP
```

DNS 永远优先于 HTTP。

---

## DNS 链规则

执行顺序：

```text
Guest
        │
RETURN

Staff
        │
RETURN

Bypass MAC
        │
RETURN

Bypass IP
        │
RETURN

Bypass Domain
        │
RETURN

其它 DNS
        │
DNAT
        │
Portal DNS
```

最终：

所有未授权 DNS：

```text
53
```

都会被：

```text
DNAT
```

到：

```text
Portal_IP:53
```

---

# 7. Forward 控制

创建：

```text
PORTAL_FWD
```

挂载：

```text
forwarding_lan_rule
```

流程：

```text
FORWARD
      │
      ▼
PORTAL_FWD
```

---

## 放行规则

优先放行：

```text
Guest
Staff
Bypass MAC
Bypass IP
Portal Server
```

注意：

```text
Bypass Domain
```

**不会放行 Forward。**

原因：

如果直接放行：

```text
portal_bypass_dns
```

Windows NCSI、

Android、

iOS Captive Portal Detection

都会直接访问互联网。

结果：

```text
200 OK
```

而不是：

```text
302 Redirect
```

Portal 登录页不会弹出。

因此：

Forward 中故意禁用：

```text
portal_bypass_dns
```

仅允许：

DNS 查询本身豁免。

---

# 8. HTTPS 处理

对于：

```text
TCP/443
```

统一：

```text
REJECT
tcp-reset
```

目的：

让客户端快速放弃 HTTPS。

随后自动访问：

```text
HTTP
```

触发：

```text
Portal Redirect
```

提高 Portal 自动弹窗成功率。

---

# 9. 默认拒绝

最后：

```text
REJECT
icmp-port-unreachable
```

阻止所有未授权访问。

---

# 整体数据流

```text
                Client
                   │
                   ▼
          nat/PREROUTING
                   │
        ┌──────────┴──────────┐
        │                     │
        ▼                     ▼
   PORTAL_DNS            PORTAL_HTTP
        │                     │
 DNS 豁免/劫持           HTTP Portal
        │                     │
        └──────────┬──────────┘
                   │
                   ▼
            FORWARD
                   │
                   ▼
             PORTAL_FWD
                   │
        ┌──────────┴──────────┐
        │                     │
   已认证/白名单            未认证
        │                     │
      放行                 REJECT
```

---

# 核心设计思想

整个 Portal 数据面遵循以下原则：

- **Runtime Authoritative**：运行时配置为唯一真相，每次执行均重新同步状态。
- **Reconcile 模型**：通过 Flush + Rebuild 保证配置一致性，避免历史残留。
- **DNS 优先**：DNS Hook 固定位于 `PREROUTING` 第一条规则，确保 DNS 豁免和劫持优先于 HTTP。
- **双层 Domain 豁免**：利用 dnsmasq 将域名解析结果动态写入 `hash:ip` ipset，实现基于域名的豁免。
- **Portal 兼容性**：仅豁免 DNS 查询，不放行基于域名的 Forward 流量，确保各操作系统能够正确触发 Captive Portal 检测与认证流程。
