
# 证书文档说明（Certificate Documentation）

本目录包含用于创建服务器证书的脚本。
要生成一套默认（即测试用途）的证书，只需执行：

```bash
$ ./bootstrap
```

`openssl` 命令将基于此目录中包含的示例配置文件运行，并生成一个**自签名的证书颁发机构（即根 CA）**以及一个**服务器证书**。
该“根 CA”需要安装到所有需要进行 **EAP-TLS、PEAP 或 EAP-TTLS** 认证的客户端设备上。

服务器证书会自动包含 **“TLS Web Server”** 的 **扩展密钥用法（EKU, Extended Key Usage）** 字段。
如果缺少这些扩展字段，许多客户端将拒绝与 FreeRADIUS 进行认证。

根 CA 证书以及 “XP Extensions” 文件中还包含 `crlDistributionPoints` 属性。
许多系统在验证 RADIUS 服务器证书时都需要该字段存在。

RADIUS 服务器**必须**定义该 URI，而 CA **不一定必须**，但从最佳实践角度来说，CA 也应当提供一个证书吊销列表（CRL）的 URI。
需要注意的是，尽管 Windows Mobile 客户端在执行 802.1X 认证时并不能真正使用 CRL，但仍然建议该 URI 指向一个**真实可访问的 URL**，并且包含有效的吊销列表文件。因为其他操作系统行为或未来版本的系统可能会使用该 URI。

在 Windows 系统中，需要导入 `p12` 和/或 `der` 格式的证书；
Linux 系统则需要使用 `pem` 格式的证书。

一般来说，在 802.1X（EAP）认证中，**应当使用自签名证书**。
如果你在 `ca_file` 中列出其他组织的根 CA，则意味着你允许它们**冒充你的服务器**、**为你的用户进行认证**，甚至**为 EAP-TLS 签发客户端证书**。

如果你已经拥有现成的 CA 证书和服务器证书，请重命名（或删除）本目录，并新建一个 `certs` 目录来存放你自己的证书。
注意：`make install` 命令**不会覆盖**已有的 `raddb/certs` 目录。

---

## FreeRADIUS 的新安装建议

我们建议在新安装时，先使用测试证书进行初始验证，然后再创建正式证书用于真实用户认证。
下面是创建各类证书的说明。

旧的测试证书可以通过以下命令删除：

```bash
$ make destroycerts
```

然后，按照下文步骤创建正式证书。

如果你不打算启用 **EAP-TLS、PEAP 或 EAP-TTLS**，请从 `raddb/mods-available/eap` 文件中删除相应的子配置段。
更多说明请参考该文件中的注释。

---

## 创建根证书（Root CA）

我们建议使用**私有证书颁发机构（Private CA）**。
虽然在多个客户端设备上安装该 CA 可能比较麻烦，但从整体上看更安全。

```bash
$ vi ca.cnf
```

* 编辑 `default_days`，设置 CA 证书的有效期
* 编辑 `input_password` 和 `output_password`，设置 CA 私钥的密码
* 编辑 `[certificate_authority]` 区段，填写正确的国家、省份等信息

创建 CA 证书：

```bash
$ make ca.pem
```

生成 Windows 所需的 DER 格式：

```bash
$ make ca.der
```

---

## 创建服务器证书（Server Certificate）

以下步骤将生成用于 **EAP-TLS、PEAP 和 TTLS** 等基于 TLS 的 EAP 方法的服务器证书。
如果你需要创建 `inner-server.pem`（用于在另一层 TLS 内部运行的 EAP-TLS），请按相同步骤操作。

```bash
$ vi server.cnf
```

* 编辑 `default_days`，设置服务器证书有效期

  * 为兼容所有客户端，**最大不得超过 825 天**
* 编辑 `input_password` 和 `output_password`，设置服务器证书的密码
* 编辑 `[server]` 区段，填写国家、省份等信息

  * 注意：`commonName` **必须与 CA 的 commonName 不同**

