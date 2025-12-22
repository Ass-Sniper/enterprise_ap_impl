
# Redis 启动、初始化与确认流程

本项目对 Redis 采用 **自定义 Docker 镜像 + entrypoint** 的方式，
确保 `cp-redis` 容器在启动时 **自动执行初始化脚本**，行为可控、可观测、可复现。

---

## 一、停止与重新启动 Redis（重要）

当修改了以下任一内容时，**必须重建并重启 Redis 容器**：

* `redis/Dockerfile`
* `redis/entrypoint.sh`
* `docker-compose.yml` 中 Redis 的 `build` 配置

### 1️⃣ 停止 Redis 容器

```bash
docker compose down redis
```

确认容器已停止：

```bash
docker ps | grep cp-redis
```

无输出即表示已停止。

---

### 2️⃣ 重新构建并启动 Redis

```bash
docker compose up -d --build redis
```

> ⚠️ 必须带 `--build`
> 否则 Docker 可能继续使用旧镜像，导致 entrypoint 不生效。

---

## 二、Redis 容器内部启动与初始化流程

执行以下命令启动 Redis：

```bash
docker compose up redis
```

容器内部实际执行流程如下：

```text
docker compose up redis
        ↓
ENTRYPOINT ["/entrypoint.sh"]
        ↓
entrypoint.sh 开始执行
        ↓
启动 redis-server（后台运行）
        ↓
循环等待 Redis 就绪（redis-cli ping）
        ↓
执行初始化脚本 00-portal-redis-schema.sh
        ↓
阻塞等待 redis-server 进程（容器保持运行）
```

---

## 三、启动日志确认（关键步骤）

Redis 启动完成后，应检查日志以确认 **entrypoint 与初始化脚本均已执行**。

```bash
docker logs cp-redis
```

### 期望日志示例

```text
[redis-entrypoint] starting redis...
[redis-entrypoint] waiting for redis...
Redis is starting oO0OoO0OoO0Oo
Ready to accept connections tcp
[redis-entrypoint] redis is ready
[redis-entrypoint] running portal redis bootstrap
[portal-init] Initializing Redis keys...
OK
OK
OK
OK
OK
[portal-init] Redis initialization done.
```

#### 日志含义说明

| 日志行                                       | 说明                  |
| ----------------------------------------- | ------------------- |
| `[redis-entrypoint] starting redis...`    | 使用自定义 entrypoint    |
| `[redis-entrypoint] waiting for redis...` | 等待 Redis 就绪，避免竞态    |
| `Ready to accept connections`             | Redis 已监听端口         |
| `[redis-entrypoint] redis is ready`       | `redis-cli ping` 成功 |
| `running portal redis bootstrap`          | 初始化脚本被触发            |
| `Redis initialization done`               | Key 写入完成            |

---

## 四、确认 Redis 使用的是自定义 entrypoint（必查）

```bash
docker inspect cp-redis | grep -i entrypoint
```

### 正确输出示例

```text
"Entrypoint": [
    "/entrypoint.sh"
]
```

如果看到的是：

```text
"docker-entrypoint.sh"
```

说明仍在使用官方 Redis 镜像，需要重新 `down + build + up`。

---

## 五、确认 Redis Key 是否已初始化

进入 Redis CLI：

```bash
docker exec -it cp-redis redis-cli
```

查看 Key：

```redis
KEYS *
```

示例输出：

```text
feature:portal
feature:portal:redirect
feature:portal:walled_garden
feature:auth:pap
feature:auth:radius
auth:strategy:pap
auth:strategy:radius
portal:hmac:key
policy:global:rev
```

> 看到上述 Key，说明初始化脚本已成功执行。

---

## 六、常见操作场景与对应行为

| 操作                            | 是否执行 entrypoint | 是否执行初始化       |
| ----------------------------- | --------------- | ------------- |
| `docker compose up redis`     | ✅               | ✅             |
| `docker restart cp-redis`     | ✅               | ✅（除非脚本内 skip） |
| `docker exec -it cp-redis sh` | ❌               | ❌             |
| 修改 init 脚本内容                  | ❌               | ❌（需重启容器）      |
| `down + up --build`           | ✅               | ✅             |

---

## 七、注意事项

* 初始化脚本 **不是 Redis 官方机制**，而是通过自定义 entrypoint 实现
* entrypoint **每次容器启动都会执行**
* 建议在初始化脚本中加入「已初始化判断」，避免重复覆盖数据

