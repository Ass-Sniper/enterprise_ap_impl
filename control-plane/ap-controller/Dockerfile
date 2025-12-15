FROM python:3.11-slim

WORKDIR /app

# ① 先拷贝依赖文件（单独一层，最大缓存收益）
COPY requirements.txt .

# ② 安装依赖（国内 PyPI + BuildKit cache）
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install \
      -i https://pypi.tuna.tsinghua.edu.cn/simple \
      -r requirements.txt

# ③ 再拷贝业务代码（改代码不会触发 pip install）
COPY app ./app

# ④ 启动
EXPOSE 8443
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8443"]