生成服务器证书：

```bash
$ make server
```

---

### 使用公共 CA 证书

如果你希望使用已有的公共证书颁发机构，可以按上述方式编辑 `server.cnf`，然后执行：

```bash
$ make server.csr
```

该命令会生成一个 **证书签名请求（CSR）**，可提交给公共 CA 进行签发。

---

## 创建客户端证书（Client Certificate）

客户端证书用于 **EAP-TLS**，并可选用于 **EAP-TTLS 和 PEAP**。
以下步骤将生成由前面创建的 CA 签名的客户端证书。

你需要在 `ca.cnf` 中正确设置 CA 私钥的 `input_password` 和 `output_password`。

```bash
$ vi client.cnf
```

* 编辑 `default_days`，设置客户端证书有效期
* 编辑 `input_password` 和 `output_password`，设置客户端证书的密码

  * 这些密码需要提供给最终使用证书的用户
* 编辑 `[client]` 区段，填写国家、省份等信息

  * **`commonName` 必须与用于登录的 `User-Name` 完全一致**

生成客户端证书：

```bash
$ make client
```

生成的客户端证书文件名为 `emailAddress.pem`，例如：

```text
user@example.com.pem
```

如需创建另一个客户端证书，只需重复以上步骤，并使用不同的 `commonName` 和密码。

---

## 性能说明（Performance）

EAP-TLS、TTLS 和 PEAP 的性能瓶颈主要来自 **SSL 计算开销**。

普通系统在使用 PAP 认证时可处理约 **10,000 包/秒**，
而 SSL 涉及 RSA 计算，其计算成本非常高。

可使用以下命令进行性能基准测试：

```bash
$ openssl speed rsa
```

或测试 2048 位密钥：

```bash
$ openssl speed rsa2048
```

输出结果表示 **每秒最多可处理的 EAP-TLS（或 TTLS / PEAP）认证次数**。

在实际环境中，真实性能通常只有该数值的一半左右。
原因是 EAP 需要大量往返数据包，而 `openssl speed` 只测试 RSA 运算本身，不包含网络交互。

---

## 兼容性说明（Compatibility）

使用本方法生成的证书已被验证可兼容 **所有操作系统**。
以下是一些常见问题：

* iOS 与 macOS 对证书有额外要求，详见：
  [https://support.apple.com/en-us/HT210176](https://support.apple.com/en-us/HT210176)

* 许多系统要求证书中包含特定 OID，例如
  `id-kp-serverAuth`（TLS Web Server Authentication）
  缺失时，客户端会在收到数个 `Access-Challenge` 后静默重启 EAP 流程

* 所有系统都要求客户端安装根 CA，否则会出现与上述相同的失败现象

* Windows XP SP2 之后存在证书链处理缺陷
  如果服务器证书是中间证书而非根证书，认证会静默失败

* 某些 Windows CE 版本无法处理 4096 位 RSA 证书

* 在这些情况下，Windows 通常不会向用户显示任何明确错误信息，
  这常常导致错误地将问题归咎于 RADIUS 服务器

* 超过 **64KB** 的证书链已知无法正常工作
  大多数客户端无法处理如此大的证书链，
  AP 通常在约 50 次 EAP 往返后终止会话，而 64KB 证书链需要约 60 次往返

* 其他操作系统（Linux、BSD、macOS、iOS、Android、Solaris、Symbian 以及各类嵌入式系统）均已确认可正常工作

---

## 安全注意事项（Security Considerations）

默认的证书配置文件使用 **SHA256** 作为消息摘要算法，以确保安全性。

---


* 📘 精简成 **运维/部署版速查文档**
* 🧩 提炼成 **EAP-TLS / PEAP / TTLS 证书关系图**
* 🧠 结合你当前 **FreeRADIUS + Portal / 802.1X** 架构给出最佳实践配置
* 🛠️ 帮你把它整理成 **docs/certificates.md**

