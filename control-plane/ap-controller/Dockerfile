FROM python:3.11-slim

WORKDIR /app

# ① Copy dependency manifest first (best cache usage)
COPY requirements.txt .

# ② Install dependencies (China mirror + BuildKit cache)
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install \
      -i https://pypi.tuna.tsinghua.edu.cn/simple \
      -r requirements.txt

# ③ Copy application code
COPY app ./app

# ✅ Ensure config directory exists
RUN mkdir -p /app/config

# ④ Run
EXPOSE 8443
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8443"]