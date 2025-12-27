
# Linux Kernel Magic SysRq 全方位技术指南

## 1. 出现背景：内核绝望时的“紧急按钮”

在 Linux 开发早期，系统死锁（Hang）是开发者的噩梦。当内核挂起时，屏幕不动、键盘无响应，开发者只能强行断电。这会导致：

* **调试信息丢失**：内存中关键的上下文（Panic log）随掉电消失，无法分析死因。
* **文件系统损坏**：磁盘缓冲区（Buffer Cache）数据未落盘，导致重启后磁盘需要 fsck 甚至无法引导。

**Magic SysRq Key** 诞生于此。它的设计初衷是：**只要内核的底层中断处理机制还在工作，就提供一种绕过所有用户态软件，直接与内核通信的机制。** 它就像是内核里的一个“超级管理员哨位”。

## 2. 工作原理：从硬件中断到内核处理

SysRq 的触发分为硬件和软件两条路径，它们最终都会汇聚到内核的分发中心。

### A. 物理按键路径 (Hardware Path)

1. **硬件中断**：用户按下 `Alt + SysRq + <key>`。
2. **驱动过滤**：内核键盘驱动识别到特定序列，不将键值传给 `bash` 或 `GUI`，而是直接调用 `handle_sysrq`。
3. **查表分发**：内核查找全局表格 `sysrq_key_op`，根据字符（如 'g', 'b'）调用对应的内核处理函数。

### B. 软件模拟路径 (Software Path - 嵌入式调试常用)

在没有物理键盘的嵌入式设备（如 MT7981）上，通过 `/proc` 模拟触发：

1. **写入 `/proc**`：执行 `echo g > /proc/sysrq-trigger`。
2. **VFS 映射**：`ProcFS` 驱动捕获写操作，直接调用内核内部的 `__handle_sysrq()`。

## 3. 核心应用场景：救命与调试

### A. 调试与排错 (Debug Flags)

| 字符 | 功能 | 描述 |
| --- | --- | --- |
| **`g`** | **进入 KGDB** | **（本实验核心）** 挂起内核，将 CPU 控制权交给远程 GDB。 |
| **`t`** | **Task Dump** | 打印当前所有进程的调用栈，排查进程卡死。 |
| **`m`** | **Mem Info** | 打印当前内存分配状态，查看是否有内存泄漏或 OOM。 |
| **`p`** | **Reg Dump** | 打印当前 CPU 寄存器和标志位，确认指令执行位置。 |

### B. 安全重启序列 (The REISUB Sequence)

当系统桌面或网络完全失去响应时，依次输入以下指令可实现优雅重启：

1. **R** (Un**R**aw): 从 X11 手中取回键盘控制权。
2. **E** (t**E**rm): 给所有进程发 `SIGTERM` 信号。
3. **I** (k**I**ll): 给所有进程发 `SIGKILL` 信号。
4. **S** (**S**ync): **（关键）** 将内存数据同步刷入磁盘。
5. **U** (**U**nmount): 将所有文件系统重新挂载为只读模式。
6. **B** (**B**oot): **（最终）** 立即重启系统。

## 4. 安全性控制

通过 `/proc/sys/kernel/sysrq` 调整权限：

* `0`: 完全禁用。
* `1`: 启用所有功能。
* `掩码`: 如 `2` 允许记录控制，`4` 允许键盘控制等。

## 5. 深度实验：GDB 里的“上帝视角”

通过 KGDB 连通后，可以清晰地观察到 SysRq 的拦截过程：

1. **设置断点**：在 GDB 中执行 `break sysrq_handle_reboot`。
2. **执行指令**：在路由器终端执行 `echo b > /proc/sysrq-trigger`。
3. **现象**：系统**不会**立即重启。
4. **结果**：GDB 会在内核执行重置指令前精准拦截。此时，通过 `bt` 可以回溯整个调用链路：
`el0_svc` -> `vfs_write` -> `proc_reg_write` -> `write_sysrq_trigger` -> `handle_sysrq` -> `sysrq_handle_reboot`。

已经为你将这个非常典型的**“设备未就绪/消失”**故障排查过程整合进了文档。这在虚拟机环境下（VMware/VirtualBox）配合串口调试时极其常见。

建议将这段内容添加到 **“4. 远程 GDB 连接流程”** 之后，作为 **“故障排查 (Troubleshooting)”** 章节。

