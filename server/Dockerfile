# WeaveNet 织网穿透 frps 镜像
FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY . .
RUN go build -o /out/frps ./cmd/frps && go build -o /out/frpc ./cmd/frpc

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /out/frps /usr/local/bin/frps
COPY --from=builder /out/frpc /usr/local/bin/frpc

# frps 控制端口 / HTTP / HTTPS 隧道端口
EXPOSE 7000 80 443

CMD ["frps", "-c", "/app/frps.toml"]
