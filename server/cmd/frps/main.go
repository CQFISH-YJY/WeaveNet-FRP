package main

// WeaveNet 织网穿透 frps 服务端入口。
//
// frps 是内网穿透的核心服务端，承担以下职责：
//   - 控制通道 API：向面板注册节点、心跳上报（每 30s）
//   - 隧道配置热拉取：每 10s 从面板拉取本节点隧道与用户限速配置
//   - 客户端鉴权：frpc 连接时校验面板签发的用户 Token
//   - 限速联动：按用户套餐带宽上限动态限速（令牌桶）
//   - 状态上报：在线隧道、连接数、流量增量
//   - 远程端口分配校验：防止端口冲突
//   - 数据转发：TCP / UDP / HTTP / HTTPS 隧道转发
//
// 协议说明：本内核实现轻量自研协议（JSON 行帧），与配套 frpc 通讯；
// 面板联动走 HTTP JSON API。生产环境可无缝替换为基于 fatedier/frp
// 的完整 fork（Agent API 契约一致）。

import (
	"flag"
	"log"
	"os"

	"github.com/weavenet/frp-server/internal/agent"
	"github.com/weavenet/frp-server/internal/config"
	"github.com/weavenet/frp-server/internal/server"
)

func main() {
	configPath := flag.String("c", "frps.toml", "配置文件路径")
	flag.Parse()

	cfg, err := config.LoadServerConfig(*configPath)
	if err != nil {
		log.Fatalf("[frps] 配置加载失败: %v", err)
	}

	logger := log.New(os.Stdout, "[frps] ", log.LstdFlags)
	logger.Printf("WeaveNet frps 启动，节点 %s，面板 %s，控制端口 %d",
		cfg.NodeName, cfg.PanelBaseURL, cfg.ControlPort)

	// 启动面板联动 Agent
	ag, err := agent.NewAgent(cfg, logger)
	if err != nil {
		log.Fatalf("[frps] Agent 初始化失败: %v", err)
	}
	ag.Start()

	// 启动 frps 转发服务
	srv := server.NewServer(cfg, ag, logger)
	if err := srv.Run(); err != nil {
		log.Fatalf("[frps] 服务异常退出: %v", err)
	}
}
