
# Troubleshooting Guide / 故障排查指南

本文档用于在 **Enterprise AP / Captive Portal Platform** 的开发、部署与运行过程中，快速定位和解决常见问题。  
覆盖 **控制平面、数据平面、Portal 行为、OS 兼容性、网络与部署** 等维度。

This document helps diagnose and resolve common issues encountered during development, deployment, and operation of the **Enterprise AP / Captive Portal Platform**.

---

## 🧭 Troubleshooting Workflow / 推荐排查顺序

> **强烈建议按顺序排查**

1. **基础连通性**
2. **数据平面是否工作**
3. **Portal 劫持是否生效**
4. **控制平面是否可达**
5. **策略 / 状态是否一致**
6. **终端 OS 行为差异**

---

## 1️⃣ Basic Connectivity / 基础连通性问题

### ❌ 现象
- 终端无法获取 IP
- 无法访问任何网站
- Ping 网关失败

### ✅ 排查步骤

```bash
# AP / 网关侧
ip addr
ip route
brctl show      # bridge 场景
```

确认：

* 终端是否获取到 IP
* 默认网关是否正确
* DNS 是否可达

📌 **注意**
Portal 系统 **不应影响 DHCP**，若 DHCP 异常，优先排查：

* bridge / VLAN
* DHCP server
* 防火墙默认策略

---

## 2️⃣ Data Plane Not Working / 数据平面异常

### ❌ 现象

* Portal 页面无法弹出
* 所有流量被阻断
* 已认证用户仍无法上网

### ✅ 排查点

#### 2.1 iptables / nftables 规则

```bash
iptables -t nat -L -n -v
iptables -L FORWARD -n -v
```

确认：

* Portal Redirect 规则是否存在
* 已认证 MAC / IP 是否命中放行规则
* 规则顺序是否正确（**顺序极其重要**）

📌 常见错误：

* 放行规则在 DROP 规则之后
* MASQUERADE 未生效

---

#### 2.2 Portal Agent 进程状态

```bash
ps aux | grep portal-agent
ss -lntp | grep <portal-port>
```

确认：

* 进程是否存活
* 监听端口是否正确
* 是否绑定在正确接口（LAN vs WAN）

---

## 3️⃣ Portal Page Not Popping / Portal 页面不弹出

### ❌ 现象

* 终端可以访问 HTTP，但不跳转
* HTTPS 页面直接失败
* 浏览器显示“无法访问网络”

### ✅ 排查重点

#### 3.1 HTTP vs HTTPS（**最常见问题**）

Portal **只能劫持 HTTP**：

* ❌ HTTPS 无法被重定向
* ✅ 依赖 OS Captive Portal Detection

测试方式：

```bash
curl http://example.com
```

若未跳转：

* Portal 劫持规则未生效
* DNS 未正确劫持

---

#### 3.2 DNS 劫持是否生效

```bash
dig www.baidu.com
```

确认：

* 未认证用户是否被劫持到 Portal IP
* 已认证用户是否恢复正常 DNS

📌 常见错误：

* DNS 规则只匹配 TCP，遗漏 UDP
* IPv6 DNS 未禁用导致绕过

---

## 4️⃣ Control Plane Issues / 控制平面异常

### ❌ 现象

* AP 无法注册
* 策略未下发
* Controller 显示 AP 离线

### ✅ 排查步骤

#### 4.1 Controller API 可达性

```bash
curl http://controller:8080/health
```

确认：

* 容器是否运行
* 端口映射是否正确
* 防火墙是否放行

---

#### 4.2 AP → Controller 心跳

检查 AP 日志：

```bash
journalctl -u portal-agent
```

或容器日志：

```bash
docker logs ap-controller
```

确认：

* 心跳是否周期发送
* 返回码是否为 200
* 时间戳是否更新

---

## 5️⃣ Authenticated but No Internet / 已认证但无法上网

### ❌ 现象

* Portal 显示认证成功
* 但无法访问外网

### ✅ 排查清单

* 是否下发放行规则
* 是否绑定正确的 MAC / IP
* NAT 是否生效

```bash
iptables -t nat -L POSTROUTING -n -v
```

📌 常见错误：

* 用户换 IP（DHCP renew）
* IPv6 流量未放行
* Fast Path 与 iptables 冲突

---

## 6️⃣ OS-Specific Issues / 终端系统差异问题

### 📱 iOS / macOS

❌ 现象：

* Portal 不自动弹出
* Wi-Fi 显示“无互联网”

✅ 排查：

* 是否能访问 `http://captive.apple.com`
* 返回内容是否为 HTTP 200 + HTML

---

### 🤖 Android

❌ 现象：

* 不弹 Portal
* 显示“需要登录网络”

✅ 排查：

* `http://connectivitycheck.gstatic.com/generate_204`
* 返回码是否被劫持（≠204）

---

### 🪟 Windows

❌ 现象：

* Portal 延迟弹出
* 需要手动打开浏览器

✅ 排查：

* `http://www.msftconnecttest.com/connecttest.txt`
* 是否被重定向

📌 **详细 OS 行为见：**

* `docs/os-portal-detection.md`

---

## 7️⃣ Docker / Deployment Issues / 部署问题

### ❌ 现象

* 服务无法启动
* 端口冲突
* 网络不通

### ✅ 排查

```bash
docker ps
docker compose logs
docker network ls
```

确认：

* bridge 网络是否正常
* 端口是否被占用
* 容器是否在同一网络

---

## 8️⃣ Logging & Debugging / 日志与调试建议

### 推荐日志点

* Portal Agent

  * 新终端接入
  * Redirect 命中
  * Auth 状态变化

* Controller

  * AP 注册 / 心跳
  * 策略下发
  * 审计事件

📌 **建议日志字段**

* MAC
* IP
* AP ID
* Session ID
* Timestamp

---

## 9️⃣ Fast Checklist / 快速自检清单

* [ ] DHCP 正常
* [ ] DNS 劫持生效
* [ ] HTTP Redirect 生效
* [ ] 已认证规则优先生效
* [ ] Controller 可达
* [ ] OS 探测 URL 正常返回
* [ ] IPv6 行为符合预期

---

## 🔚 Summary / 总结

> **90% 的 Portal 问题来源于：**

* 规则顺序错误
* HTTPS 误判
* OS 探测 URL 未覆盖
* 已认证放行不完整

遇到复杂问题，**优先用 curl / tcpdump / iptables -v 定位真实数据流路径**。

---

📌 **强烈建议结合阅读：**

* `docs/data-plane.md`
* `docs/os-portal-detection.md`
* `docs/deployment.md`
