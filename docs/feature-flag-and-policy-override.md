
# 特性/功能开关（FeatureFlag ）与 用户策略覆盖（Policy Override）的实际意义

> 本文档用于说明在 Portal / NAC 系统中
> **FeatureFlag** 与 **用户策略覆盖（policy:user:*）**
> 各自解决什么问题、什么时候使用、以及它们如何协同工作。

---

## 一、整体定位（一句话版）

* **FeatureFlag**：
  👉 *“这个能力现在是否允许被使用？”*

* **用户策略覆盖（Policy Override）**：
  👉 *“这个用户/设备是否需要一个例外的准入规则？”*

二者解决的问题不同，但**必须同时存在**，否则系统要么不灵活，要么不可控。

---

## 二、为什么需要 FeatureFlag？

### 1️⃣ FeatureFlag 解决的问题本质

> **FeatureFlag 的核心意义：
> 在不改代码、不改策略、不重启服务的情况下，
> 动态控制某个“认证能力 / 功能模块”是否可用。**

它关注的是：**“能力是否开放”**，而不是“谁能用”。

---

### 2️⃣ 没有 FeatureFlag 会怎样？

假设系统没有 FeatureFlag：

| 需求            | 你会被迫做的事   |
| ------------- | --------- |
| 临时关闭 SMS 登录   | 改代码 / 改策略 |
| 灰度开放 token 登录 | 新建策略      |
| 紧急回滚新认证方式     | 回滚版本      |
| 控制新功能风险       | 无能为力      |

👉 **能力控制和用户策略完全耦合，风险极高。**

---

### 3️⃣ FeatureFlag 的典型使用场景

#### ✅ 场景 A：能力级开关

```text
feature:auth:pap = 1
feature:auth:sms = 0
```

* pap 可以用
* sms 被全局关闭
* policy.yaml 不需要改

---

#### ✅ 场景 B：用户灰度

```text
feature:auth:sms:user:testuser = 1
```

* 只允许 testuser 使用 sms
* 其他用户仍然被拦截

---

#### ✅ 场景 C：紧急止血

```text
DEL feature:auth:pap
```

* 所有 pap 登录立刻被拒绝
* 不依赖发布流程

---

### 4️⃣ FeatureFlag 的边界（很重要）

❌ FeatureFlag **不应该**：

* 决定会话时长
* 决定落地页
* 决定用户角色
* 表达业务策略

👉 它只回答一个问题：**“这个能力现在开没开？”**

---

## 三、为什么需要用户策略覆盖（Policy Override）？

### 1️⃣ 用户策略覆盖解决的问题本质

> **用户策略覆盖的核心意义：
> 在不污染全局策略、不影响其他用户的前提下，
> 为“某一个用户 / 设备”提供例外准入行为。**

它关注的是：**“这个对象是不是特殊情况？”**

---

### 2️⃣ 没有用户策略覆盖会怎样？

如果系统只有 `policy.yaml`：

| 需求              | 你会被迫做的事   |
| --------------- | --------- |
| 单个用户临时放行        | 改全局策略     |
| VIP 使用特殊跳转页     | 新建 policy |
| 运维/管理员不走 Portal | 写 if else |
| 测试新策略           | 新环境       |

👉 **你会不断污染全局配置，系统越来越不可控。**

---

### 3️⃣ 用户策略覆盖的典型使用场景

#### ✅ 场景 A：VIP / 管理员

```redis
policy:user:admin = {
  "name": "admin-override",
  "allowed": ["token"],
  "sessionTimeout": 86400,
  "redirectURL": "https://intranet.example.com"
}
```

* 不影响普通员工
* 不改 policy.yaml
* 即时生效

---

#### ✅ 场景 B：现场救火 / 投诉处理

```redis
policy:user:testuser = {
  "allowed": ["pap"],
  "sessionTimeout": 600
}
```

* 临时生效
* 问题修完即可删除

---

#### ✅ 场景 C：灰度 / AB 测试

```redis
policy:user:userA = { "allowed": ["sms"] }
policy:user:userB = { "allowed": ["sms"] }
```

👉 不需要新建全局策略。

---

### 4️⃣ 用户策略覆盖的边界

❌ 不应该：

* 长期替代 policy.yaml
* 用来描述用户角色
* 保存复杂业务逻辑
* 无限堆积（应升级为正式策略）

👉 **它是“例外通道”，不是主干配置。**

---

## 四、FeatureFlag vs 用户策略覆盖（对照表）

| 维度       | FeatureFlag | 用户策略覆盖  |
| -------- | ----------- | ------- |
| 控制对象     | 功能 / 能力     | 用户 / 设备 |
| 回答的问题    | 能不能用        | 怎么用     |
| 是否影响所有用户 | 是           | 否       |
| 生命周期     | 中 / 长期      | 短 / 中期  |
| 是否参与策略决策 | 否（只 gate）   | 是       |
| 是否可灰度    | ✅           | ✅       |
| 是否应常驻    | ✅           | ❌       |

---

## 五、在 Portal / NAC 架构中的协作关系

你现在系统的决策链路（非常正确）是：

```
FeatureFlag（能力是否开放）
        ↓
Policy Override（是否有用户例外）
        ↓
Policy.yaml / RADIUS Hint（正式策略）
        ↓
Strategy（pap / sms / token）
        ↓
RADIUS / Auth Backend
```

### 这个顺序的意义：

* **FeatureFlag**：先挡风险
* **Policy Override**：再处理例外
* **Policy.yaml**：最后走常规规则

---

## 六、一个完整的现实例子

> “testuser 能否用 pap 登录？”

1. FeatureFlag：

   ```text
   feature:auth:pap = 1
   ```

   → 能力开放

2. 用户策略覆盖：

   ```text
   policy:user:testuser = 不存在
   ```

   → 无例外

3. policy.yaml：

   ```yaml
   allowed: ["pap"]
   ```

   → 策略允许

4. Strategy：
   → pap

5. RADIUS：
   → 接管认证

**任何一层 deny，都会明确可定位。**

---

## 七、设计总结（工程级）

> **FeatureFlag 让系统“敢上线”
> 用户策略覆盖让系统“能救火”**

二者结合，系统才能从：

* 能跑
  → 能用
  → **可运维、可灰度、可应急**

---

## 八、推荐后续增强（可选）

* 给 `policy:user:*` 增加 TTL / audit
* 提供 `/debug/policy?user=` API
* 提供 `/debug/feature` 可视化接口
* 扩展到 `policy:mac:* / policy:nas:*`

---

