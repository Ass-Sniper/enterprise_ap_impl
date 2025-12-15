#!/usr/bin/env bash
set -e

ROOT="control-plane"

echo "==> Generating control-plane files..."

############################
# ap-controller
############################

mkdir -p $ROOT/ap-controller/app

cat > $ROOT/ap-controller/app/main.py << 'EOF'
from fastapi import FastAPI

app = FastAPI(title="AP Controller")

@app.get("/")
def health():
    return {"status": "ap-controller ok"}

@app.post("/portal/auth")
def portal_auth(req: dict):
    # Demo：始终放行
    return {
        "result": "ok",
        "role": "guest",
        "ttl": 3600
    }
EOF

cat > $ROOT/ap-controller/Dockerfile << 'EOF'
FROM python:3.11-slim

WORKDIR /app
COPY app ./app

RUN pip install fastapi uvicorn

EXPOSE 8443
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8443"]
EOF

cat > $ROOT/ap-controller/README.md << 'EOF'
# AP Controller

控制面核心服务：
- AP 注册 / 心跳（后续）
- Portal 认证决策
- 策略下发

运行方式由 control-plane/docker-compose.yml 统一管理。
EOF

############################
# captive-portal / portal-server
############################

mkdir -p $ROOT/captive-portal/portal-server/app

cat > $ROOT/captive-portal/portal-server/app/main.py << 'EOF'
from fastapi import FastAPI, Form
import requests

CONTROLLER_URL = "http://ap-controller:8443"

app = FastAPI(title="Captive Portal")

@app.get("/")
def portal_page():
    return """
    <html>
      <body>
        <h3>Captive Portal</h3>
        <form method="post" action="/login">
          <input name="mac" placeholder="Client MAC" />
          <button type="submit">Login</button>
        </form>
      </body>
    </html>
    """

@app.post("/login")
def login(mac: str = Form(...)):
    r = requests.post(
        f"{CONTROLLER_URL}/portal/auth",
        json={"mac": mac, "ip": "0.0.0.0"}
    )
    return r.json()
EOF

cat > $ROOT/captive-portal/portal-server/Dockerfile << 'EOF'
FROM python:3.11-slim

WORKDIR /app
COPY app ./app

RUN pip install fastapi uvicorn requests python-multipart

EXPOSE 8080
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
EOF

cat > $ROOT/captive-portal/README.md << 'EOF'
# Captive Portal

负责：
- Portal 页面
- 用户认证交互
- 调用 AP Controller 获取授权结果

portal-agent 目录预留给 AP 侧集成使用。
EOF

############################
# control-plane docker-compose
############################

cat > $ROOT/docker-compose.yml << 'EOF'
version: "3.9"

services:
  ap-controller:
    build: ./ap-controller
    container_name: ap-controller
    ports:
      - "8443:8443"

  captive-portal:
    build: ./captive-portal/portal-server
    container_name: captive-portal
    ports:
      - "8080:8080"
    depends_on:
      - ap-controller
EOF

echo "==> All files generated successfully."

