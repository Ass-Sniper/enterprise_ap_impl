
# MT7981 内核调试归档：KGDB 与 ftrace 配置指南

## 1. 内核编译配置 (make kernel_menuconfig)

为了支持深层调试，内核必须开启以下选项。

### A. KGDB 核心配置

* `Kernel hacking`
* `Generic Kernel Debugging Instruments`
* `KGDB: kernel debugger` —— **[*]**
* `KGDB: use kgdb over the serial console` (KGDBOC) —— **[*]**
* `KGDB: internal test suite` —— **[ ]** (可选)
* `KGDB_KDB: include kdb frontend for kgdb` —— **[*]** (提供交互式 Shell)







### B. ftrace 追踪配置

* `Kernel hacking`
* `Tracers`
* `Kernel Function Tracer` (FUNCTION_TRACER) —— **[*]**
* `Kernel Function Graph Tracer` (FUNCTION_GRAPH_TRACER) —— **[*]**
* `Enable/disable function tracing dynamically` (DYNAMIC_FTRACE) —— **[*]**
* `Trace max stack usage` (STACK_TRACER) —— **[*]**





### C. 移除重置干扰 (重点)

* `Device Drivers`
* `Watchdog Timer Support`
* `Mediatek SoCs watchdog support` —— **[N]** (排除，防止调试时系统自动重启)

```text
 .config - Linux/arm64 5.4.284 Kernel Configuration
 > Search (watchdog) > Device Drivers > Watchdog Timer Support ────────────────────────────────────────────────────────────────────────────
  ┌────────────────────────────────────────────────────── Watchdog Timer Support ───────────────────────────────────────────────────────┐
  │  Arrow keys navigate the menu.  <Enter> selects submenus ---> (or empty submenus ----).  Highlighted letters are hotkeys.  Pressing │
  │  <Y> includes, <N> excludes, <M> modularizes features.  Press <Esc><Esc> to exit, <?> for Help, </> for Search.  Legend: [*]        │
  │  built-in  [ ] excluded  <M> module  < > module capable                                                                             │
  │                                                                                                                                     │
  │ ┌─────────────────────────────^(-)────────────────────────────────────────────────────────────────────────────────────────────────┐ │
  │ │                             -*-   WatchDog Timer Driver Core                                                                    │ │
  │ │                             [ ]   Disable watchdog shutdown on close                                                            │ │
  │ │                             [*]   Update boot-enabled watchdog until userspace takes over                                       │ │
  │ │                             (0)   Timeout value for opening watchdog device                                                     │ │
  │ │                             [*]   Read different watchdog information through sysfs                                             │ │
  │ │                                   *** Watchdog Pretimeout Governors ***                                                         │ │
  │ │                             [*]   Enable watchdog pretimeout governors                                                          │ │
  │ │                             < >     Noop watchdog pretimeout governor                                                           │ │
  │ │                             {*}     Panic watchdog pretimeout governor                                                          │ │
  │ │                                     Default Watchdog Pretimeout Governor (panic)  --->                                          │ │
  │ │                                   *** Watchdog Device Drivers ***                                                               │ │
  │ │                             < >   Software watchdog                                                                             │ │
  │ │                             < >   Watchdog device controlled through GPIO-line                                                  │ │
  │ │                             < >   Xilinx Watchdog timer                                                                         │ │
  │ │                             < >   Zodiac RAVE Watchdog Timer                                                                    │ │
  │ │                             < >   ARM SP805 Watchdog                                                                            │ │
  │ │                             < >   ARM SBSA Generic Watchdog                                                                     │ │
  │ │                             < >   Cadence Watchdog Timer                                                                        │ │
  │ │                             < >   Synopsys DesignWare watchdog                                                                  │ │
  │ │                             < >   Max63xx watchdog                                                                              │ │
  │ │                             <*>   Mediatek SoCs watchdog support                                                                │ │
  │ │                             < >   ALi M7101 PMU Computer Watchdog                                                               │ │
  │ │                             < >   Intel 6300ESB Timer/Watchdog                                                                  │ │
  │ │                             < >   MEN A21 VME CPU Carrier Board Watchdog Timer                                                  │ │
  │ │                                   *** PCI-based Watchdog Cards ***                                                              │ │
  │ └─────────────────────────────v(+)────────────────────────────────────────────────────────────────────────────────────────────────┘ │
  ├─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
  │                                      <Select>    < Exit >    < Help >    < Save >    < Load >                                       │
  └─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

```

