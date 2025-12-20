
## Ubuntu 20.04 安装 OpenSSL 3.5.0 及环境配置总结

### 1. 源码编译与安装

为了不破坏系统自带的 OpenSSL 1.1，采用隔离路径安装方案 ：

* 
**安装路径**：通过 `./config --prefix=/usr/local/openssl` 映射到独立目录 。


* 
**编译指令**：执行 `make -j$(nproc)` 并使用 `make install_sw` 完成安装 。



### 2. 关键：解决 "lib64" 路径陷阱

在 64 位系统上，OpenSSL 3.x 默认将库文件安装在 `lib64` 目录下，而非传统的 `lib`。

* **现象确认**：执行 `ls -F /usr/local/openssl/` 后，若看到 `lib64/` 文件夹，必须针对该路径进行配置。
* **错误示例**：若指向 `/usr/local/openssl/lib`，运行 `openssl version` 会报 `libssl.so.3: cannot open shared object file` 错误。

### 3. 环境变量持久化

为确保 `root` 用户及编译工具能正确识别新版 OpenSSL，需修改 `/root/.bashrc`：

```bash
# 编辑 root 配置文件
echo 'export OPENSSL_ROOT_DIR=/usr/local/openssl' >> /root/.bashrc
# 注意此处路径为 lib64
echo 'export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/openssl/lib64' >> /root/.bashrc
# 立即生效
source /root/.bashrc

```

同时为确保 `普通` 用户及编译工具能正确识别新版 OpenSSL，需修改 `~/.bashrc`：

```bash
# 编辑 root 配置文件
echo 'export OPENSSL_ROOT_DIR=/usr/local/openssl' >> ~/.bashrc
# 注意此处路径为 lib64
echo 'export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/openssl/lib64' >> ~/.bashrc
# 立即生效
source ~/.bashrc

```


### 4. 动态链接修复（ldconfig）

仅设置环境变量是不够的，必须让系统内核识别到新的共享库 ：

1. **创建配置文件**：
```bash
echo "/usr/local/openssl/lib64" | sudo tee /etc/ld.so.conf.d/openssl-3.5.conf

```


2. **刷新缓存**：执行 `sudo ldconfig`。
3. **验证链接**：执行 `sudo ldconfig -v | grep libssl`，确认输出包含 `libssl.so.3 -> libssl.so.3`。
4. **最终测试**：执行 `/usr/local/openssl/bin/openssl version`，应正确显示 **OpenSSL 3.5.0 8 Apr 2025**。

```text
kay@kay-vm:restbed$ /usr/local/openssl/bin/openssl version
/usr/local/openssl/bin/openssl: error while loading shared libraries: libssl.so.3: cannot open shared object file: No such file or directory
kay@kay-vm:restbed$
kay@kay-vm:restbed$
kay@kay-vm:restbed$
kay@kay-vm:restbed$ ls -F /usr/local/openssl/
bin/  include/  lib64/
kay@kay-vm:restbed$
kay@kay-vm:restbed$
kay@kay-vm:restbed$ echo "/usr/local/openssl/lib64" | sudo tee /etc/ld.so.conf.d/openssl-3.5.conf
/usr/local/openssl/lib64
kay@kay-vm:restbed$
kay@kay-vm:restbed$
kay@kay-vm:restbed$
kay@kay-vm:restbed$ sudo ldconfig -v | grep libssl
/sbin/ldconfig.real: Can't stat /usr/local/lib/x86_64-linux-gnu: No such file or directory
/sbin/ldconfig.real: Path `/usr/lib/x86_64-linux-gnu' given more than once
/sbin/ldconfig.real: Path `/usr/lib32' given more than once
/sbin/ldconfig.real: Path `/usr/libx32' given more than once
/sbin/ldconfig.real: Path `/lib/x86_64-linux-gnu' given more than once
/sbin/ldconfig.real: Path `/usr/lib/x86_64-linux-gnu' given more than once
/sbin/ldconfig.real: Path `/usr/lib' given more than once
        libssl.so.3 -> libssl.so.3
        libssl.so.1.1 -> libssl.so.1.1
        libssl3.so -> libssl3.so
/sbin/ldconfig.real: /lib/x86_64-linux-gnu/ld-2.31.so is the dynamic linker, ignoring

/sbin/ldconfig.real: /lib32/ld-2.31.so is the dynamic linker, ignoring

kay@kay-vm:restbed$
kay@kay-vm:restbed$
kay@kay-vm:restbed$ /usr/local/openssl/bin/openssl version
OpenSSL 3.5.0 8 Apr 2025 (Library: OpenSSL 3.5.0 8 Apr 2025)
kay@kay-vm:restbed$
```

### 5. 在 Dockerfile 中的整合建议

在您的 `Dockerfile` 中，应同步修正路径以确保镜像构建成功：

```dockerfile
# ------------------------------------------------------------
# Restbed 构建阶段建议
# ------------------------------------------------------------
WORKDIR /opt/restbed

# 使用 -S . [cite_start]-B build 模式，并强制指向正确的 OpenSSL 根目录 [cite: 3]
RUN cmake -S . -B build \
    -DBUILD_SSL=ON \
    -DBUILD_TESTS=OFF \
    -DBUILD_EXAMPLES=OFF \
    -DOPENSSL_ROOT_DIR=/usr/local/openssl \
 && cmake --build build -j$(nproc) \
 && cmake --install build

# ------------------------------------------------------------
# [cite_start]运行时环境配置 [cite: 4]
# ------------------------------------------------------------
ENV LD_LIBRARY_PATH=/usr/local/openssl/lib64:/usr/local/lib:/usr/local/lib64:$LD_LIBRARY_PATH

RUN echo "/usr/local/openssl/lib64" > /etc/ld.so.conf.d/openssl-3.5.conf \
 && ldconfig

```

---

### 下一步操作建议

既然 OpenSSL 3.5.0 已经配置成功，您可以继续执行 Restbed 的编译。如果在编译 `portal-server-cpp` 时遇到找不到 `restbed` 库的问题，请确保在 `cmake` 时同样传入 `-DOPENSSL_ROOT_DIR=/usr/local/openssl` 参数 。
