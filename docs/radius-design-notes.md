
# FreeRADIUS（SQL + Portal 场景）裁剪与测试设计要点

> 本文记录在 **Docker + FreeRADIUS 3.2 + MySQL** 架构下，
> 面向 **Portal / Controller 认证模型** 的 **裁剪原则、关键配置与验证路径**。

---

## 1. 设计目标与边界

### 1.1 目标

* 构建一套 **最小可用、可维护、可扩展** 的 RADIUS 服务
* 认证模型：

  * Portal / Controller 写 SQL
  * FreeRADIUS 仅做认证与策略返回
* 支持：

  * PAP 认证
  * SQL 后端（MySQL / MariaDB）
  * Docker 部署

### 1.2 明确不做的事情（本阶段）

* ❌ 不启用 802.1X / EAP
* ❌ 不启用 inner-tunnel
* ❌ 不使用静态 users 文件
* ❌ 不在容器运行时覆盖 raddb（volume）

---

## 2. raddb 裁剪设计原则

### 2.1 baseline 与 runtime 分离

| 目录                      | 用途                              |
| ----------------------- | ------------------------------- |
| `freeradius.base/raddb` | 官方完整配置（baseline，用于 diff / 升级参考） |
| `freeradius/raddb`      | 实际运行配置（runtime，裁剪版）             |

**baseline 只读，不直接挂载到运行容器。**

---

### 2.2 runtime raddb 最小结构

```text
raddb/
├── clients.conf
├── mods-enabled/
│   └── sql
└── sites-enabled/
    └── default
```

> 这是 **Portal + SQL** 场景下的最小闭环结构。

---

### 2.3 明确删除的模块 / 文件

| 模块                           | 原因                               |
| ---------------------------- | -------------------------------- |
| `mods-enabled/eap`           | 非 802.1X 场景，会触发 Auth-Type EAP 错误 |
| `sites-enabled/inner-tunnel` | EAP 专用                           |
| `sites-enabled/clients.conf` | 避免与 SQL NAS 冲突                   |
| `users`                      | 不使用静态认证                          |

---

## 3. SQL 模块关键配置点（`mods-enabled/sql`）

### 3.1 必须开启的能力

```conf
read_clients = yes
password_attribute = Cleartext-Password
```

#### 说明

* `read_clients = yes`

  * 启用 SQL `nas` 表作为 client / NAS 来源
  * 否则来自 SQL NAS 的请求会被**静默丢弃**

* `password_attribute = Cleartext-Password`

  * 明确告诉 PAP 使用 SQL 返回的明文密码
  * 避免出现 *“Expected Access-Accept got Access-Reject”*

---

## 4. 虚拟服务器设计（`sites-enabled/default`）

### 4.1 最终推荐版本

```conf
listen {
    type = auth
    ipaddr = *
    port = 1812
}

listen {
    type = acct
    ipaddr = *
    port = 1813
}

server default {

authorize {
    preprocess
    chap
    mschap

    sql

    expiration
    logintime
}

authenticate {
    Auth-Type PAP {
        pap
    }
}

post-auth {
    sql
}

accounting {
    sql
}

}
```

### 4.2 关键点说明

* `listen {}` 必须显式配置
  否则会报：

```text
The server is not configured to listen on any ports
```

* `sql` 是唯一认证来源
* 不包含任何 EAP / TLS / inner-tunnel 逻辑

---

## 5. NAS（client）设计与坑点

### 5.1 SQL NAS 表的真实作用

```sql
select * from nas;
```

| 字段        | 作用                        |
| --------- | ------------------------- |
| `nasname` | 必须与请求中的 NAS-IP-Address 匹配 |
| `secret`  | RADIUS 共享密钥               |
| `type`    | 仅标识用途                     |

### 5.2 常见误区

| 现象                      | 原因                  |
| ----------------------- | ------------------- |
| radtest 一直重发，无返回        | client 不合法（NAS 未匹配） |
| Access-Reject 而非 Accept | 密码属性未被 PAP 使用       |

> **client 不合法时，FreeRADIUS 会直接丢包，不返回 Reject。**

---

## 6. 用户认证数据模型（SQL）

### 6.1 最小可用用户

```sql
insert into radcheck (username, attribute, op, value)
values ('testuser', 'Cleartext-Password', ':=', 'testpass');
```

### 6.2 返回策略（示例）

```sql
insert into radreply (username, attribute, op, value)
values ('testuser', 'Session-Timeout', ':=', '3600');
```

---

## 7. 测试闭环设计（radtest）

### 7.1 推荐测试路径

```bash
radtest testuser testpass <radius_ip> 0 testing123
```

### 7.2 成功标志

```text
Received Access-Accept
Session-Timeout = 3600
```

### 7.3 调试模式（强烈推荐至少一次）

```bash
docker compose exec freeradius freeradius -X
```

可以清晰看到：

* SQL 查询过程
* client 匹配逻辑
* PAP 密码对比结果

---

## 8. Docker 架构关键决策

### 8.1 不使用 volume 覆盖 raddb

**原因：**

* volume 会引入：

  * 文件权限异常（globally writable）
  * 配置残留
  * 运行状态不可控

**改用：**

* Dockerfile `COPY raddb`
* build-time 冻结配置

---

### 8.2 推荐生命周期操作封装

* `Makefile`
* `justfile`
* 或脚本化：

```bash
stop → rm → build → up
```

避免“以为改了，其实跑的是旧容器”。

---

## 9. 一句话总结（设计哲学）

> **FreeRADIUS 是模块驱动系统**
>
> * 启用的模块必须与认证模型完全一致
> * SQL 模式必须显式声明“信任边界”（NAS）与“密码来源”
>
> **裁剪不是删除功能，而是删除不确定性。**

