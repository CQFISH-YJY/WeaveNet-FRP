// Package config 负责 frps / frpc 配置加载。
package config

import (
	"os"
)

// LoadServerConfig 加载 frps 配置。
//
// 采用最简键值格式解析（key = value，支持 # 注释），
// 兼容 frp 官方 toml 的子集，避免引入第三方依赖。
func LoadServerConfig(path string) (*ServerConfig, error) {
	cfg := &ServerConfig{}
	kv, err := parseSimpleConfig(path)
	if err != nil {
		return nil, err
	}
	cfg.NodeName = kv["nodeName"]
	cfg.AgentToken = kv["agentToken"]
	cfg.PanelBaseURL = kv["panelBaseURL"]
	cfg.ControlPort = atoi(kv["bindPort"], 7000)
	cfg.HTTPPort = atoi(kv["vhostHTTPPort"], 80)
	cfg.HTTPSPort = atoi(kv["vhostHTTPSPort"], 443)
	cfg.Address = kv["bindAddr"]
	if cfg.Address == "" {
		cfg.Address = "0.0.0.0"
	}
	return cfg, nil
}

// LoadClientConfig 加载 frpc 配置。
func LoadClientConfig(path string) (*ClientConfig, error) {
	cfg := &ClientConfig{}
	kv, err := parseSimpleConfig(path)
	if err != nil {
		return nil, err
	}
	cfg.ServerAddr = kv["serverAddr"]
	cfg.ServerPort = atoi(kv["serverPort"], 7000)
	cfg.Token = kv["auth.token"]
	cfg.ProxyName = kv["name"]
	cfg.ProxyType = kv["type"]
	cfg.LocalIP = kv["localIP"]
	if cfg.LocalIP == "" {
		cfg.LocalIP = "127.0.0.1"
	}
	cfg.LocalPort = atoi(kv["localPort"], 0)
	cfg.RemotePort = atoi(kv["remotePort"], 0)
	cfg.Subdomain = kv["subdomain"]
	return cfg, nil
}

// parseSimpleConfig 解析最简 key = value 配置文件。
func parseSimpleConfig(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	kv := map[string]string{}
	lines := splitLines(string(data))
	for _, line := range lines {
		line = trimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		eq := indexOf(line, '=')
		if eq < 0 {
			continue
		}
		key := trimSpace(line[:eq])
		val := trimSpace(line[eq+1:])
		val = trimQuotes(val)
		if key != "" {
			kv[key] = val
		}
	}
	return kv, nil
}

// ServerConfig frps 服务端配置。
type ServerConfig struct {
	NodeName     string // 节点名称（与面板一致）
	AgentToken   string // 面板签发的 Agent Token
	PanelBaseURL string // 面板地址，如 http://127.0.0.1:8000
	Address      string // 监听地址
	ControlPort  int    // 控制端口（frpc 连接）
	HTTPPort     int    // HTTP 隧道端口
	HTTPSPort    int    // HTTPS 隧道端口
}

// ClientConfig frpc 客户端配置。
type ClientConfig struct {
	ServerAddr string
	ServerPort int
	Token      string // 用户 Token（面板签发）
	ProxyName  string
	ProxyType  string
	LocalIP    string
	LocalPort  int
	RemotePort int
	Subdomain  string
}