---

## 6. 常见连接故障排查：Device not found

在执行 `target remote /dev/ttyUSB0` 时，如果遇到以下错误：

```text
(gdb) target remote /dev/ttyUSB0
/dev/ttyUSB0: No such file or directory.

```

### 🛠️ 故障排查逻辑图

#### A. 物理与虚拟化层检查 (Physical & VM Layer)

1. **物理状态**：检查串口线（TTL-USB转接器）是否松动。
2. **USB 挂载**：确认 USB 设备已从宿主机（Windows/Mac）“断开”并“连接”到了虚拟机（Ubuntu）。
* *VMware*: `虚拟机 -> 可移动设备 -> [你的串口芯片] -> 连接(断开与主机的连接)`。
* *VirtualBox*: `设备 -> USB -> 勾选对应的串口设备`。



#### B. 系统层识别检查 (OS Layer)

在 Ubuntu 终端（非 GDB 内部）执行：

```bash
ls /dev/ttyUSB*

```

* **结果为空**：说明驱动未加载或硬件未挂载。请重新插拔 USB。
* **显示 /dev/ttyUSB1**：说明设备号变动了。在 GDB 中应改用 `target remote /dev/ttyUSB1`。

#### C. 权限与冲突检查 (Permissions & Conflict)

1. **读写权限**：即使设备存在，GDB 可能因为没有 `root` 权限而报错。
```bash
sudo chmod 666 /dev/ttyUSB0

```


2. **串口占用**：**（最重要）** 确认 `minicom`、`picocom` 或 `screen` 等终端软件已完全关闭。
* *排查指令*：`ps aux | grep -i minicom`
* *原因*：串口设备是排他性的，一旦被其他程序占用，GDB 握手协议会立即失败。



#### D. GDB 命令细节

如果在 GDB 内部由于操作失误提示：

```text
(gdb) set remotebaud 115200
No symbol "remotebaud" in current context.

```

**修正方法**：在较新版本的 GDB 中，请使用 `set serial baud 115200` 或在 `target` 命令后设置属性。

---

## 7. 串口争抢困境与“隔空取物”调试法

### 7.1. 核心矛盾：串口独占性 (Serial Port Constraints)

在嵌入式调试中，串口是一个“排他性”资源。当你进入 KGDB 会话时，会遇到以下经典冲突：

* **内核侧**：将串口作为二进制数据通道（GDB Remote Protocol），用于传输调试指令。
* **用户侧**：习惯通过串口终端（minicom/Putty）发送控制台命令。
* **结果**：一旦 GDB 接管了 `/dev/ttyUSB0`，任何试图打开该串口的终端软件都会提示“拒绝访问”，或者导致 GDB 通讯乱码。

### 7.2. 故障现象：GDB 执行命令“卡住”

**实验场景**：在 GDB 中执行 `continue` 让内核跑起来后，紧接着输入 `print` 命令。
**现象**：

```gdb
(gdb) continue
Continuing.
(gdb) print sysrq_key_table['m']
(此处无响应，光标闪烁...)

```

**原因分析**：

* 当内核处于 **Running**（继续运行）状态时，它拥有 CPU 的绝对控制权。
* GDB 此时只是一个监听者。由于内核没有停下来，它无法响应 GDB 的内存读取请求。
* **误区**：这并不是死机，而是 GDB 在等待内核命中断点。

### 7.3. 解决方案：多维触发机制

#### 方案 A：GDB 注入（GDB Injection）

在断点命中的状态下，直接利用 GDB 的权限修改内核内存或调用函数。

* **查看处理函数**：`print sysrq_key_table['m']`
* **强行调用**：`call sysrq_handle_showmem(0)`

> *注：这种方式直接在内核上下文中执行函数，无需通过 `/proc` 接口。*

#### 方案 B：网络辅助触发（SSH Trigger）

当串口被 GDB 占用时，利用网络通道（SSH/Web）作为“第二战场”发送指令。

1. **环境清理**：若重刷固件导致 SSH 报错，需执行 `ssh-keygen -R 192.168.16.254`。
2. **远程命令**：
```bash
ssh root@192.168.16.254 "echo m > /proc/sysrq-trigger"

```


3. **联动效果**：执行后，GDB 窗口会立刻捕获到 `__handle_sysrq` 断点，并自动处理之前积压的 `print` 命令。