---

## 10. 当前系统能力评估

* ✅ SQL 用户认证
* ✅ SQL NAS 管理
* ✅ Docker 化部署
* ✅ Portal / Controller 可直接接入
* ✅ 可扩展 VLAN / Redirect / CoA

---

> 📌 建议在此状态打 Git Tag，例如：
>
> ```
> radius-sql-ok
> ```


好，这一步我帮你把 **“文件 / 脚本 / 目录路径”** 系统性补齐，直接作为 **上一份设计文档的补充章节**。
你可以**原样追加**到 `docs/radius-design-notes.md` 里。

---

## 11. 关键文件与脚本路径约定（Project Layout）

> 本节用于明确 **每一类配置 / 脚本 / 参考文件的“唯一归属路径”**，
> 防止后续维护中出现「不知道该改哪」「改了没生效」的问题。

---

### 11.1 Control Plane 顶层结构

```text
control-plane/
├── docker-compose.yml          # 服务编排入口（唯一）
├── Makefile                    # 工程级生命周期封装
├── justfile                    # 本地开发快捷封装（可选）
├── scripts/
│   └── restart-service.sh      # stop/rm/build/up 通用脚本
│
├── radius-stack/
│   ├── freeradius.base/        # 官方 raddb baseline（只读参考）
│   ├── freeradius/             # 实际运行的 FreeRADIUS 服务
│   │   ├── Dockerfile
│   │   ├── raddb/
│   │   │   ├── clients.conf
│   │   │   ├── mods-enabled/
│   │   │   │   └── sql
│   │   │   └── sites-enabled/
│   │   │       └── default
│   │   └── README.md
│   ├── mysql/                  # MySQL 初始化 / schema
│   └── supervisor/             # 可选：统一进程管理
│
└── docs/
    └── radius-design-notes.md  # 本设计文档
```

---

## 12. FreeRADIUS 相关路径说明（重点）

### 12.1 Dockerfile 路径（唯一生效点）

```text
radius-stack/freeradius/Dockerfile
```

职责：

* 基于官方镜像构建
* 安装 SQL 运行时依赖（mariadb-connector-c）
* COPY runtime raddb
* 裁剪无用模块
* 修正权限

> ⚠️ **任何 raddb 修改，都必须触发重新 build**

---

### 12.2 Runtime raddb 路径（真正生效）

```text
radius-stack/freeradius/raddb/
```

这是 **唯一会被 COPY 进容器并生效的配置目录**。

| 文件                      | 作用                    |
| ----------------------- | --------------------- |
| `clients.conf`          | 可为空；不使用静态 client      |
| `mods-enabled/sql`      | SQL / NAS / 密码来源核心配置  |
| `sites-enabled/default` | 虚拟服务器 + listen + 认证流程 |

---

### 12.3 Baseline raddb（只读参考）

```text
radius-stack/freeradius.base/raddb/
```

用途：

* 与 runtime raddb 做 diff
* 官方升级时对照
* 审计配置来源

**禁止：**

* ❌ Docker 运行时挂载
* ❌ 直接修改当作 runtime 用

---

## 13. SQL / MySQL 相关路径

### 13.1 MySQL 服务目录

```text
radius-stack/mysql/
```

通常包含：

* 初始化 SQL（schema）
* 数据卷定义（docker-compose.yml 中）

关键表：

* `nas`
* `radcheck`
* `radreply`
* `radacct`

---

### 13.2 SQL 表职责速查

| 表           | 作用                            |
| ----------- | ----------------------------- |
| `nas`       | NAS / client 白名单（RADIUS 信任边界） |
| `radcheck`  | 用户认证条件                        |
| `radreply`  | 认证成功后的返回策略                    |
| `radacct`   | 计费 / 在线会话                     |
| `operators` | 管理后台账号（非认证用户）                 |

---

## 14. 脚本与自动化路径

### 14.1 通用重启脚本（兜底）

```text
scripts/restart-service.sh
```

用法示例：

```bash
./scripts/restart-service.sh freeradius --no-cache
```

职责：

* stop → rm → build → up
* 适用于紧急排障或不走 Makefile 的场景

---

### 14.2 Makefile（推荐主入口）

```text
Makefile
```

常用命令：

```bash
make restart-freeradius
make restart-nc-freeradius
make logs-freeradius
make sh-freeradius
```

定位原则：

> **任何“对服务生命周期的操作”，优先写进 Makefile**

---

### 14.3 justfile（本地开发增强）

```text
justfile
```

示例：

```bash
just restart freeradius
just logs freeradius
```

说明：

* 不依赖 CI
* 提升本地开发体验
* 可选，但强烈推荐个人使用

---

## 15. 测试与调试路径

### 15.1 radtest（外部）

```bash
radtest testuser testpass <radius_ip> 0 testing123
```

通常在：

```text
control-plane/
```

目录下执行。

---

### 15.2 FreeRADIUS Debug（容器内）

```bash
docker compose exec freeradius freeradius -X
```

用途：

* 查看 SQL 查询
* 查看 PAP / Auth-Type 判定
* 排查 Reject / Drop 原因

---

## 16. 修改生效规则（非常重要）

> **只记住这一条即可：**

### ❗ 任何涉及以下路径的修改：

```text
radius-stack/freeradius/raddb/*
radius-stack/freeradius/Dockerfile
```

### 必须执行：

```bash
make restart-nc-freeradius
```

否则你看到的行为**可能仍然来自旧镜像**。

---

## 17. 路径设计原则总结

* **runtime 与 baseline 强隔离**
* **build-time 冻结配置，避免 volume 漂移**
* **生命周期操作统一入口（Makefile / just）**
* **SQL 是状态源，FreeRADIUS 是执行引擎**

---

