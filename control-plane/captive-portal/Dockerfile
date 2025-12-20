############################################
# build args
############################################
ARG PORTAL_IMPL=cpp
ARG RESTBED_VERSION=4.8
ARG OPENSSL_VERSION=openssl-3.5.0
ARG ASIO_VERSION=asio-1-34-2

############################################
# ========== C++ Builder ==========
############################################
FROM ubuntu:22.04 AS cpp-builder

ENV DEBIAN_FRONTEND=noninteractive

# ------------------------------------------------------------
# 基础构建依赖（APT cache 加速）
# ------------------------------------------------------------
RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt \
    apt-get update \
 && apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    git \
    perl \
    autoconf \
    automake \
    libtool \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# ------------------------------------------------------------
# Build OpenSSL 3.5.0
# ------------------------------------------------------------
WORKDIR /opt

RUN git clone --depth 1 --branch ${OPENSSL_VERSION} \
    https://github.com/openssl/openssl.git

WORKDIR /opt/openssl

RUN ./config \
      --prefix=/usr/local/openssl \
      --openssldir=/usr/local/openssl \
      shared \
 && make -j$(nproc) \
 && make install_sw

ENV OPENSSL_ROOT_DIR=/usr/local/openssl
ENV PATH=/usr/local/openssl/bin:$PATH
ENV LD_LIBRARY_PATH=/usr/local/openssl/lib:$LD_LIBRARY_PATH

# ------------------------------------------------------------
# Build Asio (standalone)
# ------------------------------------------------------------
WORKDIR /opt

RUN git clone --depth 1 --branch ${ASIO_VERSION} \
    https://github.com/chriskohlhoff/asio.git

WORKDIR /opt/asio/asio

RUN ./autogen.sh \
 && ./configure \
 && make -j$(nproc) \
 && make install

# ------------------------------------------------------------
# build restbed
# ------------------------------------------------------------
WORKDIR /opt

RUN git clone --depth 1 --branch ${RESTBED_VERSION} \
    https://github.com/Corvusoft/restbed.git

WORKDIR /opt/restbed

RUN cmake -S . -B build \
    -DBUILD_SSL=ON \
    -DBUILD_TESTS=OFF \
    -DBUILD_EXAMPLES=OFF \
    -DOPENSSL_ROOT_DIR=/usr/local/openssl \
 && cmake --build build -j$(nproc) \
 && cmake --install build

# ------------------------------------------------------------
# build portal-server-cpp
# ------------------------------------------------------------
WORKDIR /build
COPY portal-server-cpp /build

RUN cmake -S . -B build \
      -DCMAKE_BUILD_TYPE=Release \
      -DOPENSSL_ROOT_DIR=/usr/local/openssl \
 && cmake --build build -j$(nproc)

############################################
# ========== C++ Runtime ==========
############################################
FROM ubuntu:22.04 AS cpp-runtime

ENV DEBIAN_FRONTEND=noninteractive

# ------------------------------------------------------------
# 运行时最小依赖（APT cache 加速）
# ------------------------------------------------------------
RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt \
    apt-get update \
 && apt-get install -y --no-install-recommends \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# OpenSSL 3.5 + Asio headers + Restbed + portal
COPY --from=cpp-builder /usr/local /usr/local
COPY --from=cpp-builder /build/build/portal-server /app/portal-server

ENV LD_LIBRARY_PATH=/usr/local/openssl/lib:/usr/local/lib:/usr/local/lib64

RUN echo "/usr/local/openssl/lib" > /etc/ld.so.conf.d/openssl-3.5.conf \
 && ldconfig

EXPOSE 8080
CMD ["/app/portal-server"]

############################################
# ========== Python Runtime ==========
############################################
FROM python:3.11-slim AS python-runtime

WORKDIR /app

COPY portal-server/requirements.txt .

RUN pip install --no-cache-dir \
    -i https://pypi.tuna.tsinghua.edu.cn/simple \
    --default-timeout=100 \
    -r requirements.txt

COPY portal-server /app/portal-server

EXPOSE 8080
CMD ["uvicorn", "portal-server.app.main:app", "--host", "0.0.0.0", "--port", "8080"]

############################################
# ========== Final Selector ==========
############################################
FROM ${PORTAL_IMPL}-runtime AS final