### 7.4. 源码观察：`__handle_sysrq` 内部逻辑

通过 `list` 命令观察 `drivers/tty/sysrq.c`，我们可以看到内核处理 SysRq 的防御性代码：

```c
544    orig_log_level = console_loglevel;
545    console_loglevel = CONSOLE_LOGLEVEL_DEFAULT; // 强制提升日志等级
546
547    op_p = __sysrq_get_key_op(key); // 从上帝表格中查找对应的处理函数

```

内核在执行指令前，会先通过 `console_loglevel` 确保即使系统负载很高，调试信息也能强制从串口喷出。

### 7.5. 调试总结

* **GDB 卡住时**：检查内核是否在运行。如果是，请在另一个终端通过网络触发断点，或在 GDB 中按 `Ctrl+C` 尝试强行中断。
* **双通道思维**：永远保留一个网络（SSH）连接。在串口用于 KGDB 传输时，网络是你唯一的控制入口。

---

## 8. 编译器优化（Optimized Out）

### A. 现象描述：消失的上下文

在执行 `bt` (Backtrace) 时，经常会发现大量的参数显示为 `<optimized out>`，甚至部分调用栈末尾会出现 `corrupt stack?` 的警告。

**典型输出：**

```text
#1  0xffffffc010415a9c in write_sysrq_trigger (file=<optimized out>, buf=<optimized out>...)

```

### B. 深度解析：为什么变量会消失？

Linux 内核默认使用 **`-O2`** 优化等级编译。编译器为了极致的运行性能，会进行以下“破坏调试体验”的操作：

* **寄存器分配 (Register Allocation)**：变量不再存储在内存（栈）中，而是直接放在 CPU 寄存器里。一旦该变量的作用域结束，寄存器会被立即复用，原始数据被覆盖。
* **内联化 (Inlining)**：小函数被直接展开到调用处，不再生成独立的函数调用指令（如 `bl`），导致调用栈层级在视觉上被压缩或“消失”。
* **死代码消除**：如果编译器认为某个变量在后续流程中没有被读取，它甚至根本不会生成存储该变量的代码。

### C. 应对技巧：如何绕过优化读取数据

#### 技巧 1：利用架构调用约定 (ABI)

在 **ARM64** 架构下，函数的前 8 个参数固定通过寄存器 `$x0` 到 `$x7` 传递。即便 GDB 的源码关联失效，寄存器里的物理值依然真实存在。

* **实验案例**：当前的断点在 `__handle_sysrq (key=103, ...)`。
* **操作**：在 GDB 中输入 `p $x0`。
* **结果**：输出 `103`（即字符 'g' 的 ASCII 码）。这证明了通过硬件寄存器可以直接找回那些被标记为“消失”的参数。

#### 技巧 2：单步跟踪与反汇编

如果变量在函数执行中途消失，可以使用“组合拳”定位：

1. **反汇编当前函数**：执行 `disassemble`。
2. **查找赋值指令**：观察数据被移动到了哪个寄存器（例如 `mov x19, x0`，说明数据被备份到了 `x19`）。
3. **单步汇编指令**：使用 `nexti` (Step Instruction) 而非 `next`，实时监控寄存器的数值演变。

### D. 解决“Stack Corrupt”警告

`Backtrace stopped: previous frame identical to this frame (corrupt stack?)`

这通常不是真的内存损坏，而是由于：

1. **尾调用优化 (Tail Call Optimization)**：函数 A 在最后一行调用函数 B，编译器复用了 A 的栈帧以节省开销，导致回溯链条断裂。
2. **权限级切换**：在处理从用户态到内核态的切换点（如 `el0_svc`）时，GDB 难以跨越 Exception Level 边界读取完整的调用栈。

### E. 调试建议 (Best Practices)

* **保持冷静**：看到 `<optimized out>` 说明你正在调试真实的、未经过滤的物理逻辑。
* **降级编译 (可选)**：若必须深度追踪，可修改内核 Makefile 将 `-O2` 改为 `-O1` 或 `-Og` (Optimize for debugging)，但需注意这可能改变内核的竞态行为。
* **寄存器思维**：在 ARM64 上，永远记得查看 `info registers`，硬件寄存器不会欺骗开发者。

---

# 附：操作记录

## 路由器端

