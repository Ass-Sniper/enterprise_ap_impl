

## Build & Compile Guide

本项目的 **Captive Portal Server** 支持 **C++ / Python 双实现**，通过 **Docker 多阶段构建 + build args** 进行选择。
默认推荐 **C++ 实现（高性能 / 可控依赖）**，Python 实现主要用于快速验证与调试。

---

### 1. 构建方式总览

| 实现方式     | 说明                                  | 适用场景       |
| -------- | ----------------------------------- | ---------- |
| `cpp`    | 基于 **Restbed + OpenSSL** 的 C++11 实现 | 生产环境 / 高并发 |
| `python` | 基于 **FastAPI + Uvicorn**            | 开发 / 原型验证  |

构建选择由 Docker build 参数控制：

```text
PORTAL_IMPL = cpp | python
```

---

### 2. 目录结构说明

```text
captive-portal/
├── Dockerfile              # 多阶段构建入口（cpp / python selector）
├── portal-server/          # Python 实现
│   ├── app/
│   ├── templates/
│   ├── static/
│   └── requirements.txt
└── portal-server-cpp/      # C++ 实现
    ├── CMakeLists.txt
    ├── src/
    └── include/
```

---

### 3. C++ 实现构建说明（推荐）

#### 3.1 构建依赖

* Ubuntu 22.04（构建阶段）
* C++11
* CMake ≥ 3.16
* OpenSSL 3.x
* Restbed（指定版本源码编译）

所有依赖**已封装在 Dockerfile 中**，宿主机无需安装。

---

#### 3.2 Docker 构建（C++ 实现）

在 `control-plane` 目录下执行：

```bash
docker build \
  -t captive-portal-server \
  -f captive-portal/Dockerfile \
  captive-portal
```

或显式指定参数：

```bash
docker build \
  -t captive-portal-server \
  -f captive-portal/Dockerfile \
  --build-arg PORTAL_IMPL=cpp \
  --build-arg RESTBED_VERSION=4.8 \
  --build-arg OPENSSL_PKG="libssl-dev=3.0.*" \
  captive-portal
```

构建流程说明：

1. **cpp-builder**

   * 编译 Restbed（指定版本）
   * 编译 `portal-server-cpp`
2. **cpp-runtime**

   * 仅包含运行所需的 `libssl3` + 二进制
   * 镜像体积最小化

---

#### 3.3 运行（C++）

```bash
docker run -d \
  -p 8080:8080 \
  --name captive-portal-server \
  captive-portal-server
```

---

### 4. Python 实现构建说明（可选）

Python 实现主要用于开发调试，**不建议用于高并发生产环境**。

#### 4.1 Docker 构建（Python 实现）

```bash
docker build \
  -t captive-portal-server \
  -f captive-portal/Dockerfile \
  --build-arg PORTAL_IMPL=python \
  captive-portal
```

#### 4.2 运行（Python）

```bash
docker run -d \
  -p 8080:8080 \
  --name captive-portal-server \
  captive-portal-server
```

---

### 5. 使用 Docker Compose（推荐）

在项目根目录（`control-plane/`）：

```bash
docker compose build captive-portal-server
docker compose up -d captive-portal-server
```

对应配置（节选）：

```yaml
captive-portal-server:
  build:
    context: ./captive-portal
    dockerfile: Dockerfile
    args:
      PORTAL_IMPL: cpp
  ports:
    - "8080:8080"
```

---

### 6. 常见问题（FAQ）

#### Q1: `failed to read dockerfile: no such file or directory`

请确认 **构建上下文和 Dockerfile 路径匹配**：

```bash
docker build -f captive-portal/Dockerfile captive-portal
```

---

#### Q2: 如何切换 C++ / Python 实现？

只需修改 build arg：

```text
PORTAL_IMPL=cpp     # 默认
PORTAL_IMPL=python
```

---

#### Q3: 为什么 C++ 实现需要源码编译 Restbed？

* 发行版中 Restbed 版本不可控
* 与 OpenSSL 版本强绑定
* 生产环境需要 ABI 稳定性

---

### 7. 设计说明（简要）

* **C++ Portal**：高性能 HTTP / HTTPS / 状态机友好
* **Python Portal**：开发效率优先
* **接口层完全一致**（URI / JSON / HMAC）
* 可无缝接入：

  * FreeRADIUS
  * ap-controller
  * WebSocket / Agent 模式（规划中）

---

下一步：

* ✅ **portal-server-cpp 的 CMakeLists.txt 注释说明**
* ✅ **C++ / Python Portal 行为对齐表**
* ✅ **README + 架构图 + Mermaid 时序图整合版**
* ✅ **一键 `make docker-cpp / docker-python`**

