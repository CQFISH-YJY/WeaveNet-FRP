// Package agent 实现 frps 与面板的控制通道联动。
//
// 对应设计文档 8.3 内核联动 API：
//   - POST /api/agent/register 节点注册
//   - POST /api/agent/heartbeat 心跳上报（状态/流量）
//   - GET  /api/agent/tunnels  拉取本节点隧道+限速配置
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/weavenet/frp-server/internal/config"
)

// TunnelConfig 面板下发的隧道配置。
type TunnelConfig struct {
	TunnelID           int             `json:"tunnel_id"`
	UserToken          string          `json:"user_token"`
	Username           string          `json:"username"`
	Type               string          `json:"type"`
	Name               string          `json:"name"`
	LocalIP            string          `json:"local_ip"`
	LocalPort          int             `json:"local_port"`
	RemotePort         int             `json:"remote_port"`
	Subdomain          string          `json:"subdomain"`
	CustomDomain       string          `json:"custom_domain"`
	KCP                bool            `json:"kcp"`
	Encryption         bool            `json:"encryption"`
	Compression        bool            `json:"compression"`
	SecretKey          string          `json:"secret_key"`
	BandwidthLimitKbps int             `json:"bandwidth_limit_kbps"`
	LoadBalancers      json.RawMessage `json:"load_balancers"`
}

// SpeedLimit 面板下发的用户限速配置。
type SpeedLimit struct {
	UserToken          string `json:"user_token"`
	Username           string `json:"username"`
	BandwidthLimitKbps int    `json:"bandwidth_limit_kbps"`
	Status             string `json:"status"`
}

// DomainRoute 域名路由表。
type DomainRoute struct {
	FullDomain string `json:"full_domain"`
	Subdomain  string `json:"subdomain"`
	TunnelID   int    `json:"tunnel_id"`
}

// PullResult 配置拉取结果。
type PullResult struct {
	Tunnels     []TunnelConfig `json:"tunnels"`
	SpeedLimits []SpeedLimit   `json:"speed_limits"`
	Domains     []DomainRoute  `json:"domains"`
}

// Agent 面板联动 Agent。
type Agent struct {
	cfg    *config.ServerConfig
	logger *log.Logger
	client *http.Client

	mu         sync.RWMutex
	tunnels    map[int]TunnelConfig
	limits     map[string]SpeedLimit // key: userToken
	domains    map[string]DomainRoute
	stopCh     chan struct{}
	registered bool
	// 统计缓存
	statsMu sync.Mutex
	stats   map[int]*TrafficCounters
}

// TrafficCounters 隧道流量计数器。
type TrafficCounters struct {
	InBytes     int64 `json:"in"`
	OutBytes    int64 `json:"out"`
	Connections int64 `json:"connections"`
	Online      bool  `json:"online"`
}

// NewAgent 创建 Agent。
func NewAgent(cfg *config.ServerConfig, logger *log.Logger) (*Agent, error) {
	if cfg.PanelBaseURL == "" {
		return nil, fmt.Errorf("面板地址未配置（panelBaseURL）")
	}
	if cfg.AgentToken == "" {
		return nil, fmt.Errorf("Agent Token 未配置（agentToken）")
	}
	return &Agent{
		cfg:     cfg,
		logger:  logger,
		client:  &http.Client{Timeout: 10 * time.Second},
		tunnels: make(map[int]TunnelConfig),
		limits:  make(map[string]SpeedLimit),
		domains: make(map[string]DomainRoute),
		stopCh:  make(chan struct{}),
		stats:   make(map[int]*TrafficCounters),
	}, nil
}

// Start 启动注册与轮询任务。
func (a *Agent) Start() {
	go a.loop()
}

// Stop 停止 Agent。
func (a *Agent) Stop() {
	close(a.stopCh)
}

func (a *Agent) loop() {
	// 启动时立即注册并拉取一次
	a.register()
	a.pullTunnels()

	// 心跳 30s，配置拉取 10s
	heartbeatTicker := time.NewTicker(30 * time.Second)
	pullTicker := time.NewTicker(10 * time.Second)
	defer heartbeatTicker.Stop()
	defer pullTicker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-heartbeatTicker.C:
			a.register()
			a.heartbeat()
		case <-pullTicker.C:
			a.pullTunnels()
		}
	}
}

// register 向面板注册节点。
func (a *Agent) register() {
	payload := map[string]any{
		"agent_token": a.cfg.AgentToken,
		"name":        a.cfg.NodeName,
		"address":     a.cfg.Address,
		"port":        a.cfg.ControlPort,
	}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			NodeID int `json:"node_id"`
		} `json:"data"`
	}
	if err := a.postJSON("/api/agent/register", payload, &resp); err != nil {
		a.logger.Printf("节点注册失败: %v", err)
		return
	}
	a.registered = resp.Code == 0
}

