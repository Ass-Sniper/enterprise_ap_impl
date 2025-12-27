# Kernel Bring-up & Debug Knowledge Base (ARM64 / MT7981)

本目录包含 **ARM64 SoC（以 MT7981 为例）内核 bring-up、异常路径分析、
IRQ / GICv3、KGDB/KDB 调试的系统化文档**。

目标不是“能断下来”，而是：

> **可复现、可追溯、可长期维护的内核调试体系。**

---

## 📚 文档阅读建议顺序

### 1️⃣ 方法论 / 工作流（必读）

- `workflow/kernel-debug-workflow.md`  
  内核调试整体流程、artifact 管理、KGDB 使用规范

- `workflow/bootargs-modes.md`  
  调试态 / 生产态 bootargs 对照与常见陷阱

---

### 2️⃣ KGDB / KDB / SysRq

- `kgdb/kgdb-kdb-immortalwrt.md`  
  KGDB / KDB 基础与 ImmortalWrt 实践

- `kgdb/sysrq-guide.md`  
  SysRq 机制与调试入口

- `kgdb/mt7981-kgdb-ftrace.md`  
  MT7981 上 KGDB + ftrace 联合使用

---

### 3️⃣ ARM64 架构与运行期实证（核心价值）

- `arm64/arm_64_vbar_el_1_exception_vectors.md`  
  ARM64 异常向量表理论

- `arm64/arm_64_vbar_el_1_vectors_runtime_verification.md`  
  VBAR_EL1 向量运行期验证

- `arm64/arm_64_el_0_svc_kgdb_analysis.md`  
  EL0 SVC → syscall → 内核路径实证

- `arm64/arm_64_el_1_irq_gicv_3_irqdomain_runtime_verification.md`  
  EL1 IRQ → GICv3 → irqdomain → handler 完整实证

---

## 🧭 使用建议

- **新接手 / 新平台 bring-up**  
  → 先读 `workflow/`

- **调试异常 / IRQ / panic**  
  → `kgdb/` + `arm64/`

- **SoC / 内核深入分析**  
  → `arm64/` 全部文档

---

## 🔒 工程原则

- 文档只增不删
- 实证优于推测
- 所有结论应可被 KGDB / runtime 行为验证

---

**End of Index**
