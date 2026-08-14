# RM65 FreeRADIUS PEAP 证书生成与部署

本目录保存 RM65 WPA2-Enterprise（PEAP/MSCHAPv2）使用的 RADIUS CA 与服务器证书。

## 文件用途

| 文件 | 用途 | 是否可提交到 Git |
| --- | --- | --- |
| `ca.der`、`ca.pem` | 根 CA 公钥；导入 Windows、Android 或 iOS 终端信任库 | 可以 |
| `server.crt` | RADIUS 服务器公钥证书 | 可以 |
| `ca.key` | 根 CA 私钥 | 禁止 |
| `server.key` | RADIUS 服务器私钥 | 禁止 |
| `server.pem` | 加密服务器私钥与服务器证书的组合文件 | 禁止 |
| `server.csr`、`server-ext.cnf`、`ca.srl` | 生成过程临时文件 | 不建议 |

当前 `eap` 配置使用：

```text
private_key_file = ${certdir}/server.pem
certificate_file = ${certdir}/server.pem
ca_file          = ${cadir}/ca.pem
```

因此 `server.pem` 的私钥口令必须与 `private_key_password` 一致。生产环境应通过 Docker Secret 或只读挂载注入私钥，而不要把私钥打入镜像或提交 Git。

## 生成命令

以下命令在 Ubuntu 的控制面目录执行，要求 OpenSSL 1.1.1 或更新版本。

```bash
cd ~/codebase/enterprise_ap_impl/control-plane
export CERT_DIR="$PWD/radius-stack/freeradius/raddb/certs"
mkdir -p "$CERT_DIR"
chmod 700 "$CERT_DIR"

# 口令不会显示在终端，也不会写入 shell history。
read -rsp 'CA 私钥口令: ' CA_KEY_PASS; echo
read -rsp '服务器私钥口令: ' SERVER_KEY_PASS; echo
export CA_KEY_PASS SERVER_KEY_PASS
```

生成 10 年有效期的实验室根 CA：

```bash
openssl genrsa -aes256 -passout env:CA_KEY_PASS \
  -out "$CERT_DIR/ca.key" 4096

openssl req -x509 -new -sha256 -days 3650 \
  -key "$CERT_DIR/ca.key" -passin env:CA_KEY_PASS \
  -out "$CERT_DIR/ca.pem" \
  -subj '/C=CN/O=RM65 Lab/OU=WiFi/CN=RM65 Lab RADIUS CA'

openssl x509 -in "$CERT_DIR/ca.pem" -outform DER \
  -out "$CERT_DIR/ca.der"
```

生成带 DNS SAN 的 RADIUS 服务器证书。`radius.rm65.lab` 是 Android/iOS 终端 Wi-Fi Enterprise 配置中的“域 / Domain”；它不要求客户端能够 DNS 解析，只需与证书 SAN 匹配。

```bash
openssl genrsa -aes256 -passout env:SERVER_KEY_PASS \
  -out "$CERT_DIR/server.key" 3072

openssl req -new -sha256 \
  -key "$CERT_DIR/server.key" -passin env:SERVER_KEY_PASS \
  -out "$CERT_DIR/server.csr" \
  -subj '/C=CN/O=RM65 Lab/OU=WiFi/CN=radius.rm65.lab'

cat > "$CERT_DIR/server-ext.cnf" <<'EOF'
basicConstraints = critical, CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = DNS:radius.rm65.lab
EOF

openssl x509 -req -sha256 -days 825 \
  -in "$CERT_DIR/server.csr" \
  -CA "$CERT_DIR/ca.pem" -CAkey "$CERT_DIR/ca.key" \
  -passin env:CA_KEY_PASS -CAcreateserial \
  -extfile "$CERT_DIR/server-ext.cnf" \
  -out "$CERT_DIR/server.crt"

# FreeRADIUS 当前 eap 配置要求同一个 PEM 同时含加密私钥和证书。
cat "$CERT_DIR/server.key" "$CERT_DIR/server.crt" > "$CERT_DIR/server.pem"
chmod 600 "$CERT_DIR/ca.key" "$CERT_DIR/server.key" "$CERT_DIR/server.pem"
unset CA_KEY_PASS SERVER_KEY_PASS
```

## 验证

确认服务器证书链、私钥匹配与 SAN：

```bash
openssl verify -CAfile "$CERT_DIR/ca.pem" "$CERT_DIR/server.crt"
openssl x509 -in "$CERT_DIR/server.crt" -noout -subject -issuer -dates -ext subjectAltName
openssl x509 -noout -modulus -in "$CERT_DIR/server.crt" | openssl sha256
openssl rsa  -noout -modulus -in "$CERT_DIR/server.key" -passin pass:'<服务器私钥口令>' | openssl sha256
```

最后两行摘要必须相同。口令只用于本机临时验证；执行后建议清除终端历史或改用 `-passin env:SERVER_KEY_PASS`。

## 构建、部署与回归测试

当前开发镜像会从 `certs/` 复制证书材料：

```bash
cd ~/codebase/enterprise_ap_impl/control-plane
docker compose build freeradius
docker compose up -d --force-recreate freeradius
docker compose logs -f freeradius
```

测试内层 MSCHAPv2：

```bash
docker compose exec -T freeradius \
  radtest -t mschap testuser '<测试账号密码>' 127.0.0.1:18120 0 '<RADIUS NAS 密钥>'
```

预期为 `Access-Accept`，且应返回 `MS-CHAP-MPPE-Keys`。

## 手机/电脑终端

- 导入 `ca.der`（Android 推荐）或 `ca.pem`（按系统支持情况）。
- Wi-Fi 安全类型：WPA2-Enterprise；EAP：PEAP；第二阶段：MSCHAPv2。
- CA / 域：选择该 CA，并填写 `radius.rm65.lab`。
- 不要选择“不要验证证书”。

## Git 安全检查

提交前执行：

```bash
git diff --cached --name-only | grep -E '(^|/)(ca\.key|server\.key|server\.pem)$' \
  && { echo '错误：检测到私钥，停止提交'; exit 1; } \
  || true
```
