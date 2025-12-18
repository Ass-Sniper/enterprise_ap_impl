
# Captive Portal

## Overview

Captive Portal 是用户侧可见的 **认证与引导入口**，用于在用户尚未完成认证时，
通过 Web 页面引导其完成登录、授权或确认操作。

本项目中的 Captive Portal **与数据面解耦**：
- 不直接操作 iptables / ipset
- 仅通过 HTTP + 间接状态联动完成认证闭环

The captive portal provides the **user-facing authentication and onboarding UI**.
It is intentionally decoupled from the dataplane and never manipulates firewall
rules directly.

---

## Responsibilities

Captive Portal 的核心职责包括：

- 向未认证用户展示 Portal 页面
- 处理登录 / 确认 / 授权请求
- 返回认证结果（成功 / 失败）
- 为数据面“放行”提供决策依据（间接）

---

## Directory Structure

```text
control-plane/captive-portal/
├── portal-agent          # （预留）与 AP 通信的代理
├── portal-server         # Portal Web Server
│   ├── app/
│   │   ├── main.py       # Portal HTTP server
│   │   ├── static/       # JS / CSS / assets
│   │   └── templates/    # HTML templates
│   ├── Dockerfile
│   └── requirements.txt
└── README.md
```

---

## Portal Server

### Technology Stack

* Python
* Flask（轻量 Web Framework）
* HTML / CSS / JavaScript（前端）

---

### Entry Point

```text
portal-server/app/main.py
```

该服务负责：

* HTTP 路由注册
* Portal 页面渲染
* 表单处理
* 认证结果反馈

---

## Portal Pages

### 1️⃣ Portal Page

```text
templates/portal.html
```

功能：

* 向用户展示 Portal 页面
* 显示认证提示
* 提供“继续 / 登录”按钮

---

### 2️⃣ Result Page

```text
templates/result.html
```

功能：

* 显示认证结果
* 成功 / 失败提示
* 指引用户刷新或重新连接网络

---

## Static Assets

```text
static/
├── app.js
├── style.css
└── favicon.ico
```

* `app.js`：前端逻辑（按钮 / 提交）
* `style.css`：页面样式
* `favicon.ico`：浏览器图标

---

## Authentication Model

### Current Model

当前 Portal 实现采用 **简化认证模型**：

* Portal 接收用户请求
* 返回“认证成功”或“失败”结果
* 数据面通过 **IP / MAC / Session** 等方式完成放行

---

### Design Note

> Portal Server 不应直接调用 AP 的防火墙接口
> 所有放行逻辑均应通过数据面脚本完成

这样做的好处：

* 降低耦合
* 提高安全性
* 便于未来替换 Portal 实现

---

## Integration with Data Plane

### Traffic Flow

1. 未认证客户端访问外网
2. 数据面检测未授权
3. HTTP 流量被重定向至 Portal Server
4. Portal Server 返回 Portal 页面
5. 用户完成操作
6. 数据面根据策略放行流量

---

### Key Principle

* Portal 只负责 **“用户交互”**
* Data Plane 负责 **“是否放行”**
* Control Plane 负责 **“策略决策”**

---

## OS Compatibility

为了兼容主流操作系统的 Captive Portal 行为：

* Portal 页面使用 HTTP
* 避免 HTTPS 劫持
* 与 DNS bypass 机制协同工作

支持的系统包括：

* Windows
* iOS / macOS
* Android
* HarmonyOS / Huawei

---


## Domain Normalization for OS Probe Bypass

> 本节说明 Captive Portal 在处理各类操作系统（Windows / iOS / Android / 厂商定制系统）
> 探测流量时，如何通过“域名归一化（Domain Normalization）”机制，在控制面与数据面之间
> 建立稳定、可扩展的 Bypass 策略。

### Background and Motivation

各类操作系统在判断网络是否可访问互联网时，都会访问一组**固定但不完全稳定的探测域名**
（详见 [OS Portal Detection](os-portal-detection.md)）。

这些探测请求必须被 Portal **有选择地放行**，否则操作系统将不会弹出登录页面或直接判定
网络不可用。

然而，不同操作系统：

- 使用的探测域名并不统一
- 同一 OS 在不同版本中可能更换探测域名
- 探测域名通常存在多个子域变体

为解决上述问题，Portal 在控制面与数据面之间引入了**域名归一化机制**。

---

### Multi-Layer Domain Model

Portal 系统在不同层面对“域名”的使用具有不同语义，可分为以下三层。

