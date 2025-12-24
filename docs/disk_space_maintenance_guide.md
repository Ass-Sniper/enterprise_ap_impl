

# 虚拟机磁盘空间管理与 Confluence 运维手册

**文档说明**：本手册用于记录在 Ubuntu 环境下运行 Docker 版 Confluence 时遇到的磁盘空间危机处理过程及常用清理命令。
**适用环境**：Ubuntu 20.04/22.04+, Docker, Confluence 8.5 LTS, PostgreSQL 14

---

## 1. 空间排查命令 (Inspection)

当 Confluence 出现 “Free Disk Space Health Check Failed” 警告时，应立即执行以下命令定位问题。

### 1.1 分区占用概览

```bash
# 查看磁盘挂载点及剩余空间
df -h /

```

### 1.2 目录深度分析

```bash
# 查找根目录下占用最大的前 10 个目录
sudo du -sh /* 2>/dev/null | sort -hr | head -n 10

# 查找用户目录下的大文件 (重点检查 codebase)
sudo du -sh /home/kay/* 2>/dev/null | sort -hr | head -n 10

```

---

## 2. Docker 专项清理 (Docker Cleanup)

Docker 运行久了会产生大量的冗余数据。

| 命令 | 作用 | 风险等级 |
| --- | --- | --- |
| `docker system df` | 查看 Docker 资源占用汇总 | 低 |
| `docker system prune` | 删除停止的容器、无用网络、悬空镜像 | 中 |
| `docker system prune -a` | 删除所有未被使用的镜像 | 高 (需重新下载) |
| `docker volume prune` | **清理孤儿数据卷 (空间占用大户)** | 中 |

---

## 3. 系统瘦身与日志管理 (System Maintenance)

### 3.1 APT 缓存清理

```bash
# 清理下载的软件包缓存
sudo apt-get clean
# 卸载不再需要的依赖包
sudo apt-get autoremove --purge

```

### 3.2 系统日志 (Journalctl) 限制

防止系统日志无限增长：

```bash
# 限制日志仅保留最近 500MB
sudo journalctl --vacuum-size=500M
# 或者保留最近 3 天
sudo journalctl --vacuum-time=3d

```

### 3.3 Snap 包清理

Ubuntu 默认会保留旧版本的 Snap 软件包，执行以下脚本释放空间：

```bash
set -eu
snap list --all | awk '/disabled/{print $1, $3}' |
    while read snapname revision; do
        snap remove "$snapname" --revision="$revision"
    done

```

---

## 4. 开发环境专项清理 (Codebase/Build)

针对 OpenWrt / rm65 等项目的编译清理：

```bash
# 进入源码目录执行
make clean      # 清理编译产物
make dirclean  # 深度清理编译环境 (含工具链)

# 查找目录下超过 500MB 的大文件
find /home/kay/codebase -type f -size +500M

```

---

## 5. Confluence 运维建议

1. **红线预警**：根分区 `Avail` 必须保持在 **5GB** 以上，建议预留 **15GB**。
2. **健康检查更新**：清理磁盘后，若警告未消失，可尝试 `docker restart confluence` 重启触发自检。
3. **End of Life 警告**：对于 8.5 LTS 版本，此警告为官方生命周期提醒，不影响个人/本地开发使用，可安全忽略。

---

**归档日期**：2025年12月
**状态**：已修复 (磁盘占用从 98% 降至 86%)

---

