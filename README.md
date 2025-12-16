
# Enterprise AP / Captive Portal Platform

一个面向 **企业级 AP（Access Point）与园区网络** 的 **控制平面 + 数据平面 + Captive Portal** 的完整实现示例工程，用于演示 **集中控制、Portal 认证、数据转发与可运维性设计**。

A reference implementation of an **Enterprise Access Point & Captive Portal Platform**, demonstrating **centralized control, captive portal authentication, data-plane forwarding, and operability**.

---

## ✨ Features / 核心特性

- **Control Plane（控制面）**
  - AP Controller（集中控制）
  - 策略下发 / 配置管理
  - 审计与状态存储
  - RESTful API

- **Data Plane（数据面）**
  - AP 侧数据转发
  - Captive Portal 劫持与放行
  - ACL / FDB / iptables 规则管理
  - 可扩展 Fast Path 设计

- **Captive Portal**
  - Web Portal 认证流程
  - Session 生命周期管理
  - 多终端（手机 / PC）兼容

- **Cloud-Native Friendly**
  - Docker / Docker Compose 部署
  - 组件解耦，易于扩展
  - 适合嵌入式 AP + 云端控制器模型

---

## 📁 Project Structure / 项目结构

```text
.
├── control-plane/        # 控制平面（AP Controller / API / 策略）
│   ├── ap-controller/    # Python 实现的控制器示例
│   └── ap-controller-go/ # Go 实现的控制器示例
│
├── data-plane/           # 数据平面（AP 侧逻辑）
│   └── portal-agent/     # Portal 劫持与放行代理
│
├── docs/                 # 📘 项目文档（推荐从这里开始）
│   ├── README.md         # 文档索引
│   ├── overview.md
│   ├── control-plane.md
│   ├── data-plane.md
│   ├── captive-portal.md
│   ├── os-portal-detection.md
│   ├── deployment.md
│   └── healthcheck.md
│
├── docker/               # Docker / Compose 相关文件
├── scripts/              # 辅助脚本
└── README.md             # ← 当前文件
```

---

## 🚀 Quick Start / 快速开始

### 1️⃣ 阅读文档（推荐）

👉 **从这里开始：**

```text
docs/README.md
```

文档包含完整的：

* 架构说明
* 控制面 / 数据面设计
* Portal 行为
* 部署与运维

---

### 2️⃣ 启动控制平面（示例）

```bash
cd control-plane/ap-controller
docker compose up -d
```

或使用 Go 版本：

```bash
cd control-plane/ap-controller-go
docker build -t ap-controller-go .
docker run -p 8080:8080 ap-controller-go
```

---

### 3️⃣ 启动数据平面（Portal Agent）

```bash
cd data-plane/portal-agent
make run
```

（具体参数与运行方式见 `docs/data-plane.md`）

---

## 🧠 Architecture / 架构概览

整体采用 **Controller / Agent** 架构：

* **Controller（控制平面）**

  * 集中管理 AP
  * 下发策略与 Portal 规则
  * 提供统一 API

* **Agent（数据平面）**

  * 驻留在 AP / 网关
  * 处理真实数据流量
  * 执行 Portal 劫持与放行

📌 **完整架构图与时序图请参考：**

* `docs/overview.md`
* `docs/control-plane.md`
* `docs/data-plane.md`

---

## 🧩 Typical Use Cases / 典型场景

* 企业 / 园区 Wi-Fi Portal 认证
* 酒店 / 商场 / 校园网络
* OpenWrt / 嵌入式 AP 二次开发
* Portal / AAA / 接入控制 PoC

---

## 🛠 Tech Stack / 技术栈

* **Control Plane**

  * Python / Go
  * REST API
  * SQLite / Redis（可扩展）

* **Data Plane**

  * Linux networking
  * iptables / nftables
  * Netfilter / TProxy（可选）

* **Deployment**

  * Docker
  * Docker Compose

---

## 📘 Documentation / 文档

📌 **完整文档位于：**

```text
docs/
```

入口文档：

```text
docs/README.md
```

---

## 🧭 Design Philosophy / 设计理念

* **控制面与数据面解耦**
* **逻辑清晰、可演进**
* **贴近真实商用 AP / Portal 架构**
* **适合嵌入式 + 云端混合部署**

---

## 📄 License / 许可

This project is provided as a **reference / educational implementation**.
可用于学习、原型验证与二次开发。

---

## 🙌 Contribution / 贡献

欢迎：

* 架构改进建议
* 数据平面性能优化
* Portal / AAA 扩展
* OpenWrt / 嵌入式适配

---

**👉 下一步推荐阅读：**
➡️ `docs/README.md`
