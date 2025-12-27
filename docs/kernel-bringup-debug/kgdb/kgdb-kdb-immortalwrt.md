
# ImmortalWrt 上配置与使用 KGDB / KDB 实战指南

> 适用场景：  
> - ImmortalWrt / OpenWrt  
> - MT798x / IPQ / 嵌入式 SoC  
> - 内核崩溃、驱动调试、死机定位  
>
> 目标：  
> 在嵌入式设备上 **真正用起来 KGDB / KDB**，而不仅仅是“能编译”。

---

## 一、KGDB / KDB 是什么？

### 1. KGDB（Kernel GNU Debugger）

- 内核态调试器
- 可配合 **外部 GDB**
- 支持：
  - 下断点
  - 单步执行
  - 查看内核变量 / 结构体
  - 精确定位 crash 点

> KGDB ≠ printk  
> KGDB 是“现场调试器”，printk 只是“事后日志”。

---

### 2. KDB（Kernel Debugger Shell）

- KGDB 的 **内核内前端**
- 直接在 **串口控制台** 使用
- 不依赖 PC / GDB
- 非常适合嵌入式设备

示例：
```text
kdb> bt
kdb> ps
kdb> dmesg
````

> 实际工程中：
> **90% 场景用 KDB，10% 深度问题再上 KGDB + GDB**

---

## 二、menuconfig 中各选项解释

你在 menuconfig 里看到的选项：

```text
KGDB: kernel debugger
[ ]   KGDB: internal test suite
[*]   KGDB_KDB: include kdb frontend for kgdb
(0x1)   KDB: Select kdb command functions to be enabled by default
(0)     KDB: continue after catastrophic errors
```

### 1. `KGDB: kernel debugger`

* 核心选项，必须开启
* 否则后续所有功能无效

✅ **必须开启**

---

### 2. `KGDB: internal test suite`

* 内核开发者自测用
* 会引入额外测试代码

❌ **不建议开启（嵌入式无意义）**

---

### 3. `KGDB_KDB: include kdb frontend for kgdb`

* 启用 KDB 内核调试 shell
* 可以不接 PC，直接串口调试

✅ **强烈建议开启**

---

### 4. `KDB: Select kdb command functions`

* 控制 KDB 默认可用命令的 bitmask

推荐值：

```text
0x1f
```

包含：

* 栈回溯
* 进程查看
* 内存读写
* 寄存器查看

---

### 5. `KDB: continue after catastrophic errors`

* 内核发生严重错误后是否继续运行

推荐：

```text
0
```

原因：

* catastrophic error 后系统状态已不可信
* 继续运行无调试价值

---

## 三、ImmortalWrt 推荐配置组合（可直接照抄）

```text
Kernel hacking  --->
    [*] KGDB: kernel debugger
    [*] KGDB_KDB: include kdb frontend for kgdb
    (0x1f) KDB: Select kdb command functions
    (0)   KDB: continue after catastrophic errors
```

并且务必开启：

```text
Kernel hacking  --->
    [*] Kernel debugging
    [*] Compile the kernel with debug info
```

否则：

* 看不到函数名
* 栈全是 `??`

---

## 四、使用方式一：KDB（嵌入式首选）

### 1. 启动时进入 KDB

在 bootargs 中加入：

```text
kgdbwait
```

示例（DTS / bootargs）：

```dts
bootargs = "console=ttyS0,115200n8 kgdbwait";
```

启动后串口会停在：

```text
Entering kdb (current=0xffff8000xxxxxxxx)
kdb>
```

---

### 2. 常用 KDB 命令速查

```text
kdb> bt              # 当前 CPU 内核栈
kdb> ps              # 进程列表
kdb> dmesg           # 内核日志
kdb> lsmod           # 模块列表
kdb> md <addr>       # 内存 dump
kdb> go              # 继续运行
```

> 调网卡驱动 / DMA / HNAT / PPE 时非常有用

---

## 五、使用方式二：KGDB + 外部 GDB（进阶）

### 1. 准备 vmlinux（带符号）

ImmortalWrt 编译后生成：

```text
build_dir/target-*/linux-*/vmlinux
```

⚠️ 不要 strip！

---

### 2. PC 端启动 GDB

```bash
aarch64-linux-gnu-gdb vmlinux
```

---

### 3. 连接目标板

#### 串口方式：

```gdb
target remote /dev/ttyUSB0
```

#### 以太网（kgdboe）：

```gdb
target remote 192.168.1.1:1234
```

---

### 4. 常见调试示例

```gdb
b mtk_eth_soc_init
b netif_receive_skb
c
```

---

## 六、常见坑与解决方案

### 1. 栈全是 `??`

原因：

* 未开启 debug info
* vmlinux 被 strip

解决：

* 开启 `Compile the kernel with debug info`
* 使用原始 vmlinux

---

### 2. KDB / KGDB 没反应

原因：

* 串口被 earlycon / U-Boot 抢占
* console 参数错误

解决：

* 确认 `console=ttyS0`
* 避免多个 earlycon

---

### 3. 内核 panic 后直接重启

解决：

```text
panic=0
```

加入 bootargs，便于进入 KDB 查看现场

---

## 七、什么时候该用 KGDB / KDB？

### 非常适合

* Wi-Fi / Ethernet 驱动
* DMA / RX-TX ring 崩溃
* Soft lockup / hard lockup
* netif_receive_skb 路径异常
* HNAT / PPE offload 问题

---

### 不适合

* 普通配置错误
* UCI / shell 脚本
* 用户态程序（直接用 gdb）

---

## 八、工程师级总结

> * **KDB 是嵌入式内核调试第一生产力**
> * **KGDB + GDB 是深度剖析利器**
> * printk 只能“拍照”，KGDB 才能“单步回放事故现场”

---

## 九、建议归档位置

推荐放在以下任一位置：

```text
docs/kernel-debug/kgdb-kdb-immortalwrt.md
```

或：

```text
docs/debug/immortalwrt-kgdb-kdb.md
```

---

下一步：

- ✅ 拆成 **《KDB 命令速查表》**
- ✅ 增加 **MT7981 / IPQ5018 专用 bootargs 模板**
- ✅ 补一节 **“真实 panic → KDB 定位示例”**

---
