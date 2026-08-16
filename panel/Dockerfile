# WeaveNet 织网穿透 面板镜像
FROM python:3.12-slim

WORKDIR /app

# 安装依赖
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 复制应用与前端产物
COPY app ./app
COPY templates ./templates
COPY static ./static
COPY web/dist ./web/dist

# 数据目录
RUN mkdir -p /app/data /app/logs

# 暴露面板端口
EXPOSE 8000

# 启动（uvicorn 单进程）
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