这是一个非常重要的补充。没有 **Debug Info**，GDB 就像“盲人摸象”——它能连接上内核，但找不到函数名，看不到变量，更无法对齐源代码。

你可以将以下内容添加到归档文档的 **“1. 内核编译配置”** 章节中：

---

### D. 内核符号与调试信息 (Essential for GDB)

要让 GDB 能够将内存地址映射到源代码，必须开启内核调试符号。

* `Kernel hacking`
* `Compile-time checks and compiler options`
* `Compile the kernel with debug info` (DEBUG_INFO) —— **[*]**
* `Reduce debugging information` (DEBUG_INFO_REDUCED) —— **[ ]** (确保**取消勾选**，否则调试信息不全)
* `Generate dwarf4 debuginfo` (DEBUG_INFO_DWARF4) —— **[*]** (如果你的 GDB 版本较老，选这个；通常选默认的 DWARF 格式即可)
* `Provide GDB scripts for kernel debugging` (GDB_SCRIPTS) —— **[*]** (提供非常有用的 `lx-*` 辅助命令)





---

### 📂 补充后的完整配置清单（供拷贝）

为了方便以后在 `.config` 中直接搜索确认，可以对照以下符号名：

```bash
# Debug Information
CONFIG_DEBUG_INFO=y
# CONFIG_DEBUG_INFO_REDUCED is not set
CONFIG_GDB_SCRIPTS=y

# KGDB Base
CONFIG_KGDB=y
CONFIG_KGDB_SERIAL_CONSOLE=y
CONFIG_KGDB_KDB=y
CONFIG_HAVE_ARCH_KGDB=y

# Ftrace Base
CONFIG_FUNCTION_TRACER=y
CONFIG_FUNCTION_GRAPH_TRACER=y
CONFIG_DYNAMIC_FTRACE=y

# Disable Interference
# CONFIG_MTK_WDT is not set

```

```text
root@kay-vm:immortalwrt-mt798x# grep -E "CONFIG_KGDB|CONFIG_FTRACE|CONFIG_DEBUG_INFO|CONFIG_FUNCTION_TRACER|WATCHDOG" build_dir/ta
rget-aarch64_cortex-a53_musl/linux-mediatek_mt7981/linux-5.4.284/.config
CONFIG_WATCHDOG=y
CONFIG_WATCHDOG_CORE=y
# CONFIG_WATCHDOG_NOWAYOUT is not set
CONFIG_WATCHDOG_HANDLE_BOOT_ENABLED=y
CONFIG_WATCHDOG_OPEN_TIMEOUT=0
CONFIG_WATCHDOG_SYSFS=y
CONFIG_WATCHDOG_PRETIMEOUT_GOV=y
CONFIG_WATCHDOG_PRETIMEOUT_GOV_SEL=m
# CONFIG_WATCHDOG_PRETIMEOUT_GOV_NOOP is not set
CONFIG_WATCHDOG_PRETIMEOUT_GOV_PANIC=y
CONFIG_WATCHDOG_PRETIMEOUT_DEFAULT_GOV_PANIC=y
# CONFIG_SOFT_WATCHDOG is not set
# CONFIG_GPIO_WATCHDOG is not set
# CONFIG_XILINX_WATCHDOG is not set
# CONFIG_ZIIRAVE_WATCHDOG is not set
# CONFIG_ARM_SP805_WATCHDOG is not set
# CONFIG_ARM_SBSA_WATCHDOG is not set
# CONFIG_CADENCE_WATCHDOG is not set
# CONFIG_DW_WATCHDOG is not set
# CONFIG_MAX63XX_WATCHDOG is not set
# CONFIG_MEDIATEK_WATCHDOG is not set
# CONFIG_PCIPCWATCHDOG is not set
# CONFIG_USBPCWATCHDOG is not set
CONFIG_DEBUG_INFO=y
# CONFIG_DEBUG_INFO_REDUCED is not set
# CONFIG_DEBUG_INFO_SPLIT is not set
# CONFIG_DEBUG_INFO_DWARF4 is not set
# CONFIG_DEBUG_INFO_BTF is not set
# CONFIG_WQ_WATCHDOG is not set
CONFIG_FTRACE=y
CONFIG_FUNCTION_TRACER=y
# CONFIG_FTRACE_SYSCALLS is not set
CONFIG_FTRACE_MCOUNT_RECORD=y
# CONFIG_FTRACE_STARTUP_TEST is not set
CONFIG_KGDB=y
CONFIG_KGDB_SERIAL_CONSOLE=y
# CONFIG_KGDB_TESTS is not set
CONFIG_KGDB_KDB=y
root@kay-vm:immortalwrt-mt798x#
```

