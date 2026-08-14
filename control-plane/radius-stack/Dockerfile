FROM freeradius/freeradius-server:latest-3.2-alpine

# ------------------------------------------------------------
# Alpine 国内源
# ------------------------------------------------------------
RUN sed -i 's|https://dl-cdn.alpinelinux.org|https://mirrors.tuna.tsinghua.edu.cn|g' \
    /etc/apk/repositories

# ------------------------------------------------------------
# SQL runtime dependency（MySQL / MariaDB）
# ------------------------------------------------------------
RUN apk add --no-cache mariadb-connector-c freeradius-utils

# ------------------------------------------------------------
# overlay：最小、确定、无隐式 include
# ------------------------------------------------------------

# clients
COPY freeradius/raddb/clients.conf \
     /opt/etc/raddb/clients.conf

# Lab PEAP CA and server certificate.  These replace the expired example
# certificate shipped in the upstream image.
COPY freeradius/raddb/certs/ \
     /opt/etc/raddb/certs/

# SQL 模块（已验证可用）
COPY freeradius/raddb/mods-enabled/sql \
     /opt/etc/raddb/mods-enabled/sql

# 802.1X / EAP（PEAP-MSCHAPv2、EAP-TLS 等）
COPY freeradius/raddb/mods-enabled/eap \
     /opt/etc/raddb/mods-enabled/eap

# 标准服务同时处理 PAP/SQL 与 802.1X；避免 portal 和 default
# 同时绑定 UDP 1812/1813。
COPY freeradius/raddb/sites-enabled/default \
     /opt/etc/raddb/sites-enabled/default
COPY freeradius/raddb/sites-enabled/inner-tunnel \
     /opt/etc/raddb/sites-enabled/inner-tunnel

# 基础镜像可能残留同端口的 portal server；标准 default 已兼容
# PAP/SQL，删除 portal 以避免 1812/1813 重复绑定。
RUN rm -f /opt/etc/raddb/sites-enabled/portal

# ------------------------------------------------------------
# 权限修正
# ------------------------------------------------------------
RUN find /opt/etc/raddb -type f -exec chmod 644 {} \; \
 && find /opt/etc/raddb -type d -exec chmod 755 {} \;
