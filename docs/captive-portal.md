
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
