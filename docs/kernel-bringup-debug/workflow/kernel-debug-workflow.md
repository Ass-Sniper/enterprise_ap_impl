
# Kernel Debug Workflow (MT7981 / ImmortalWrt)

本文档描述一套 **可复现、可追溯、工程级** 的内核调试流程，适用于：

- SoC：MediaTek MT7981
- 板卡：HiLink RM65
- 系统：ImmortalWrt / OpenWrt
- 内核：Linux 5.4.x
- 调试方式：KGDB + GDB（串口）

目标是解决以下核心问题：

> **“某一次内核调试，到底是基于哪一个固件、哪一个内核、哪一份符号完成的？”**

---

## 1. 总体设计原则

### 1.1 固件与符号必须绑定

- 运行在设备上的 **sysupgrade.bin**
- 主机上用于调试的 **vmlinux（debug symbols）**

**必须作为一个不可分割的整体进行归档**。

任何只保存 vmlinux、不保存固件的调试结果，都是不可复现的。

---

### 1.2 构建产物 ≠ 调试工件

- `build_dir/` 下的文件：
  - 可能被重新编译覆盖
  - **不适合长期依赖**
- 调试必须基于：
  - **显式拷贝**
  - **版本化命名**
  - **校验绑定**

---

### 1.3 单 UART 场景下的现实约束

在仅有一个串口的嵌入式设备上：

- console
- printk
- KGDB

**共享同一 UART 时极不稳定**。

因此：
- IRQ / el1_irq / gic_handle_irq 等高频路径  
  **必须极度克制断点数量**
- 调试更强调：
  - 实证
  - 路径验证
  - 一次命中

---

## 2. 目录结构约定

```text
tools/kernel-debug/
├── scripts/
│   └── package-kernel-debug.sh
├── artifacts/
│   └── mediatek-mt7981-hilink_rm65-5.4.284-YYYYMMDD-HHMMSS.tar.gz
└── docs/
    └── README.md
```

### 各目录职责

* `scripts/`

  * 所有**调试相关自动化脚本**
* `artifacts/`

  * **只增不删**
  * 每个 tar.gz = 一次真实调试世界线
* `docs/`

  * 原理说明
  * 实证过程
  * 调试结论

---

## 3. Debug Artifact 内容规范

每一个 debug artifact **必须至少包含**：

```text
sysupgrade.bin      # 实际刷机固件
vmlinux             # 内核符号（with debug_info）
System.map          # 符号地址对照
kernel.config       # 内核配置（可复现性）
VERSION.txt         # 元信息 + SHA1 校验
```

### VERSION.txt 示例

```text
Target            : mediatek-mt7981
Board             : hilink_rm65
Kernel             : Linux 5.4.284
Build Time        : 2025-12-27 15:42:33

VMLINUX SHA1      : xxxx
Sysupgrade SHA1   : yyyy
```

---

## 4. 打包流程（标准操作）

### 4.1 构建完成后

```bash
make -j$(nproc)
```

确认以下文件存在：

```bash
bin/targets/mediatek/mt7981/*sysupgrade.bin
build_dir/.../vmlinux
```

---

### 4.2 执行打包脚本

在源码根目录：

```bash
./tools/kernel-debug/scripts/package-kernel-debug.sh
```

输出：

```text
tools/kernel-debug/artifacts/
└── mediatek-mt7981-hilink_rm65-5.4.284-YYYYMMDD-HHMMSS.tar.gz
```

---

## 5. 运行与调试关系说明

### 5.1 设备侧

* 刷写：

  ```bash
  sysupgrade sysupgrade.bin
  ```
* 启动参数：

  * 启用 KGDB（推荐 kgdbwait）
  * 尽量降低 loglevel

---

### 5.2 主机侧

* 解压 artifact：

  ```bash
  tar xzf mediatek-mt7981-*.tar.gz
  ```

* 启动 GDB：

  ```bash
  gdb-multiarch vmlinux
  ```

* 连接 KGDB：

  ```gdb
  set architecture aarch64
  set serial baud 115200
  target remote /dev/ttyUSB0
  ```

---

## 6. IRQ / el1_irq 实证调试规范

### 6.1 推荐断点层级（由粗到细）

1. `el1_irq`
2. `gic_handle_irq`
3. `handle_domain_irq`
4. `__handle_domain_irq`
5. `generic_handle_irq`
6. `handle_irq_event_percpu`

⚠️ **不建议长期同时开启所有断点**。

---

### 6.2 单 UART 场景下的经验法则

* 只保留 **一个关键断点**
* 命中后：

  * `bt`
  * `info registers`
  * `c`
* 不做单步
* 不在 IRQ 中 printf / printk

---

## 7. 常见问题与经验总结

