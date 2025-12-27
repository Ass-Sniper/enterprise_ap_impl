# Kernel Bootargs Modes: Debug vs Production

本文档定义 **内核调试态（Debug）** 与 **生产态（Production）**
两种启动参数（bootargs）的标准配置与适用场景。

目标是避免以下常见问题：

- 调试环境误用生产参数，导致 **现场丢失**
- 生产环境误用调试参数，导致 **系统不可用**
- UART / KGDB / panic 行为混乱

---

## 1. 设计原则

### 1.1 调试态（Debug）

调试态的核心目标是：

> **系统出问题时，必须“停得住、看得清、能接管”。**

因此强调：

- KGDB 稳定
- UART 分离
- panic 保留现场
- 控制 printk 噪声

---

### 1.2 生产态（Production）

生产态的核心目标是：

> **系统必须自动恢复，不能卡死。**

因此强调：

- 自动重启
- 日志可控
- 不依赖外部调试器

---

## 2. 启动参数对照表（核心）

| 参数项 | 调试态（Debug） | 生产态（Production） | 说明 |
|----|----|----|----|
| `console=` | `ttyS0,115200n1` | `ttyS0,115200n1` | 主控制台 |
| `earlycon=` | 启用 | 可选 | 早期启动日志 |
| `loglevel=` | `1` | `4~7` | 调试态必须降噪 |
| `ignore_loglevel` | ❌ 禁用 | 可选 | 调试态禁止 |
| `kgdboc=` | `ttyS1,115200` | ❌ 禁用 | UART 分离 |
| `kgdbwait` | ✅ 启用 | ❌ 禁用 | 启动即停 |
| `panic=` | ❌ 不设置 | `panic=0` | 调试态保留现场 |
| `nowatchdog` | 可选 | ❌ 禁用 | 避免调试误判 |
| `quiet` | 可选 | ✅ 建议 | 生产环境降噪 |

---

## 3. 推荐配置示例

### 3.1 调试态（Debug）推荐 bootargs

```text
console=ttyS0,115200n1
earlycon=uart8250,mmio32,0x11002000
loglevel=1
kgdboc=ttyS1,115200
kgdbwait
```

#### 特性总结：

* UART0：console / shell
* UART1：KGDB（纯协议）
* panic 时系统停住
* 适合 IRQ / el1_irq / GIC / bring-up 调试

---

### 3.2 生产态（Production）推荐 bootargs

```text
console=ttyS0,115200n1
loglevel=6
panic=0
quiet
```

#### 特性总结：

* panic 后自动重启
* 不依赖 KGDB
* 适合无人值守设备

---

## 4. 常见误用警告（非常重要）

### ❌ 在调试态使用 `panic=0`

后果：

* panic 现场丢失
* KGDB 无法介入
* 调试价值归零

---

### ❌ 在调试态使用 `ignore_loglevel`

后果：

* printk 混入 KGDB 串口
* GDB 协议损坏
* 出现 `Invalid remote reply`

---

### ❌ 在生产态启用 `kgdbwait`

后果：

* 设备卡在启动阶段
* 无法自动恢复

---

## 5. 结论

| 场景  | 关键词             |
| --- | --------------- |
| 调试态 | 停得住 / 看得清 / 可接管 |
| 生产态 | 自动恢复 / 稳定运行     |

**bootargs 是调试体系的一部分，而不是随意拼接的字符串。**

---

