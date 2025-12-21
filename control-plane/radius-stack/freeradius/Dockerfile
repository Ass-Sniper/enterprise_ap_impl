
FROM freeradius/freeradius-server:latest-3.2-alpine

# ------------------------------------------------------------
# Alpine 国内源加速
# ------------------------------------------------------------
RUN sed -i 's|https://dl-cdn.alpinelinux.org|https://mirrors.tuna.tsinghua.edu.cn|g' /etc/apk/repositories

# ------------------------------------------------------------
# SQL runtime dependency (MySQL / MariaDB)
# ------------------------------------------------------------
RUN apk add --no-cache mariadb-connector-c

# ------------------------------------------------------------
# Runtime raddb
# ------------------------------------------------------------
COPY raddb /opt/etc/raddb

# ---- 关键裁剪：删除不用的模块 ----
RUN rm -f /opt/etc/raddb/mods-enabled/eap \
 && rm -f /opt/etc/raddb/sites-enabled/inner-tunnel \
 && rm -f /opt/etc/raddb/sites-enabled/clients.conf

# ------------------------------------------------------------
# 修正 FreeRADIUS 配置文件权限（非常关键）
# ------------------------------------------------------------
RUN find /opt/etc/raddb -type f -exec chmod 644 {} \; \
 && find /opt/etc/raddb -type d -exec chmod 755 {} \;
