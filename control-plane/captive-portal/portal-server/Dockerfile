FROM python:3.11-slim

WORKDIR /app
COPY app ./app

RUN pip install fastapi uvicorn requests python-multipart

EXPOSE 8080
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