---

### 💡 为什么 `lx-scripts` 如此重要？

当你开启了 `CONFIG_GDB_SCRIPTS`，在编译完成后，内核目录会生成一个 `vmlinux-gdb.py`。在 Ubuntu 上启动 GDB 时，它会自动加载一些增强命令。比如：

* **`lx-dmesg`**: 在 GDB 里直接查看内核日志（不需要退出到终端）。
* **`lx-lsmod`**: 查看当前加载的内核模块。
* **`lx-ps`**: 列出当前系统运行的所有进程及其 PCB 地址。

---

## 2. 启动参数 (Bootargs) 配置

在 U-Boot 中设置环境变量，确保内核在初始化串口后立即进入等待状态。

```bash
setenv bootargs "console=ttyS0,115200n1 loglevel=8 earlycon=uart8250,mmio32,0x11002000 panic=0 nowatchdog ignore_loglevel kgdboc=ttyS0,115200 kgdbwait"

```

* **kgdboc=ttyS0,115200**: 将 KGDB 绑定到第一个串口。
* **kgdbwait**: 告知内核在启动早期挂起，等待 GDB 连接。
* **panic=0**: 发生内核错误时不重启。

---

## 3. 运行中触发 KGDB (运行时)

如果系统已经启动，可以通过以下方式手动进入调试模式：

```bash
# 1. 临时绑定串口（如果启动参数未起效）
echo ttyS0 > /sys/module/kgdboc/parameters/kgdboc

# 2. 停止系统看门狗（针对 OpenWrt/procd）
ubus call system watchdog '{"stop": true}'

# 3. 激活断点
echo g > /proc/sysrq-trigger

```

---

## 4. 远程 GDB 连接流程 (Ubuntu 端)

### A. 准备工作

确保你拥有带符号表的内核文件 `vmlinux`（位于编译目录的 `build_dir/target-.../linux-.../vmlinux`）。

### B. 连接指令

```bash
# 启动 GDB
gdb-multiarch vmlinux

# GDB 内部执行
(gdb) set architecture aarch64
(gdb) set remotebaud 115200
(gdb) target remote /dev/ttyUSB0

```

---

## 5. ftrace 常用调试指令

连上系统后，通过 DebugFS 进行追踪分析：

```bash
mount -t debugfs nodev /sys/kernel/debug
cd /sys/kernel/debug/tracing

# 1. 设置追踪函数 (例如网络包接收)
echo ip_rcv > set_ftrace_filter

# 2. 开启函数调用图
echo function_graph > current_tracer

# 3. 开启追踪
echo 1 > tracing_on

# 4. 查看结果
cat trace | less

```

---

## 6. 常见陷阱与对策

| 问题 | 原因 | 对策 |
| --- | --- | --- |
| **系统自动重启** | 硬件看门狗超时 | 在内核配置中禁用 `MTK_WDT` 或在 U-Boot 中关狗。 |
| **GDB 连接超时** | 串口被占用 | 确保关闭了 minicom/putty 等工具。 |
| **KGDB 无响应** | 串口驱动未就绪 | 去掉 `earlycon` 启动参数尝试。 |
| **找不到符号** | vmlinux 不匹配 | 确保 GDB 使用的 vmlinux 与运行中的内核是同一次生成的。 |

---

