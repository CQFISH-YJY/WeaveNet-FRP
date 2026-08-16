# WeaveNet 织网穿透 面板镜像（Go 版，单二进制）
# 多阶段构建：编译阶段 golang，运行阶段 alpine
# 文档：CQFISH&喵酱出品

FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/weavenet-panel .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/weavenet-panel .
COPY templates ./templates
COPY static ./static
COPY web/dist ./web/dist

# 数据与日志目录
RUN mkdir -p /app/data /app/logs

# 暴露面板端口
EXPOSE 8000

CMD ["/app/weavenet-panel"]