示例：

```sh
redis-cli EXISTS feature:portal | grep -q 1 && {
  echo "[portal-init] already initialized, skip"
  exit 0
}
```

---

## 八、总结

* Redis 启动流程完全可控
* 初始化行为可观察、可复现
* 当前实现适用于 **开发 / 测试环境**
* 后续可平滑演进为 config-plane / 管理页面方案


## 附：操作日志

```text
kay@kay-vm:control-plane$ docker compose down redis
[+] Running 2/2
 ✔ Container cp-redis                       Removed                                                          0.2s
 ! Network control-plane_control-plane-net  Resource is still i...                                           0.0s
kay@kay-vm:control-plane$
kay@kay-vm:control-plane$ docker compose up -d --build redis
[+] Building 0.3s (9/9) FINISHED
 => [internal] load local bake definitions                                                                   0.0s
 => => reading from stdin 420B                                                                               0.0s
 => [internal] load build definition from Dockerfile                                                         0.0s
 => => transferring dockerfile: 135B                                                                         0.0s
 => [internal] load metadata for docker.io/library/redis:7-alpine                                            0.0s
 => [internal] load .dockerignore                                                                            0.0s
 => => transferring context: 2B                                                                              0.0s
 => [internal] load build context                                                                            0.0s
 => => transferring context: 602B                                                                            0.0s
 => [1/2] FROM docker.io/library/redis:7-alpine                                                              0.0s
 => [2/2] COPY --chmod=755 entrypoint.sh /entrypoint.sh                                                      0.0s
 => exporting to image                                                                                       0.0s
 => => exporting layers                                                                                      0.0s
 => => writing image sha256:6fdcd948b3a063668ec5c7ce5c4a22d89ec7763b057f9407b236eaaff6906e53                 0.0s
 => => naming to docker.io/library/control-plane-redis                                                       0.0s
 => resolving provenance for metadata file                                                                   0.0s
[+] Running 2/2
 ✔ redis               Built                                                                                 0.0s
 ✔ Container cp-redis  Started                                                                               0.3s
kay@kay-vm:control-plane$
kay@kay-vm:control-plane$ docker inspect cp-redis | grep -i entrypoint
        "Path": "/entrypoint.sh",
                "/home/kay/codebase/enterprise_ap_impl/control-plane/redis/docker-entrypoint-initdb.d:/docker-entrypoint-initdb.d:ro"
                "Source": "/home/kay/codebase/enterprise_ap_impl/control-plane/redis/docker-entrypoint-initdb.d",
                "Destination": "/docker-entrypoint-initdb.d",
            "Entrypoint": [
                "/entrypoint.sh"
kay@kay-vm:control-plane$
kay@kay-vm:control-plane$
kay@kay-vm:control-plane$ docker logs cp-redis
[redis-entrypoint] starting redis...
[redis-entrypoint] waiting for redis...
7:C 22 Dec 2025 12:50:40.255 # WARNING Memory overcommit must be enabled! Without it, a background save or replication may fail under low memory condition. Being disabled, it can also cause failures without low memory condition, see https://github.com/jemalloc/jemalloc/issues/1328. To fix this issue add 'vm.overcommit_memory = 1' to /etc/sysctl.conf and then reboot or run the command 'sysctl vm.overcommit_memory=1' for this to take effect.
7:C 22 Dec 2025 12:50:40.255 * oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
7:C 22 Dec 2025 12:50:40.255 * Redis version=7.4.0, bits=64, commit=00000000, modified=0, pid=7, just started
7:C 22 Dec 2025 12:50:40.255 # Warning: no config file specified, using the default config. In order to specify a config file use redis-server /path/to/redis.conf
7:M 22 Dec 2025 12:50:40.255 * monotonic clock: POSIX clock_gettime
7:M 22 Dec 2025 12:50:40.256 * Running mode=standalone, port=6379.
7:M 22 Dec 2025 12:50:40.257 * Server initialized
7:M 22 Dec 2025 12:50:40.257 * Ready to accept connections tcp
[redis-entrypoint] redis is ready
[redis-entrypoint] running portal redis bootstrap
[portal-init] Initializing Redis keys...
OK
OK
OK
OK
OK
OK
OK
OK
OK
OK
OK
[portal-init] Redis initialization done.
kay@kay-vm:control-plane$
```

---