```text
root@ImmortalWrt:/#
root@ImmortalWrt:/# cat /sys/module/kgdboc/parameters/kgdboc

root@ImmortalWrt:/#
root@ImmortalWrt:/#
root@ImmortalWrt:/# echo ttyS0 > /sys/module/kgdboc/parameters/kgdboc
[   71.684432] KGDB: Registered I/O driver kgdboc
root@ImmortalWrt:/#
root@ImmortalWrt:/#
root@ImmortalWrt:/# echo g > /proc/sysrq-trigger
[   77.730323] sysrq: DEBUG

Entering kdb (current=0xffffff800ce6da00, pid 1176) on processor 0 due to Keyboard Entry
[0]kdb>

[0]kdb>
[0]kdb> error : 拒绝访问。
```

## Ubuntu虚拟机端(Win10 SSH方式：SSH终端1)

```text
root@kay-vm:immortalwrt-mt798x# gdb
gdb            gdb-add-index  gdb-multiarch  gdbserver      gdbtui         gdbus          gdbus-codegen
root@kay-vm:immortalwrt-mt798x#
root@kay-vm:immortalwrt-mt798x#
root@kay-vm:immortalwrt-mt798x# gdb-multiarch build_dir/target-aarch64_cortex-a53_musl/linux-mediatek_mt7981/linux-5.4.284/vmlinux
GNU gdb (Ubuntu 9.2-0ubuntu1~20.04.2) 9.2
Copyright (C) 2020 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
Type "show copying" and "show warranty" for details.
This GDB was configured as "x86_64-linux-gnu".
Type "show configuration" for configuration details.
For bug reporting instructions, please see:
<http://www.gnu.org/software/gdb/bugs/>.
Find the GDB manual and other documentation resources online at:
    <http://www.gnu.org/software/gdb/documentation/>.

For help, type "help".
Type "apropos word" to search for commands related to "word"...
Reading symbols from build_dir/target-aarch64_cortex-a53_musl/linux-mediatek_mt7981/linux-5.4.284/vmlinux...
(gdb) set architecture aarch64
The target architecture is assumed to be aarch64
(gdb) set remotebaud 115200
No symbol "remotebaud" in current context.
(gdb) set serial baud 115200
(gdb) target remote /dev/ttyUSB0
/dev/ttyUSB0: No such file or directory.
(gdb) target remote /dev/ttyUSB0
Remote debugging using /dev/ttyUSB0
arch_kgdb_breakpoint () at ./arch/arm64/include/asm/kgdb.h:21
21              asm ("brk %0" : : "I" (KGDB_COMPILED_DBG_BRK_IMM));
(gdb) bt
#0  arch_kgdb_breakpoint () at ./arch/arm64/include/asm/kgdb.h:21
#1  kgdb_breakpoint () at kernel/debug/debug_core.c:1165
#2  0xffffffc010146cd8 in sysrq_handle_dbg (key=<optimized out>) at kernel/debug/debug_core.c:925
#3  0xffffffc0104154ac in __handle_sysrq (key=103, check_mask=false) at drivers/tty/sysrq.c:556
#4  0xffffffc010415a9c in write_sysrq_trigger (file=<optimized out>, buf=<optimized out>, count=2, ppos=<optimized out>)
    at drivers/tty/sysrq.c:1105
#5  0xffffffc0102684ac in proc_reg_write (file=<optimized out>, buf=<optimized out>, count=<optimized out>, ppos=<optimized out>)
    at fs/proc/inode.c:238
#6  0xffffffc0101f5bd8 in __vfs_write (file=<optimized out>, p=<optimized out>, count=<optimized out>, pos=<optimized out>)
    at fs/read_write.c:494
#7  0xffffffc0101f7c20 in vfs_write (pos=<optimized out>, count=2, buf=<optimized out>, file=<optimized out>) at fs/read_write.c:558
#8  vfs_write (file=0xffffff800cf17e00, buf=0x7f88730ec0 "g\n{\210\177", count=<optimized out>, pos=0xffffffc0118bbe68)
    at fs/read_write.c:542
#9  0xffffffc0101f7ee4 in ksys_write (fd=<optimized out>, buf=0x7f88730ec0 "g\n{\210\177", count=2) at fs/read_write.c:611
#10 0xffffffc0101f7f78 in __do_sys_write (count=<optimized out>, buf=<optimized out>, fd=<optimized out>) at fs/read_write.c:623
#11 __se_sys_write (count=<optimized out>, buf=<optimized out>, fd=<optimized out>) at fs/read_write.c:620
#12 __arm64_sys_write (regs=<optimized out>) at fs/read_write.c:620
#13 0xffffffc010095824 in __invoke_syscall (syscall_fn=<optimized out>, regs=<optimized out>) at arch/arm64/kernel/syscall.c:48
#14 invoke_syscall (syscall_table=<optimized out>, sc_nr=<optimized out>, scno=<optimized out>, regs=<optimized out>)
    at arch/arm64/kernel/syscall.c:48
#15 el0_svc_common (regs=0xffffffc0118bbec0, scno=<optimized out>, syscall_table=0xffffffc0108006f0 <sys_call_table>,
    sc_nr=<optimized out>) at arch/arm64/kernel/syscall.c:114
#16 0xffffffc0100958d8 in el0_svc_handler (regs=<optimized out>) at arch/arm64/kernel/syscall.c:160
#17 0xffffffc010083988 in el0_svc () at arch/arm64/kernel/entry.S:1020
Backtrace stopped: previous frame identical to this frame (corrupt stack?)
(gdb)
(gdb) list drivers/tty/sysrq.c
Function "drivers/tty/sysrq.c" not defined.
(gdb) list drivers/tty/sysrq.c:1
1       // SPDX-License-Identifier: GPL-2.0
2       /*
3        *      Linux Magic System Request Key Hacks
4        *
5        *      (c) 1997 Martin Mares <mj@atrey.karlin.mff.cuni.cz>
6        *      based on ideas by Pavel Machek <pavel@atrey.karlin.mff.cuni.cz>
7        *
8        *      (c) 2000 Crutcher Dunnavant <crutcher+kernel@datastacks.com>
9        *      overhauled to use key registration
10       *      based upon discusions in irc://irc.openprojects.net/#kernelnewbies
(gdb)
11       *
12       *      Copyright (c) 2010 Dmitry Torokhov
13       *      Input handler conversion
14       */
15
16      #define pr_fmt(fmt) KBUILD_MODNAME ": " fmt
17
18      #include <linux/sched/signal.h>
19      #include <linux/sched/rt.h>
20      #include <linux/sched/debug.h>
(gdb)
21      #include <linux/sched/task.h>
22      #include <linux/interrupt.h>
23      #include <linux/mm.h>
24      #include <linux/fs.h>
25      #include <linux/mount.h>
26      #include <linux/kdev_t.h>
27      #include <linux/major.h>
28      #include <linux/reboot.h>
29      #include <linux/sysrq.h>
30      #include <linux/kbd_kern.h>
(gdb) list __handle_sysrq
528             struct sysrq_key_op *op_p;
529             int orig_log_level;
530             int orig_suppress_printk;
531             int i;
532
533             orig_suppress_printk = suppress_printk;
534             suppress_printk = 0;
535
536             rcu_sysrq_start();
537             rcu_read_lock();
(gdb)
538             /*
539              * Raise the apparent loglevel to maximum so that the sysrq header
540              * is shown to provide the user with positive feedback.  We do not
541              * simply emit this at KERN_EMERG as that would change message
542              * routing in the consumers of /proc/kmsg.
543              */
544             orig_log_level = console_loglevel;
545             console_loglevel = CONSOLE_LOGLEVEL_DEFAULT;
546
547             op_p = __sysrq_get_key_op(key);
(gdb) break __handle_sysrq
Breakpoint 1 at 0xffffffc010415428: file drivers/tty/sysrq.c, line 533.
(gdb) continue
Continuing.
(gdb) print sysrq_key_table['m']
[New Thread 4712]
[New Thread 4706]
[New Thread 4707]
[New Thread 4708]
[New Thread 4709]
[New Thread 4711]
[Switching to Thread 4712]

Thread 96 hit Breakpoint 1, __handle_sysrq (key=109, check_mask=false) at drivers/tty/sysrq.c:533
533             orig_suppress_printk = suppress_printk;
(gdb) print sysrq_key_table['m']
$1 = (struct sysrq_key_op *) 0x0
(gdb) bt
#0  __handle_sysrq (key=109, check_mask=false) at drivers/tty/sysrq.c:533
#1  0xffffffc010415a9c in write_sysrq_trigger (file=<optimized out>, buf=<optimized out>, count=2, ppos=<optimized out>)
    at drivers/tty/sysrq.c:1105
#2  0xffffffc0102684ac in proc_reg_write (file=<optimized out>, buf=<optimized out>, count=<optimized out>, ppos=<optimized out>)
    at fs/proc/inode.c:238
#3  0xffffffc0101f5bd8 in __vfs_write (file=<optimized out>, p=<optimized out>, count=<optimized out>, pos=<optimized out>)
    at fs/read_write.c:494
#4  0xffffffc0101f7c20 in vfs_write (pos=<optimized out>, count=2, buf=<optimized out>, file=<optimized out>) at fs/read_write.c:558
#5  vfs_write (file=0xffffff80079f2e00, buf=0x7f93b50380 "m\n\275\223\177", count=<optimized out>, pos=0xffffffc011023e68)
    at fs/read_write.c:542
#6  0xffffffc0101f7ee4 in ksys_write (fd=<optimized out>, buf=0x7f93b50380 "m\n\275\223\177", count=2) at fs/read_write.c:611
#7  0xffffffc0101f7f78 in __do_sys_write (count=<optimized out>, buf=<optimized out>, fd=<optimized out>) at fs/read_write.c:623
#8  __se_sys_write (count=<optimized out>, buf=<optimized out>, fd=<optimized out>) at fs/read_write.c:620
#9  __arm64_sys_write (regs=<optimized out>) at fs/read_write.c:620
#10 0xffffffc010095824 in __invoke_syscall (syscall_fn=<optimized out>, regs=<optimized out>) at arch/arm64/kernel/syscall.c:48
#11 invoke_syscall (syscall_table=<optimized out>, sc_nr=<optimized out>, scno=<optimized out>, regs=<optimized out>)
    at arch/arm64/kernel/syscall.c:48
#12 el0_svc_common (regs=0xffffffc011023ec0, scno=<optimized out>, syscall_table=0xffffffc0108006f0 <sys_call_table>,
    sc_nr=<optimized out>) at arch/arm64/kernel/syscall.c:114
#13 0xffffffc0100958d8 in el0_svc_handler (regs=<optimized out>) at arch/arm64/kernel/syscall.c:160
#14 0xffffffc010083988 in el0_svc () at arch/arm64/kernel/entry.S:1020
Backtrace stopped: previous frame identical to this frame (corrupt stack?)
(gdb) c
Continuing.
```