// heartbeat 上报隧道状态与流量。
func (a *Agent) heartbeat() {
	a.statsMu.Lock()
	snapshot := make([]map[string]any, 0, len(a.stats))
	for id, s := range a.stats {
		snapshot = append(snapshot, map[string]any{
			"tunnel_id":  id,
			"online":     s.Online,
			"connections": s.Connections,
			"in_delta":   s.InBytes,
			"out_delta":  s.OutBytes,
		})
		// 上报后清零增量
		s.InBytes = 0
		s.OutBytes = 0
	}
	a.statsMu.Unlock()

	payload := map[string]any{
		"tunnels": snapshot,
	}
	var resp struct {
		Code int `json:"code"`
	}
	if err := a.postJSON("/api/agent/heartbeat", payload, &resp); err != nil {
		a.logger.Printf("心跳上报失败: %v", err)
	}
}

// pullTunnels 从面板拉取本节点隧道与限速配置。
func (a *Agent) pullTunnels() {
	var result PullResult
	if err := a.getJSON("/api/agent/tunnels", &result); err != nil {
		a.logger.Printf("隧道配置拉取失败: %v", err)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// 隧道配置
	newTunnels := make(map[int]TunnelConfig, len(result.Tunnels))
	for _, t := range result.Tunnels {
		newTunnels[t.TunnelID] = t
	}
	a.tunnels = newTunnels

	// 限速配置
	newLimits := make(map[string]SpeedLimit, len(result.SpeedLimits))
	for _, l := range result.SpeedLimits {
		newLimits[l.UserToken] = l
	}
	a.limits = newLimits

	// 域名路由
	newDomains := make(map[string]DomainRoute, len(result.Domains))
	for _, d := range result.Domains {
		newDomains[d.FullDomain] = d
	}
	a.domains = newDomains

	a.logger.Printf("配置已热更新: %d 条隧道, %d 个限速, %d 条域名路由",
		len(a.tunnels), len(a.limits), len(a.domains))
}

// GetTunnels 返回当前隧道配置快照。
func (a *Agent) GetTunnels() []TunnelConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]TunnelConfig, 0, len(a.tunnels))
	for _, t := range a.tunnels {
		out = append(out, t)
	}
	return out
}

// GetTunnel 返回指定隧道配置。
func (a *Agent) GetTunnel(id int) (TunnelConfig, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	t, ok := a.tunnels[id]
	return t, ok
}

// GetBandwidthLimit 返回用户 Token 对应的限速（Kbps），默认 8000。
func (a *Agent) GetBandwidthLimit(userToken string) int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if l, ok := a.limits[userToken]; ok {
		return l.BandwidthLimitKbps
	}
	return 8000
}

// GetDomainRoute 返回完整域名对应的隧道 ID。
func (a *Agent) GetDomainRoute(host string) (int, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if d, ok := a.domains[host]; ok {
		return d.TunnelID, true
	}
	return 0, false
}

// RecordTraffic 记录隧道流量与状态。
func (a *Agent) RecordTraffic(tunnelID int, in, out int64) {
	a.statsMu.Lock()
	defer a.statsMu.Unlock()
	s := a.stats[tunnelID]
	if s == nil {
		s = &TrafficCounters{Online: true}
		a.stats[tunnelID] = s
	}
	s.InBytes += in
	s.OutBytes += out
}

// SetTunnelOnline 设置隧道在线状态。
func (a *Agent) SetTunnelOnline(tunnelID int, online bool) {
	a.statsMu.Lock()
	defer a.statsMu.Unlock()
	s := a.stats[tunnelID]
	if s == nil {
		s = &TrafficCounters{Online: online}
		a.stats[tunnelID] = s
	}
	s.Online = online
}

// AddConnection 增加/减少隧道连接数。
func (a *Agent) AddConnection(tunnelID int, delta int64) {
	a.statsMu.Lock()
	defer a.statsMu.Unlock()
	s := a.stats[tunnelID]
	if s == nil {
		s = &TrafficCounters{Online: true}
		a.stats[tunnelID] = s
	}
	s.Connections += delta
	if s.Connections < 0 {
		s.Connections = 0
	}
}

func (a *Agent) postJSON(path string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, a.cfg.PanelBaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.cfg.AgentToken)
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out != nil {
		return json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(out)
	}
	return nil
}

func (a *Agent) getJSON(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, a.cfg.PanelBaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+a.cfg.AgentToken)
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(out)
}
