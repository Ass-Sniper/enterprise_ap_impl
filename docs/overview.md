
# Enterprise AP Captive Portal – Overview

## Introduction

本项目实现了一套 **企业级 AP Captive Portal（强制门户）系统**，
采用 **Control Plane / Data Plane / Portal Plane 三层解耦架构**，
面向 OpenWrt / ImmortalWrt 等嵌入式 AP 场景。

系统重点解决的问题包括：

- 现代操作系统 Captive Portal 探测兼容性
- HTTPS 普及背景下的 Portal 正确弹出
- AP 侧策略幂等、可回滚、可自检
- 控制面与数据面的清晰解耦

This project implements an **enterprise-grade captive portal system**
with a clear separation between control plane, data plane, and portal plane,
designed for OpenWrt / ImmortalWrt-based access points.

---

## What Problem Does This Project Solve?

### 1. Captive Portal in the HTTPS Era

传统 Captive Portal 方案通常依赖：

- HTTP 劫持
- DNS 污染
- HTTPS 中间人（MITM）

在现代操作系统（Windows / iOS / Android / HarmonyOS）中，
这些手段要么失效、要么导致严重的用户体验问题，
例如 Portal 不弹、反复弹窗、网络被判定为不可用等。

---

### 2. Core Idea of This Project

> **Do not intercept HTTPS. Use DNS as the only implicit signal.**

本项目通过 dnsmasq + ipset 的组合：

- 在 **DNS 解析阶段**识别操作系统的 Portal 探测域名
- 将解析得到的目标 IP **动态写入 ipset**
- 防火墙基于 ipset 进行放行或重定向决策

这种方式与操作系统原生行为完全一致，避免了协议层破坏。

---

## High-Level Architecture

系统被清晰地拆分为三个逻辑平面：

```text
+---------------------+
|   Control Plane     |
|  Policy & Runtime   |
+---------------------+
           |
           v
+---------------------+
|    Data Plane       |
| DNS / IPSet / FW    |
+---------------------+
           |
           v
+---------------------+
|   Captive Portal    |
|   User-facing UI    |
+---------------------+
```

---

## Control Plane

控制面是系统的 **策略中枢（Single Source of Truth）**，负责：

* 统一管理 Portal、Bypass、Dataplane 策略
* 通过 HTTP API 向 AP 下发 Runtime
* 维护策略版本（Policy Version）
* 记录审计日志，支持回溯与分析
* 支持多 AP / 多 Site 部署模型

实现形式：

* **Python**：参考实现 / PoC
* **Go**：生产导向实现，强调并发与可维护性

---

## Data Plane

数据面运行在 AP（OpenWrt / ImmortalWrt）设备上，负责：

* DNS 处理（dnsmasq）
* 状态集合维护（ipset）
* 流量控制与重定向（iptables）
* 策略应用、回滚与健康检查

关键特性：

* AP 无状态（Stateless）
* 策略幂等（Idempotent）
* 可重复执行、失败可恢复
* 内建 Self-Test 与 Health Check

---

## Captive Portal Plane

Portal Plane 是系统中 **唯一面向最终用户的组件**，负责：

* 展示 Portal 页面
* 引导用户完成认证 / 确认
* 返回认证结果页面

设计原则：

* Portal **不直接操作防火墙**
* 所有放行逻辑由 Data Plane 决定
* Portal 实现可替换、可扩展

---

## Key Design Principles

### 1️⃣ Control / Data Plane Decoupling

* Controller 永不直接修改 iptables / ipset
* AP 不保存长期策略，仅执行 Runtime
* Runtime 是唯一事实源

---

### 2️⃣ DNS-first Portal Detection

* DNS 是唯一决策信号
* 不破坏 HTTPS 语义
* 与 OS Portal Detection 行为天然兼容

---

### 3️⃣ Idempotency & Safety

* 所有数据面脚本可重复执行
* 不依赖初始状态
* 任意失败点均可回滚

---

### 4️⃣ Observability by Design

* 内建 `portal-agent --check`
* DNS → IPSet → Firewall 全链路可验证
* 失败可检测、可定位、可自动处理

---

## Typical Deployment Scenario

```text
[Client Device]
       |
       v
[AP / Data Plane]
       |
       v
[Controller + Portal]
```

* AP 定期或按需拉取 Runtime
* 客户端触发 OS Portal Detection
* Portal 自动弹出
* 用户完成操作后网络被放行

---

## Target Audience

* 企业 / 园区 Wi-Fi 网络
* 酒店 / 公共接入网络
* OpenWrt / ImmortalWrt AP 场景
* 嵌入式网络系统开发者
* Captive Portal 架构学习与参考

---

## Project Status

* 架构稳定
* 功能完整
* 文档齐全
* 适合作为 **生产参考实现或教学样板**

---

## Summary

本项目并非传统意义上的“Portal 脚本集合”，
而是一套：

* 架构清晰
* 行为可验证
* 设计可演进

的 **现代 Captive Portal 工程实现**。

它展示了一种在 HTTPS 普及的网络环境下，
**依然能够正确、优雅实现 Captive Portal 的工程方法论**。