#### Client OS Layer (Exact FQDN)

在客户端侧，操作系统使用**完整限定域名（FQDN）**进行探测，例如：

- Windows: `www.msftconnecttest.com`, `dns.msftncsi.com`
- Apple (iOS/macOS): `captive.apple.com`
- Android: `clients3.google.com`, `connectivitycheck.gstatic.com`
- 厂商定制系统: `connectivitycheck.platform.hicloud.com`

该层要求**精确匹配**，任何 DNS 或 HTTP 行为异常都会导致 OS 探测失败。

---

#### Control Plane Layer (Policy Declaration)

在控制面（Controller → Agent）中，Portal 使用**明确的 FQDN 列表**
表达策略意图，例如：

```json
[
  "www.msftconnecttest.com",
  "dns.msftncsi.com",
  "captive.apple.com",
  "clients3.google.com",
  "connectivitycheck.gstatic.com"
]
```

该层强调：

* 可读性与可维护性
* 与操作系统官方文档的一致性
* 清晰表达“哪些 OS 探测行为需要被支持”

控制面关注的是**策略语义**，而非数据面的具体匹配方式。

---

#### Data Plane Layer (Normalized Domains)

在数据面（dnsmasq / iptables / nftables）中，Portal 会将控制面声明的
FQDN **归一化为可注册主域**，以减少规则数量并提升鲁棒性：

| Control Plane FQDN              | Normalized Domain     |
| ------------------------------- | --------------------- |
| `www.msftconnecttest.com`       | `msftconnecttest.com` |
| `dns.msftncsi.com`              | `msftncsi.com`        |
| `captive.apple.com`             | `apple.com`           |
| `clients3.google.com`           | `google.com`          |
| `connectivitycheck.gstatic.com` | `gstatic.com`         |

这一过程称为 **Domain Normalization**。

---

### Why Domain Normalization Is Required

#### OS Probe Variability

操作系统厂商在历史版本中多次调整探测域名，例如 Android 曾使用过：

* `clients3.google.com`
* `connectivitycheck.gstatic.com`
* `connectivitycheck.android.com`

如果数据面严格依赖精确 FQDN，系统将对 OS 版本变化高度敏感，
并显著增加维护成本。

---

#### Data Plane Constraints

数据面规则需要满足：

* 数量尽可能少
* 匹配逻辑尽可能简单
* 避免大量精确字符串匹配

通过将多个 FQDN 归一化为单一主域，可以显著降低规则复杂度。

---

#### Forward Compatibility

域名归一化使系统能够：

* 自动覆盖未来新增的探测子域
* 减少因 OS 升级导致的 Portal 行为变化
* 提升系统的长期稳定性

---

### Security Considerations

域名归一化**并不意味着完全放行该域名的所有流量**。

Portal 在数据面配合以下限制措施（详见 [Firewall and DNS Enforcement](data-plane.md)）：

* **仅放行 DNS 查询**
* **HTTPS 流量保持阻断或 TCP Reset**
* **HTTP 流量仍由 Portal 接管**

示意如下：

```text
DNS   → Allowed
HTTP  → Captive Portal
HTTPS → Blocked / TCP Reset
```

---

### Summary

通过域名归一化机制，Portal 在以下三者之间建立了清晰分层：

* 操作系统精确的探测行为
* 控制面语义化的策略声明
* 数据面高效、稳健的执行规则

该设计在保证 OS Captive Portal 探测可靠性的同时，避免了数据面规则膨胀，
是商用 AC / 网关设备的通用工程实践。

---

## Nginx Integration on ImmortalWrt (MT798x)

> 本节描述如何在 **不影响 uhttpd / LuCI 管理面的前提下**，将 nginx 集成到 ImmortalWrt，
> 并作为 Captive Portal 的 **HTTP Gateway（Portal Gateway）** 使用。
> 该步骤是后续 Header 注入、HMAC Token、auth_request、本地 token-signer 的基础。

### Design Goals

- nginx 仅作为 **数据面 Portal Gateway**，不承担管理面职责
- uhttpd 继续监听 80 / 443，用于设备管理（LuCI）
- 客户端 HTTP 80 是否进入 Portal，由 **iptables DNAT** 决定
- nginx 不使用 UCI 配置模式，采用 **原生 nginx.conf + conf.d**，便于脚本或 Agent 动态生成

---

### Configuration Model in ImmortalWrt

