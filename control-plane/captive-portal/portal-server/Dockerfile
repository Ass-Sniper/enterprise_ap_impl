FROM python:3.11-slim

WORKDIR /app

# ① Copy dependency manifest first (best cache usage)
COPY requirements.txt .

# ② Install dependencies (China mirror + BuildKit cache)
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install \
      -i https://pypi.tuna.tsinghua.edu.cn/simple \
      -r requirements.txt

COPY app ./app

EXPOSE 8080
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
