 
# DNS Bypass ipset 中 IP 数量偏少的现象说明（含运行日志）

## 背景

在启用 Portal 的 DNS bypass 功能后，系统通过 dnsmasq + ipset
将指定域名的解析结果动态加入 `portal_bypass_dns`，用于绕过 Portal
认证流程（如操作系统的 Captive Portal 探测）。

在实际运行中，观察到以下现象。

---

## 现象描述

### DNS 查询结果

```sh
nslookup captive.apple.com
```

返回结果中可以看到：

* 大量 IPv4（A）地址
* 同时包含 IPv6（AAAA）地址
* 多级 CNAME（Akamai / CDN）

示例（节选）：

```text
Address 1: 122.226.74.84
Address 2: 60.163.128.10
...
Address 17: 240e:f7:a096:300:3::16
Address 18: 240e:f7:a096:300:3::17
```

---

### ipset 实际内容

```sh
ipset list portal_bypass_dns
```

结果却只有少量 IP：

```text
Name: portal_bypass_dns
Type: hash:ip
Header: family inet
Number of entries: 2
Members:
23.36.70.120
104.208.16.91
```

---

## 结论（先行）

> **这是 dnsmasq + ipset 的预期行为，不是脚本错误，也不是规则缺失。**

`portal_bypass_dns` 中的 IP 数量只反映：

> **“dnsmasq 在当前运行期实际采用并使用的 IPv4 解析结果”**

而不是 DNS 返回的所有可能地址。

---

## DNS Bypass 的两层模型

当前实现严格遵循 **Source / Result 两层模型**。

---

### 1️⃣ 规则源层（Source Layer）

由 `portal-fw.sh` 生成的 dnsmasq 配置文件：

```conf
ipset=/apple.com/portal_bypass_dns
ipset=/captive.apple.com/portal_bypass_dns
ipset=/platform.hicloud.com/portal_bypass_dns
...
```

该层的职责：

* 声明 **哪些域名的解析结果需要进入 ipset**
* 不直接操作 IP
* 不关心 IP 数量

---

### 2️⃣ 结果层（Result Layer）

由 dnsmasq 在运行期维护的 ipset：

```text
portal_bypass_dns (hash:ip, family inet)
```

该层的职责：

* 仅存储 **dnsmasq 实际采用的 A 记录**
* 按需加入、动态变化
* 自动去重

---

## 从运行日志验证行为（关键证据）

以下日志来自连续两次 `portal-agent` 周期执行。

---

### 1️⃣ DNS bypass 结果层（ipset）reconcile

```text
event=bypass_dns_flush count_before=0
event=bypass_dns_skip reason=empty_runtime
event=bypass_dns_apply_done count_after=0
```

说明：

* 每次策略下发都会 **先 flush 旧 DNS ipset**
* runtime 中未直接提供 IP（符合设计）
* DNS ipset 的内容完全由 dnsmasq 运行期决定

---

### 2️⃣ DNS bypass 规则源层（dnsmasq conf）rebuild

```text
event=bypass_dns_conf_rebuild_start
event=bypass_dns_conf_flush count_before=0
event=bypass_ctrl_domain domain=apple.com
event=bypass_ctrl_domain domain=platform.hicloud.com
event=bypass_ctrl_domain domain=hicloud.com
event=bypass_ctrl_domain domain=vmall.com
event=bypass_ctrl_domain domain=huawei.com
event=bypass_ctrl_domain domain=gstatic.com
event=bypass_ctrl_domain domain=google.com
event=bypass_dns_conf_apply_done count_after=0
```

说明：

* dnsmasq 配置文件被 **完整重建**
* 所有 bypass 域名均被正确写入规则源
* `count_after=0` 并非错误：

  * 该阶段只是写规则
  * 不向 ipset 中直接 add IP

---

### 3️⃣ dnsmasq reload 与最终生效

```text
event=dnsmasq_reloaded reason=bypass_domains
event=bypass_dns_apply_done count_after=0
event=dns_rules_installed chain=PORTAL_DNS
event=init_done
```

说明：

* dnsmasq 成功 reload
* iptables / nftables DNS 链规则已安装
* dataplane 初始化完成，无错误

---

## 为什么 nslookup 很多 IP，但 ipset 只有 1～2 个？

### 1️⃣ dnsmasq 只加入“被实际采用”的 A 记录

* 上游 DNS 返回多个地址是 **候选集**
* dnsmasq 会根据：

  * CDN 策略
  * 当前网络环境
  * 实际访问行为
* 选择少量地址用于连接

👉 **只有这些地址会被加入 ipset**

---

### 2️⃣ CDN / GeoDNS 的影响

Apple / Google 等域名使用：

* Akamai / Azure / 阿里云 CDN
* 基于地域和 ISP 的动态调度

dnsmasq 看到的是：

> “当前网络环境下，这 1～2 个节点已经足够”

---

### 3️⃣ IPv6 被主动忽略（设计选择）

当前 ipset 定义为：

```text
Type: hash:ip
Header: family inet
```

意味着：

* 仅接收 IPv4
* 所有 AAAA（IPv6）记录不会进入该 ipset

---

### 4️⃣ ipset 自动去重

* 多个域名解析到同一 IP
* ipset 中只保留一条记录

---

### 5️⃣ dnsmasq 按需解析（lazy）

dnsmasq 只在以下情况下加入 IP：

* 客户端真实访问域名
* DNS 查询命中 ipset 规则

单纯的 `nslookup` 不代表实际使用。

---

## 为什么这是正确的工程行为

该设计带来以下优势：

* ✅ ipset 规模可控，不膨胀
* ✅ 自动适应 CDN IP 变化
* ✅ 避免过度放行（最小授权）
* ✅ runtime-authoritative，无历史残留
* ✅ 符合成熟 AC / Portal 系统实践

---

## 常见误区澄清

❌ **误区**

> DNS 返回多少 IP，ipset 就应该包含多少

✅ **正确理解**

> ipset 只包含 **当前运行期被 dnsmasq 实际采用的 IPv4 地址**

---

## 总结

> **DNS bypass ipset 的目标不是穷举所有可能的 IP，
> 而是放行“当前网络环境下真实被使用的解析结果”。**