### 7.1 GDB 报错 `Invalid remote reply`

原因：

* console 输出混入 KGDB 协议

解决：

* 降低 loglevel
* 减少断点
* 优先考虑 UART 分离

---

### 7.2 build_dir 下的 vmlinux 能不能直接用？

不建议。

原因：

* 可能被下一次编译覆盖
* 无版本绑定

**必须通过 artifact 归档使用。**

---

## 8. 本流程的最终目标

这套流程的目标不是：

* “能不能断下来”

而是：

> **“半年后还能不能解释清楚当时发生了什么”**

当你能做到：

* 固件可追溯
* 符号可验证
* 过程可复盘

你就已经站在 **SoC / 内核 bring-up 级调试能力** 的门槛之上。

---

## 9. 后续可扩展方向

* 增加 `verify-running-firmware.sh`
* 增加 KGDB 专用 `.gdbinit`
* 将 el0_svc / el1_irq / GICv3 / irqdomain 分别整理成独立文档
* 与 crash / vmcore 联动

---

好，这里我已经把 **“为什么调试阶段要去掉 `panic=0`”** 这一整套**工程级解释**，整理成一段**可直接补充进你现有文档的 Markdown**。

你可以 **原样追加** 到：

```
tools/kernel-debug/docs/README.md
```

建议放在 **IRQ / KGDB 调试规范之后**，作为一个“调试参数陷阱”章节。

---

## 10. 关于 `panic=0` 与 KGDB 调试的关系（重要）

在内核启动参数中，`panic=<timeout>` 决定了 **内核发生 panic 后的行为**：

| 参数 | 行为 |
|----|----|
| `panic=0` | panic 后 **立刻重启** |
| `panic=N` | panic 后 N 秒重启 |
| 不设置 | panic 后 **停住，保留现场**（默认） |

### 10.1 为什么调试阶段不应使用 `panic=0`

在 KGDB / IRQ / 异常路径调试场景中，`panic=0` 会直接破坏调试能力。

KGDB 的核心设计目标是：

> **在问题发生的瞬间冻结系统状态，让调试器接管。**

而 `panic=0` 的行为恰好相反：

1. 内核触发 panic  
2. 现场刚刚打印  
3. **系统立刻复位重启**  
4. KGDB 来不及介入  
5. 调试信息永久丢失  

这在以下调试场景中尤其致命：

- `el1_irq`
- `gic_handle_irq`
- `irqdomain`
- softirq / tasklet
- 内核 bring-up 阶段的不稳定路径

这些路径 **极易 panic，但正是最有价值的调试对象**。

---

### 10.2 KGDB 与 panic / oops 的理想协作关系

在 **不设置 `panic=`** 的情况下：

- panic / oops 发生后
- CPU 停住
- 栈、寄存器、IRQ 状态保持
- KGDB / GDB 可接管分析

这正是 KGDB、crash、JTAG 等工具所期望的状态。

---

### 10.3 什么时候 `panic=0` 是正确选择？

`panic=0` 并非“错误参数”，而是 **使用场景不同**。

#### ✅ 适合使用 `panic=0` 的场景

- 生产环境
- 无人值守设备
- 路由器 / AP 必须自动恢复
- 不进行现场调试

#### ❌ 不适合使用 `panic=0` 的场景

- 内核 bring-up
- KGDB 调试
- IRQ / 异常路径分析
- SoC 初期验证

---

### 10.4 推荐的调试阶段 bootargs（最终版）

```text
console=ttyS0,115200n1
earlycon=uart8250,mmio32,0x11002000
loglevel=1
kgdboc=ttyS1,115200
kgdbwait
````

该组合具备以下特性：

* UART 分离（console / KGDB）
* panic 时保留现场
* IRQ 调试不刷屏
* KGDB 稳定可复现

---

### 10.5 调试态 vs 生产态参数对照

| 场景          | 建议               |
| ----------- | ---------------- |
| 内核调试 / KGDB | **不设置 `panic=`** |
| 生产运行        | `panic=0`        |
| IRQ / 异常调试  | 禁止 `panic=0`     |
| 无人值守系统      | 建议 `panic=0`     |

---

> **总结：**
>
> 在调试阶段，目标是 **“死得清楚”**；
> 在生产阶段，目标是 **“死得快、能自愈”**。
>
> `panic=0` 只适合后者，不适合前者。

---

## 11. 相关子文档索引

- [Bootargs 调试态 / 生产态对照](bootargs-modes.md)
- KGDB 专用 GDB 脚本：
  - `tools/kernel-debug/scripts/kgdb.gdbinit`

建议在进行 IRQ / el1_irq / GICv3 调试前，
**务必确认当前设备使用的是调试态 bootargs**，
并使用上述 `.gdbinit` 连接 KGDB。

