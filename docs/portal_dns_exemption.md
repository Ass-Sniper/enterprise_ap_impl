# Portal DNS 豁免机制

## 背景

Portal 认证场景下，未认证用户的 HTTP 请求通常会被重定向到 Portal 页面。

如果 DNS 请求也被 Portal 拦截或重定向，客户端将无法正常解析域名，导致：

- Portal 域名无法解析。
- 用户无法访问认证页面。
- 操作系统网络连通性检测（Captive Portal Detection）异常。
- 其他依赖 DNS 的网络服务无法正常工作。

因此，需要对 DNS 流量进行**豁免**，保证 DNS 请求优先处理，不受 Portal 重定向规则影响。

---

## 实现方式

通过在 `nat` 表 `PREROUTING` 链最前面挂载 DNS 处理链，实现 DNS 流量优先处理。

```sh
iptables -t nat -I PREROUTING 1 -j "$CHAIN_DNS"
```

随后再挂载 HTTP Portal 重定向规则：

```sh
iptables -t nat -I PREROUTING 2 -p tcp --dport 80 -j "$CHAIN_HTTP"
```

---

## 参数说明

### DNS Hook

```sh
iptables -t nat -I PREROUTING 1 -j "$CHAIN_DNS"
```

| 参数 | 说明 |
|------|------|
| `-t nat` | 操作 `nat` 表。 |
| `-I PREROUTING 1` | 在 `PREROUTING` 链第 1 条位置插入规则。 |
| `-j "$CHAIN_DNS"` | 将数据包跳转到自定义 DNS 处理链。 |

作用：

- 所有进入 `PREROUTING` 的数据包首先进入 `CHAIN_DNS`。
- DNS 数据包根据策略进行放行、重定向或其他处理。
- 非 DNS 数据包执行 `RETURN` 后继续匹配后续规则。

---

### HTTP Portal Hook

```sh
iptables -t nat -I PREROUTING 2 -p tcp --dport 80 -j "$CHAIN_HTTP"
```

由于 DNS Hook 已占据第 1 条规则，因此 HTTP Portal Hook 固定插入第 2 条，实现：

- DNS 优先处理。
- HTTP 请求再进入 Portal 重定向逻辑。

---

## 规则执行顺序

```text
                数据包到达
                     │
                     ▼
           nat/PREROUTING
                     │
      ┌──────────────┴──────────────┐
      │                             │
      ▼                             ▼
Rule #1                      Rule #2
CHAIN_DNS                    CHAIN_HTTP
      │                             │
DNS 放行/处理                 HTTP Portal 重定向
      │
   RETURN
      │
继续执行后续规则
```

最终规则顺序：

```text
PREROUTING
├── 1. CHAIN_DNS
├── 2. CHAIN_HTTP
├── 3. 其他 NAT 规则
└── ...
```

---

## 为什么 DNS 必须优先处理？

如果 HTTP Portal 规则位于 DNS 之前，可能导致：

- DNS 请求无法正常到达 DNS 服务器。
- Portal 域名解析失败。
- 浏览器无法打开认证页面。
- 操作系统误判网络状态。
- Portal 认证流程无法正常完成。

因此，DNS Hook 必须始终位于 `PREROUTING` 链最前面。

---

## 总结

Portal 的 DNS 豁免机制通过以下方式实现：

1. 在 `PREROUTING` 链第 1 条挂载 `CHAIN_DNS`。
2. 所有数据包首先进入 DNS 处理链。
3. DNS 流量优先处理，不受 Portal 重定向影响。
4. 非 DNS 流量返回 `PREROUTING` 后，再进入 HTTP Portal 重定向链。
5. 保证 Portal 认证页面能够正常解析和访问，同时维持整个认证流程的正常运行。