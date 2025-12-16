# 使用精简版 Python 镜像
FROM python:3.11-slim

# 关闭 .pyc 并启用无缓冲输出
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# 安装运行所需依赖
RUN pip install --no-cache-dir fastapi uvicorn requests jinja2

# 复制应用代码（包含 templates）
COPY portal-server /app/portal-server

# 暴露服务端口
EXPOSE 8080

# 启动应用
CMD ["uvicorn", "portal-server.app.main:app", "--host", "0.0.0.0", "--port", "8080"]
