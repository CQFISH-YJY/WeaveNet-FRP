#!/bin/bash
# WeaveNet 织网穿透 服务器构建部署脚本（Go 全栈轻量化版）
# 用法：cd deploy && ./build.sh
# 文档：CQFISH&喵酱出品
set -e
cd "$(dirname "$0")/.."

echo "[1/4] 同步前端产物与静态资源到 Go 面板目录"
rm -rf panel-go/web/dist panel-go/templates panel-go/static
cp -r panel/web/dist panel-go/web/dist
cp -r panel/templates panel-go/templates
cp -r panel/static panel-go/static

echo "[2/4] 构建并启动容器"
cd deploy
docker compose build
docker compose up -d

echo "[3/4] 清理旧容器与镜像（保留数据卷）"
docker container prune -f >/dev/null 2>&1 || true
docker image prune -f >/dev/null 2>&1 || true

echo "[4/4] 部署完成"
docker compose ps
