FROM freeradius/freeradius-server:latest-3.2-alpine

# ------------------------------------------------------------
# Alpine 国内源
# ------------------------------------------------------------
RUN sed -i 's|https://dl-cdn.alpinelinux.org|https://mirrors.tuna.tsinghua.edu.cn|g' \
    /etc/apk/repositories

# ------------------------------------------------------------
# SQL runtime dependency（MySQL / MariaDB）
# ------------------------------------------------------------
RUN apk add --no-cache mariadb-connector-c

# ------------------------------------------------------------
# overlay：最小、确定、无隐式 include
# ------------------------------------------------------------

# clients
COPY freeradius/raddb/clients.conf \
     /opt/etc/raddb/clients.conf

# SQL 模块（已验证可用）
COPY freeradius/raddb/mods-enabled/sql \
     /opt/etc/raddb/mods-enabled/sql

# 👉 使用你自己的 server（例如 portal）
COPY freeradius/raddb/sites-enabled/portal \
     /opt/etc/raddb/sites-enabled/portal

# ------------------------------------------------------------
# 裁剪无关模块
# ------------------------------------------------------------
RUN rm -f /opt/etc/raddb/mods-enabled/eap \
 && rm -f /opt/etc/raddb/sites-enabled/inner-tunnel \
 && rm -f /opt/etc/raddb/sites-enabled/default

# ------------------------------------------------------------
# 权限修正
# ------------------------------------------------------------
RUN find /opt/etc/raddb -type f -exec chmod 644 {} \; \
 && find /opt/etc/raddb -type d -exec chmod 755 {} \;
