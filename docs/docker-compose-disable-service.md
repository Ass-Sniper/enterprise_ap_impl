
# Docker Compose 中如何“注释 / 禁用”服务

在 `docker-compose.yml` 中，YAML 本身**不支持块注释**，因此“注释掉一个服务”在实践中有几种不同做法。下面按**推荐程度**说明。

---

## 方法一：使用 `#` 逐行注释（临时调试用）

> 适合：临时禁用某个服务  
> 不适合：长期维护、频繁开关

```yaml
services:
  # captive-portal-server:
  #   build:
  #     context: ./captive-portal
  #   container_name: captive-portal-server
  #   environment:
  #     - PORTAL_PORT=8080
  #   ports:
  #     - "8080:8080"
```

### 特点

* ✅ 简单直观
* ❌ 注释多了很乱
* ❌ 不能通过命令行灵活控制

---

## ⭐ 方法二：使用 `profiles`（强烈推荐，工程化方案）

这是 **Docker Compose 官方推荐** 的服务启停方式，**不需要改 YAML 文件本身**。

### 1️⃣ 给服务加 `profiles`

```yaml
services:
  freeradius:
    build: ./radius-stack/freeradius
    container_name: freeradius

  captive-portal-server:
    profiles: ["portal"]
    build: ./captive-portal
    container_name: captive-portal-server
    ports:
      - "8080:8080"
```

### 2️⃣ 默认启动（不包含 profile 的服务）

```bash
docker compose up -d
```

只会启动 `freeradius`，**不会启动** `captive-portal-server`。

### 3️⃣ 启动指定 profile 的服务

```bash
docker compose --profile portal up -d
```

### 4️⃣ 启动多个 profile

```bash
docker compose --profile portal --profile controller up -d
```

### 优点

* ✅ 不需要改动 compose 文件
* ✅ 适合开发 / CI / 生产不同组合
* ✅ 非常适合多服务 control-plane 项目
* ✅ 比“注释 YAML”专业得多

---

## 方法三：`deploy.replicas: 0`（了解即可，不推荐）

```yaml
services:
  captive-portal-server:
    deploy:
      replicas: 0
```

### 注意

* 这是 **Docker Swarm** 的语义
* 在 `docker compose`（非 swarm）模式下：

  * 行为不稳定
  * 经常被忽略

👉 **不推荐在本地或开发环境使用**

---

## ❌ 不推荐的做法

### ❌ 删除服务名，只保留配置

```yaml
services:
  build: ./xxx
```

会导致：

* YAML 结构错误
* `docker compose` 直接失败

---

## 推荐的 control-plane 结构示例

```yaml
services:
  mysql:
    image: mysql:8.0

  redis:
    image: redis:7-alpine

  freeradius:
    build: ./radius-stack/freeradius

  captive-portal-server:
    profiles: ["portal"]
    build: ./captive-portal

  ap-controller:
    profiles: ["controller"]
    build: ./ap-controller-go
```

常用命令：

```bash
# 只启动基础设施
docker compose up -d

# 启动 portal
docker compose --profile portal up -d

# 启动 portal + controller
docker compose --profile portal --profile controller up -d
```

---

## 总结

| 场景         | 推荐方式              |
| ---------- | ----------------- |
| 临时禁用       | `#` 注释            |
| 长期可控 / 多环境 | `profiles` ⭐      |
| Swarm 部署   | `deploy.replicas` |

**结论：**

> 对于中大型项目，**不要靠注释 YAML 控制服务**，
> 用 `profiles` 才是长期正确解。