ImmortalWrt 默认使用 **UCI 驱动 nginx**，其配置链路如下：

```text
/etc/config/nginx
   ↓
/etc/nginx/uci.conf.template
   ↓
/var/lib/nginx/uci.conf
```

因此系统默认 **不会提供 `/etc/nginx/nginx.conf`**。

对于 Captive Portal 场景，该模式不利于复杂路由、Header 注入与动态生成配置，
因此需要切换到 **原生 nginx.conf 模式**。

---

### Switching to Native nginx.conf Mode

#### Native Configuration Files

创建 `/etc/nginx/nginx.conf`：

```nginx
user root;
worker_processes 1;

events {
    worker_connections 1024;
}

http {
    include mime.types;
    default_type application/octet-stream;
    sendfile on;
    keepalive_timeout 30;
    include /etc/nginx/conf.d/*.conf;
}
```

Portal Gateway 占位配置（监听 8081）：

```nginx
server {
    listen 8081;

    location / {
        return 200 "nginx portal gateway alive\n";
    }
}
```

---

#### Init Script Adjustment

由于 `/etc/init.d/nginx` 默认会优先使用 `nginx-util` 生成的 UCI 配置，
需要在 init 脚本中显式禁用该逻辑。

推荐方式是在脚本中引入 **原生配置开关**：

```sh
USE_NATIVE_CONF=1
```

并在 `nginx_init()` 中调整配置选择逻辑：

```sh
if [ "${USE_NATIVE_CONF}" = "1" ]; then
    CONF="/etc/nginx/nginx.conf"
else
    rm -f "$(readlink "${UCI_CONF}")"
    ${NGINX_UTIL} init_lan

    if [ -e "${UCI_CONF}" ]; then
        CONF="${UCI_CONF}"
    else
        CONF="${NGINX_CONF}"
    fi
fi
```

该方式保留 procd 的 respawn / reload 机制，
同时完全脱离 UCI 配置模型。

---

### Startup and Verification

```bash
nginx -t -c /etc/nginx/nginx.conf
/etc/init.d/nginx restart
/etc/init.d/nginx enable
```

验证：

```bash
ps | grep nginx
netstat -lntp | grep 8081
curl http://127.0.0.1:8081
```

---

### Relationship with uhttpd

* **uhttpd**：继续运行，负责 LuCI / 管理面
* **nginx**：仅监听 8081，不直接暴露给客户端
* **iptables**：决定客户端 HTTP 80 是否被 DNAT 至 Portal Gateway

---

### Portal HTTP Traffic Flow

<pre class="mermaid">
flowchart LR
    Client[LAN Client]
    Iptables[iptables DNAT&lt;br/&gt;PORTAL_HTTP]
    Nginx[nginx :8081&lt;br/&gt;Portal Gateway]
    Portal[Portal Server]
    Controller[Controller]

    Client -- TCP 80 --&gt; Iptables
    Iptables -- DNAT --&gt; Nginx
    Nginx --&gt; Portal
    Portal --&gt; Controller

    subgraph Router
        Iptables
        Nginx
    end
</pre>

---

### Phase Conclusion

完成以上步骤后：

* nginx 已稳定集成到 ImmortalWrt
* 不依赖 UCI / LuCI 配置模型
* 不影响管理面（uhttpd）
* Portal HTTP 流量具备 **可控入口点**

这是后续实现 OS 探测处理、Header 注入、HMAC Token、auth_request 的前提。

---

## Portal Gateway Authentication Context (Header Injection & HMAC)

> 本节描述 **Portal Gateway（nginx）如何在数据面构造可信认证上下文**，
> 通过 **HTTP Header 注入 + HMAC Token** 的方式，将客户端身份、接入点信息安全地传递给 Portal Server。
> 该机制用于替代“纯参数透传”，防止伪造、篡改和重放。

---

### 1. 设计动机

在 Captive Portal 场景中，Portal Server 需要可靠地获取以下信息：

* 客户端身份：MAC / IP
* 接入上下文：SSID / AP ID / Radio ID
* 访问语义：Portal 探测 / 登录 / 业务访问

如果这些信息 **完全由客户端提交（Query/Form）**，将存在：

* 参数可伪造
* 无法区分真实 AP 注入还是客户端构造
* 难以在控制面统一校验

因此，本方案将 **可信边界前移到 Portal Gateway（nginx）**。

---

### 2. 信任模型与责任划分

