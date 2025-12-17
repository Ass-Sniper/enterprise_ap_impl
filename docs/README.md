# Documentation Index / 文档索引

本目录包含 **Enterprise AP / Captive Portal 平台** 的完整技术文档，覆盖总体架构、控制面、数据面、Portal 行为、部署与运维等内容。

This directory contains the complete technical documentation for the **Enterprise AP / Captive Portal Platform**, covering architecture, control plane, data plane, portal behavior, deployment, and operations.

---

## 📘 Overview / 总览

- **[overview.md](./overview.md)**  
  **System Overview & Architecture**  
  系统总体介绍、设计目标、组件划分与核心能力概览。  
  High-level system description, goals, components, and core capabilities.

---

## 🧠 Architecture & Components / 架构与组件

- **[control-plane.md](./control-plane.md)**  
  **Control Plane**  
  AP Controller、策略下发、审计、配置管理、API 设计。  
  AP controller, policy distribution, auditing, configuration management, APIs.

- **[data-plane.md](./data-plane.md)**  
  **Data Plane**  
  AP 侧数据转发、Portal 劫持、ACL/FDB/iptables、Fast Path 设计。  
  Packet forwarding, portal interception, ACL/FDB/iptables, fast-path design.

- **[captive-portal.md](./captive-portal.md)**  
  **Captive Portal**  
  Portal 认证流程、重定向机制、会话生命周期、常见实现模式。  
  Authentication flow, redirection, session lifecycle, common portal patterns.

---

## 🔄 System Behavior / 系统行为

- **[os-portal-detection.md](./os-portal-detection.md)**  
  **OS Captive Portal Detection**  
  Android / iOS / Windows / macOS 的 Portal 探测机制与适配策略。  
  OS-specific captive portal detection mechanisms and handling strategies.

- **[healthcheck.md](./healthcheck.md)**  
  **Health Check & Monitoring**  
  AP / Controller 健康探测、心跳机制、故障检测与恢复。  
  Health probing, heartbeat mechanisms, failure detection and recovery.

---

## 🚀 Deployment & Operations / 部署与运维

- **[deployment.md](./deployment.md)**  
  **Deployment Guide**  
  Docker / Docker Compose 部署方式，配置示例，启动与升级流程。  
  Docker & Docker Compose deployment, configuration examples, startup and upgrade.

---

## 🧭 Recommended Reading Order / 推荐阅读顺序

1. `overview.md`
2. `control-plane.md`
3. `data-plane.md`
4. `captive-portal.md`
5. `os-portal-detection.md`
6. `deployment.md`
7. `healthcheck.md`

---

## 📎 Notes / 说明

- 所有文档均为 **设计级 + 实现级**，适合开发、调试与运维人员。
- 架构图与时序图统一
