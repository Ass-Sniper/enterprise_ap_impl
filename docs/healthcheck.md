
# Health Check & Self-Test

## Overview

健康检查（Health Check）用于验证 **AP 数据面是否处于可用、正确、可预期状态**。
本项目将健康检查视为 **一等公民能力**，而不是调试工具。

Health checks are first-class features, designed to continuously validate
dataplane correctness and prevent silent misconfiguration.

---

## Design Goals

- 启动即校验（Fail Fast）
- 可自动化（Controller / CI / Prometheus）
- 可扩展（插件化）
- 不破坏现有流量
- 与真实运行路径一致（Runtime-level validation）

---

## Entry Point

### portal-agent --check

```bash
portal-agent.sh --check
```

该命令会：

1. 执行所有内建检查
2. 调用 self-test 脚本
3. 汇总结果
4. 返回整体健康状态

---

## Exit Codes

| Code | Meaning             |
| ---- | ------------------- |
| `0`  | Dataplane healthy   |
| `1`  | Dataplane unhealthy |
| `2`  | Partial / degraded  |

Controller 或上层系统可据此做决策。

---

## Built-in Checks

portal-agent 内置检查包括：

### 1️⃣ dnsmasq Process Check

* dnsmasq 是否运行
* 是否为系统唯一实例

```bash
ps w | grep [d]nsmasq
```

---

### 2️⃣ dnsmasq ipset Capability

验证 dnsmasq 是否支持 ipset：

```bash
dnsmasq --version | grep ipset
```

---

### 3️⃣ dnsmasq Configuration

验证是否启用：

* `conf-dir=/tmp/dnsmasq.d`
* 或 `conf-file`

---

### 4️⃣ IPSet Existence

验证关键 ipset 是否存在：

* portal_allow_guest
* portal_allow_staff
* portal_bypass_mac
* portal_bypass_ip
* portal_bypass_dns

---

### 5️⃣ IPSet Type Validation

重点校验：

```text
portal_bypass_dns -> hash:ip
```

错误类型（如 `hash:mac`）会直接判定失败。

---

## Runtime DNS → IPSet Test

### Purpose

验证 **dnsmasq 是否真的在解析时向 ipset 写入 IP**，
这是整个 DNS Bypass 机制的核心。

---

### Test Logic

1. 记录 ipset 初始成员数量
2. 触发真实 DNS 查询
3. 等待 dnsmasq 处理
4. 再次统计 ipset 成员数量
5. 判断是否有新增 IP

---

### Example Code Snippet

```sh
before="$(ipset list "$IPSET_NAME" | awk '/Members:/ {f=1;next} f {print}' | wc -l)"

nslookup "$TEST_DOMAIN" >/dev/null 2>&1 || true
sleep 1

after="$(ipset list "$IPSET_NAME" | awk '/Members:/ {f=1;next} f {print}' | wc -l)"

[ "$after" -gt "$before" ]
```

---

## Self-Test Script

### Script

```text
data-plane/tools/portal-dnsmasq-ipset-selftest.sh
```

---

### Usage

```bash
portal-dnsmasq-ipset-selftest.sh \
  portal_bypass_dns \
  www.microsoft.com
```

---

### What It Checks

* dnsmasq 运行状态
* dnsmasq 编译选项
* dnsmasq 配置目录
* ipset 存在性
* ipset 类型
* DNS → ipset 写入行为

---

## Plugin-based Health Checks

### Plugin Directory

```text
/usr/lib/portal/health.d/
```

---

### Plugin Contract

每个插件必须：

* 可独立执行
* 返回 exit code
* 可输出 JSON 或 text

示例：

```sh
#!/bin/sh
echo '{"name":"dnsmasq","status":"ok"}'
exit 0
```

---

### Execution Order

```text
portal-agent
 ├── core checks
 ├── selftest
 └── health.d/*.sh
```

---

## JSON Output Mode

### Usage

```bash
portal-agent.sh --check --json
```

---

### Example Output

```json
{
  "status": "healthy",
  "checks": [
    {"name": "dnsmasq", "status": "ok"},
    {"name": "ipset", "status": "ok"},
    {"name": "dns_bypass", "status": "ok"}
  ]
}
```

---

## Integration Scenarios

### Controller Integration

* Controller 定期调用 `--check`
* 标记 AP Online / Degraded / Offline

---

### Prometheus / Monitoring

* JSON 输出可直接采集
* exit code 用于告警

---

## Failure Handling Strategy

| Failure           | Action                  |
| ----------------- | ----------------------- |
| dnsmasq down      | Restart / alert         |
| ipset missing     | Reapply dataplane       |
| dns bypass broken | Disable portal redirect |

---

## Best Practices

* 在 AP 启动完成后执行 `--check`
* 在策略更新后执行 `--check`
* 在升级前后执行 `--check`
* 失败即回滚

---

## Summary

Health Check 是系统稳定性的 **最后一道防线**：

* 防止 silent failure
* 防止错误策略上线
* 为自动化运维提供可靠信号

在 Captive Portal 系统中，
**“能跑 ≠ 正确，正确 ≠ 健康”**，
而健康检查正是两者之间的桥梁。
