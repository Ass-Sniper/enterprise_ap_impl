
# Lazygit 全能部署与使用手册

**文档说明**：本手册记录了在 Ubuntu 虚拟机环境下安装 Lazygit 的多种方式（含代理加速与离线方案），以及高效交互的操作指南。

---

## 1. 安装方式 (Installation)

根据网络环境选择最合适的安装方式。

### 方案 A：PPA 官方源安装 (推荐)

适用于网络良好或有代理的环境。

```bash
# 1. 添加 PPA 存储库
sudo add-apt-repository ppa:lazygit-team/release
sudo apt update

# 2. 使用代理安装 (防止下载缓慢)
# -E 参数确保 sudo 继承当前用户的 http_proxy 环境变量
sudo -E apt install lazygit

```

### 方案 B：GitHub 离线安装 (最快/最新)

如果 PPA 下载极慢，直接下载预编译二进制文件。

```bash
# 1. 从 GitHub 下载压缩包 (以 v0.44.1 为例)
wget https://github.com/jesseduffield/lazygit/releases/download/v0.44.1/lazygit_0.44.1_Linux_x86_64.tar.gz

# 2. 解压
tar -zxvf lazygit_0.44.1_Linux_x86_64.tar.gz

# 3. 安装到系统路径
sudo install lazygit /usr/local/bin/lazygit

# 4. 验证
lazygit --version

```

---

## 2. 界面布局与导航 (Navigation)

### 2.1 五大核心面板

通过数字键 **1 - 5** 快速切换焦点：

1. **Status**: 状态栏，查看分支及冲突。
2. **Files**: 文件栏，相当于 `git status` + `git add`。
3. **Branches**: 分支栏，切换、创建、合并分支。
4. **Commits**: 历史栏，**查看 Log、合并提交、重置代码**。
5. **Stash**: 暂存栏，管理隐藏进度。

### 2.2 左右窗口联动

* **查看详情**：在左侧移动光标，右侧大窗口自动显示 Diff。
* **进入右侧**：选中项目按 **`Enter`**，光标跳入右侧（可进行滚动或搜索）。
* **返回左侧**：按 **`Esc`** 或 **`h`**。
* **搜索内容**：在右侧窗口按 **`/`** 输入关键词。

---

## 3. 核心快捷键 (Hotkeys)

| 面板 | 键位 | 动作 | 说明 |
| --- | --- | --- | --- |
| **通用** | `x` | 帮助菜单 | 忘记快捷键时随时按 `x` 查看当前可用操作 |
| **Files** | `Space` | 暂存 (Stage) | 相当于 `git add <file>` |
| **Files** | `c` | 提交 (Commit) | 弹出窗口输入提交信息 |
| **Files** | `d` | 丢弃修改 | **危险操作**：彻底删除该文件未提交的内容 |
| **Commits** | `s` | 合并 (Squash) | 将选中的 Commit 与前一个合并 |
| **Commits** | `r` | 修改信息 | 重新编辑 Commit Message (Reword) |
| **Commits** | `g` | 强推/重置 | 重置当前分支到此位置 (Reset) |

---

## 4. 进阶技巧

### 4.1 局部代码行提交

1. 在 **Files** 面板选中文件，按 **`Enter`**。
2. 使用上下键选中特定代码行。
3. 按 **`Space`** 只暂存这几行代码，而非整个文件。

### 4.2 代理维护建议 (针对 sudo)

如果在安装或 Push/Pull 时遇到网络问题，请记住：

* **临时代理**：`sudo -E apt install ...`
* **持久化代理 (APT)**：
```bash
echo 'Acquire::http::Proxy "http://127.0.0.1:7897";' | sudo tee /etc/apt/apt.conf.d/99proxy

```



---

## 5. 归档总结

通过 Lazygit，你可以更直观地管理 `enterprise_ap_impl` 项目。由于你之前将文档和脚本分成了两次提交，现在你可以进入 **Commits 面板 (4)**，选中这两条记录，按 **`s`** 键体验一键合并的快感。

---

**归档日期**：2025年12月