## Ubuntu虚拟机端(Win10 SSH方式：SSH终端2)

```text
kay@kay-vm:~$ ssh root@192.168.16.254 "echo m > /proc/sysrq-trigger"
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY!
Someone could be eavesdropping on you right now (man-in-the-middle attack)!
It is also possible that a host key has just been changed.
The fingerprint for the ED25519 key sent by the remote host is
Please contact your system administrator.
Add correct host key in /home/kay/.ssh/known_hosts to get rid of this message.
Offending ED25519 key in /home/kay/.ssh/known_hosts:2
  remove with:
  ssh-keygen -f "/home/kay/.ssh/known_hosts" -R "192.168.16.254"
ED25519 host key for 192.168.16.254 has changed and you have requested strict checking.
Host key verification failed.
kay@kay-vm:~$
kay@kay-vm:~$
kay@kay-vm:~$
kay@kay-vm:~$ ssh-keygen -f "/home/kay/.ssh/known_hosts" -R "192.168.16.254"
# Host 192.168.16.254 found: line 2
/home/kay/.ssh/known_hosts updated.
Original contents retained as /home/kay/.ssh/known_hosts.old
kay@kay-vm:~$
kay@kay-vm:~$
kay@kay-vm:~$
kay@kay-vm:~$
kay@kay-vm:~$ ssh root@192.168.16.254 "echo m > /proc/sysrq-trigger"  <--- 注意：在此中断触发echo m命令后，SSH终端1中会触发断点。continue后这里恢复
The authenticity of host '192.168.16.254 (192.168.16.254)' can't be established.
ED25519 key fingerprint is 
Are you sure you want to continue connecting (yes/no/[fingerprint])? yes
Warning: Permanently added '192.168.16.254' (ED25519) to the list of known hosts.
kay@kay-vm:~$
```