```
Client  ──(HTTP)──▶  nginx (Portal Gateway)  ──▶  Portal Server
             ↑              ↑
        不可信           可信注入点
```

* Client：不可信
* nginx（运行在 AP / Router）：可信执行环境
* Portal Server：信任来自 nginx 的 Header + Token

---

### 3. Header 注入规范

Portal Gateway 在转发请求至 Portal Server 前，统一注入以下 Header：

| Header 名称            | 说明          |
| -------------------- | ----------- |
| `X-Portal-MAC`       | 客户端 MAC 地址  |
| `X-Portal-IP`        | 客户端 IPv4 地址 |
| `X-Portal-SSID`      | 当前接入 SSID   |
| `X-Portal-AP-ID`     | AP 唯一标识     |
| `X-Portal-Radio-ID`  | 无线射频标识      |
| `X-Portal-Timestamp` | Unix 时间戳（秒） |
| `X-Portal-Nonce`     | 随机字符串，防重放   |
| `X-Portal-Signature` | HMAC 签名     |

客户端提交的同名 Header **必须被清理或覆盖**。

---

### 4. HMAC Token 设计

#### 4.1 签名输入

签名原文（Canonical String）：

```
METHOD 

PATH 

MAC 

IP 

SSID 

AP_ID 

RADIO_ID 

TIMESTAMP 

NONCE
```

示例：

```
POST
/portal/login
62:0d:f1:32:b6:69
192.168.16.155
GuestWiFi
ap-01
radio0
1766025000
9f3a1c8e
```

---

#### 4.2 签名算法

* 算法：`HMAC-SHA256`
* 密钥：`PORTAL_SHARED_SECRET`
* 输出：Hex 或 Base64

```
X-Portal-Signature = HMAC(secret, canonical_string)
```

---

### 5. nginx 中的实现方式（示意）

```nginx
set $portal_ts  $msec;
set $portal_nonce  $request_id;

proxy_set_header X-Portal-MAC        $remote_addr;   # 实际实现由 fw / map 提供
proxy_set_header X-Portal-IP         $remote_addr;
proxy_set_header X-Portal-SSID       $portal_ssid;
proxy_set_header X-Portal-AP-ID      $portal_ap_id;
proxy_set_header X-Portal-Radio-ID   $portal_radio_id;
proxy_set_header X-Portal-Timestamp  $portal_ts;
proxy_set_header X-Portal-Nonce      $portal_nonce;
proxy_set_header X-Portal-Signature  $portal_hmac;
```

> 注：`$portal_hmac` 通常由 **auth_request + 本地 signer 服务** 生成，而不是直接在 nginx 中计算。

---

### 6. Portal Server 侧校验流程

1. 读取所有 `X-Portal-*` Header
2. 校验时间戳（窗口，例如 ±300s）
3. 校验 Nonce 是否已使用
4. 按同一 Canonical 规则重建字符串
5. 使用共享密钥计算 HMAC
6. 对比 `X-Portal-Signature`

校验失败：

* 返回 `401 Unauthorized`
* 不触发登录 / 放行逻辑

---

### 7. 与 Phase 1 / Phase 2 的关系

* **Phase 1（HTTP 劫持）**：iptables → nginx
* **Phase 2（认证上下文构造）**：Header + HMAC（本节）
* **Phase 3（控制面下发）**：Portal Server → Controller → portal-agent

---

### 8. 阶段性结论

通过 Header 注入 + HMAC Token：

* Portal Server 不再信任客户端参数
* AP / Router 成为唯一可信身份注入点
* 控制面与数据面解耦
* 为 auth_request、本地 signer、零信任扩展奠定基础

---

## Deployment

### Docker Deployment

```bash
cd control-plane/captive-portal
docker build -t portal-server .
```

Portal Server 通常通过 `docker-compose` 与 Controller 一同启动。

---

## Security Considerations

* Portal 不处理敏感凭据（当前实现）
* 不存储用户密码
* 不直接暴露数据面接口
* 所有外部访问通过反向代理或 NAT

---

## Extensibility

Captive Portal 可扩展方向：

* OAuth / SMS / 企业账号认证
* 用户协议确认
* Session / Token 机制
* 与 RADIUS / AAA 系统集成

---

## Summary

Captive Portal 是系统中 **唯一面向最终用户的组件**，
其设计重点是：

* 简单
* 解耦
* 可替换

通过将复杂逻辑留在数据面与控制面，
Portal 本身保持轻量、可维护、可扩展。
