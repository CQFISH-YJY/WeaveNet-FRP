# WeaveNet 织网穿透 应急服务镜像
FROM python:3.12-slim

WORKDIR /app
COPY emergency.py .

# 应急服务端口（仅内网/VPN）
EXPOSE 9001

CMD ["python", "emergency.py", "--port", "9001"]
