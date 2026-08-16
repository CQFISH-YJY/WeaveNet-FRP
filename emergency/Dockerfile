# WeaveNet 织网穿透 应急服务镜像（Go 版，单二进制）
# 多阶段构建：编译阶段 golang，运行阶段 alpine
# 文档：CQFISH&喵酱出品

FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/emergency .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates docker-cli
WORKDIR /app
COPY --from=build /out/emergency .

# 应急服务端口（仅内网/VPN）
EXPOSE 9001

ENTRYPOINT ["/app/emergency", "--port", "9001"]
