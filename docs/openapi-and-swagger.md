
# OpenAPI & Swagger 编译与使用说明

本文档说明 ap-controller 中 **OpenAPI 规范的生成方式、Swagger 的暴露策略**，
以及 **API Handler 注释的编写规范**。

本文面向：
- 控制面开发者
- CI / 构建维护者
- API 使用方（联调 / 自动化）

---

## 1. 设计目标

本项目对 OpenAPI / Swagger 的设计遵循以下原则：

- **编译期生成 OpenAPI**（保证接口契约与代码一致）
- **运行期可选择是否暴露 Swagger UI**
- **生产环境默认不暴露 UI**
- **机器接口（/openapi.json）始终稳定**

---

## 2. 核心参数说明

### 2.1 GEN_OPENAPI（编译期参数）

**类型**：Docker build ARG  
**默认值**：`true`

```bash
--build-arg GEN_OPENAPI=true|false
````

#### 作用

* 是否在 Docker build 阶段运行 `swag init`
* 是否启用 Go build tag：`swagger`

#### 行为说明

| GEN_OPENAPI | 行为                              |
| ----------- | ------------------------------- |
| `true`      | 生成 OpenAPI 文档 + 编译 swagger 支持代码 |
| `false`     | 不生成文档 + 编译不包含 swagger 代码        |

> ⚠️ 若 `GEN_OPENAPI=false`，二进制中 **不会包含任何 swagger 相关代码**

---

### 2.2 ENABLE_SWAGGER_UI（运行期参数）

**类型**：容器运行期环境变量
**默认值**：`false`

```bash
ENABLE_SWAGGER_UI=true|false
```

#### 作用

* 是否暴露 Swagger UI 页面 `/swagger/`

#### 行为说明

| ENABLE_SWAGGER_UI | 行为             |
| ----------------- | -------------- |
| `true`            | 暴露 `/swagger/` |
| `false`           | 不暴露 UI         |

> ✅ **即使 ENABLE_SWAGGER_UI=false，/openapi.json 仍然可用**

---

## 3. 对外接口路径说明

### 3.1 OpenAPI JSON（机器接口）

```text
GET /openapi.json
```

* 标准 OpenAPI v2 (Swagger 2.0)
* 供：

  * API Gateway
  * 自动化测试
  * 文档生成
  * 合规审计

示例：

```bash
curl http://localhost:8443/openapi.json | jq '.info'
```

---

### 3.2 Swagger UI（人机接口）

```text
GET /swagger/
```

* 仅用于开发 / 调试
* 默认关闭
* 受 `ENABLE_SWAGGER_UI` 控制

---

## 4. Docker 构建示例

### 4.1 本地开发（完整调试）

```bash
docker build \
  --build-arg GEN_OPENAPI=true \
  --build-arg ENABLE_SWAGGER_UI=true \
  -t ap-controller-go:debug \
  ./control-plane/ap-controller-go
```

---

### 4.2 CI / 生产构建（不暴露 UI）

```bash
docker build \
  --build-arg GEN_OPENAPI=true \
  --build-arg ENABLE_SWAGGER_UI=false \
  -t ap-controller-go:release \
  ./control-plane/ap-controller-go
```

---

### 4.3 极速构建（无 OpenAPI）

```bash
docker build \
  --build-arg GEN_OPENAPI=false \
  -t ap-controller-go:minimal \
  ./control-plane/ap-controller-go
```

---

## 5. API Handler 注释编写规范（非常重要）

本项目使用 `swag` 生成 OpenAPI，**注释必须符合以下规范**。

---

### 5.1 基本规则

* 所有 `@Param body` **必须引用已命名 struct**
* ❌ 禁止使用 inline struct / map / anonymous type
* 每个 Handler 必须包含：

  * `@Summary`
  * `@Description`
  * `@Tags`
  * `@Router`

---

### 5.2 标准示例（POST + JSON Body）

```go
// portalLogin handles portal login.
// @Summary 门户登录授权
// @Description 根据 MAC / SSID / Auth 等信息创建或更新会话
// @Tags Portal
// @Accept json
// @Produce json
// @Param body body LoginReq true "登录请求体"
// @Success 200 {object} map[string]interface{} "authorized=true 时返回会话信息"
// @Failure 400 {object} ErrorResponse "bad_json"
// @Failure 422 {object} ErrorResponse "mac_required"
// @Router /portal/login [post]
func portalLogin(w http.ResponseWriter, r *http.Request) {
    // handler logic
}
```

---

### 5.3 Path 参数示例（GET）

```go
// @Summary 门户状态查询
// @Description 查询单个 MAC 的授权状态
// @Tags Portal
// @Produce json
// @Param mac path string true "客户端 MAC 地址" example(aa:bb:cc:dd:ee:ff)
// @Success 200 {object} map[string]interface{} "会话状态"
// @Router /portal/status/{mac} [get]
```

---

### 5.4 Body Struct 示例（model.go）

```go
// LoginReq portal login request
type LoginReq struct {
    MAC     string `json:"mac" example:"aa:bb:cc:dd:ee:ff"`
    SSID    string `json:"ssid,omitempty" example:"GuestWiFi"`
    Auth    string `json:"auth,omitempty" example:"portal"`
    APID    string `json:"ap_id,omitempty" example:"ap-123"`
    RadioID string `json:"radio_id,omitempty" example:"radio-1"`
    IP      string `json:"ip,omitempty" example:"192.168.1.23"`
}
```

---

## 6. 常见错误速查

### ❌ 1. `missing required param comment`

原因：

* 使用了 inline struct

修复：

* 抽取为命名 struct

---

### ❌ 2. `/openapi.json` 404

原因：

* Swagger handler 未注册到 router

修复：

* 确认 `registerSwagger(r)` 被调用

---

### ❌ 3. 编译时报 `docs/openapi not in std`

原因：

* 未启用 `swagger` build tag
* 但代码 import 了 openapi

修复：

* GEN_OPENAPI=true 时使用 `-tags swagger`

---

## 7. 总结

* **GEN_OPENAPI 决定是否生成 OpenAPI**
* **ENABLE_SWAGGER_UI 决定是否暴露 UI**
* **/openapi.json 是标准机器接口**
* **/swagger/ 仅用于调试**
