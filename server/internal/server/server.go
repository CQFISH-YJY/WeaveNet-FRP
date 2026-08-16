// Package server 实现 frps 转发服务核心。
//
// 架构：
//   - 控制端口：frpc 鉴权连接 + 隧道代理连接池
//   - 隧道端口：TCP/UDP 隧道按 remote_port 监听
//   - HTTP/HTTPS 隧道：vhostHTTPPort/vhostHTTPSPort 按 Host 路由
//   - 限速：按用户 Token 令牌桶限速
package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/weavenet/frp-server/internal/agent"
	"github.com/weavenet/frp-server/internal/config"
)

// 协议帧类型。
const (
	frameAuth       = "auth"
	frameAuthResult = "auth_result"
	frameProxy      = "proxy"
	frameProxyOK    = "proxy_ok"
	frameProxyErr   = "proxy_err"
	frameData       = "data"
)

// Frame 协议帧。
type Frame struct {
	Type     string `json:"type"`
	Token    string `json:"token,omitempty"`
	TunnelID int    `json:"tunnel_id,omitempty"`
	OK       bool   `json:"ok,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Tunnel 运行中的隧道。
type Tunnel struct {
	ID       int
	Type     string
	Name     string
	RemotePort int
	Subdomain  string
	UserToken  string
	Kbps       int
	loadBalancers []agent.TunnelConfig

	mu       sync.Mutex
	idle     []net.Conn // 空闲代理连接池
	active   int        // 活跃代理连接数
	stopping bool
}

// Server frps 服务端。
type Server struct {
	cfg    *config.ServerConfig
	agent  *agent.Agent
	logger *log.Logger

	mu      sync.RWMutex
	tunnels map[int]*Tunnel

	limiter *RateLimiter
}

// NewServer 创建服务端。
func NewServer(cfg *config.ServerConfig, ag *agent.Agent, logger *log.Logger) *Server {
	return &Server{
		cfg:     cfg,
		agent:   ag,
		logger:  logger,
		tunnels: make(map[int]*Tunnel),
		limiter: NewRateLimiter(),
	}
}

// Run 启动 frps。
func (s *Server) Run() error {
	// 监听控制端口
	addr := fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.ControlPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("控制端口监听失败: %w", err)
	}
	s.logger.Printf("控制端口监听 %s", addr)

	// 启动隧道端口监听（动态注册）
	go s.watchTunnels()

	// 启动 HTTP vhost
	if s.cfg.HTTPPort > 0 {
		httpAddr := fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.HTTPPort)
		go s.serveHTTPVhost(httpAddr, "http")
	}

	s.logger.Printf("frps 就绪，等待 frpc 连接")
	for {
		conn, err := ln.Accept()
		if err != nil {
			s.logger.Printf("控制连接接受失败: %v", err)
			continue
		}
		go s.handleControlConn(conn)
	}
}

// watchTunnels 每 3s 对比面板配置，动态启停隧道监听。
func (s *Server) watchTunnels() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		desired := map[int]bool{}
		for _, t := range s.agent.GetTunnels() {
			desired[t.TunnelID] = true
			s.ensureTunnel(t)
		}
		// 下线不再需要的隧道
		s.mu.Lock()
		for id, t := range s.tunnels {
			if !desired[id] {
				t.stopping = true
				delete(s.tunnels, id)
				s.logger.Printf("隧道 %d 已从面板移除，停止监听", id)
			}
		}
		s.mu.Unlock()
	}
}

// ensureTunnel 确保隧道已创建并监听。
func (s *Server) ensureTunnel(tc agent.TunnelConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tunnels[tc.TunnelID]; ok {
		return
	}
	t := &Tunnel{
		ID:         tc.TunnelID,
		Type:       tc.Type,
		Name:       tc.Name,
		RemotePort: tc.RemotePort,
		Subdomain:  tc.Subdomain,
		UserToken:  tc.UserToken,
		Kbps:       tc.BandwidthLimitKbps,
	}
	s.tunnels[t.ID] = t

	// 注册限速
	s.limiter.SetLimit(tc.UserToken, tc.BandwidthLimitKbps)

	switch tc.Type {
	case "tcp", "udp", "kcp":
		if t.RemotePort > 0 {
			go s.listenTunnelPort(t, tc.Type)
		}
	case "http", "https":
		// 由 vhost 统一处理，无需独立监听
		s.logger.Printf("HTTP 隧道 %d 由 vhost 处理", t.ID)
	case "stcp", "xtcp":
		// 安全隧道：由 frpc 建立虚拟端口（简化实现，走 tcp 转发）
		if t.RemotePort > 0 {
			go s.listenTunnelPort(t, "tcp")
		}
	case "loadbalance":
		if t.RemotePort > 0 {
			go s.listenTunnelPort(t, "tcp")
		}
	default:
		s.logger.Printf("未知隧道类型 %s，忽略", tc.Type)
	}
	s.logger.Printf("隧道 %d (%s) 已就绪", t.ID, t.Name)
}

// listenTunnelPort 监听隧道远程端口。
func (s *Server) listenTunnelPort(t *Tunnel, proto string) {
	addr := fmt.Sprintf("%s:%d", s.cfg.Address, t.RemotePort)
	if proto == "udp" {
		s.serveUDPTunnel(t, addr)
		return
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Printf("隧道 %d 端口 %d 监听失败: %v", t.ID, t.RemotePort, err)
		// 端口冲突时上报面板标记
		return
	}
	s.logger.Printf("隧道 %d 监听 %s (%s)", t.ID, addr, proto)
	for {
		conn, err := ln.Accept()
		if err != nil {
			if t.stopping {
				return
			}
			s.logger.Printf("隧道 %d accept 失败: %v", t.ID, err)
			continue
		}
		go s.handlePublicConn(t, conn, "tcp")
	}
}

// handleControlConn 处理 frpc 控制连接。
func (s *Server) handleControlConn(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	reader := bufio.NewReader(conn)
	frame, err := readFrame(reader)
	if err != nil {
		return
	}

	switch frame.Type {
	case frameAuth:
		s.handleAuth(conn, reader, frame)
	case frameProxy:
		s.handleProxyJoin(conn, reader, frame)
	default:
		writeFrame(conn, Frame{Type: frameProxyErr, Error: "unknown frame"})
	}
}

// handleAuth 处理 frpc 鉴权。
func (s *Server) handleAuth(conn net.Conn, reader *bufio.Reader, frame Frame) {
	// 校验用户 Token：与面板下发的限速表匹配
	if _, ok := s.limiter.Get(frame.Token); !ok {
		writeFrame(conn, Frame{Type: frameAuthResult, OK: false, Error: "invalid token"})
		s.logger.Printf("鉴权失败: token=%s", frame.Token)
		return
	}
	s.logger.Printf("frpc 鉴权通过: token=%s", maskToken(frame.Token))
	writeFrame(conn, Frame{Type: frameAuthResult, OK: true})
}

// handleProxyJoin 处理 frpc 隧道代理连接加入连接池。
func (s *Server) handleProxyJoin(conn net.Conn, reader *bufio.Reader, frame Frame) {
	if frame.TunnelID == 0 || frame.Token == "" {
		writeFrame(conn, Frame{Type: frameProxyErr, Error: "bad params"})
		conn.Close()
		return
	}
	// 校验 token 与该隧道归属
	tc, ok := s.agent.GetTunnel(frame.TunnelID)
	if !ok || tc.UserToken != frame.Token {
		writeFrame(conn, Frame{Type: frameProxyErr, Error: "forbidden"})
		conn.Close()
		return
	}
	// 校验用户状态（banned 拒绝）
	limit, _ := s.limiter.Get(frame.Token)
	if limit == -1 { // -1 表示封禁
		writeFrame(conn, Frame{Type: frameProxyErr, Error: "banned"})
		conn.Close()
		return
	}

	s.mu.RLock()
	t := s.tunnels[frame.TunnelID]
	s.mu.RUnlock()
	if t == nil {
		writeFrame(conn, Frame{Type: frameProxyErr, Error: "tunnel not ready"})
		conn.Close()
		return
	}

	// 入池
	t.mu.Lock()
	t.idle = append(t.idle, conn)
	t.mu.Unlock()
	s.agent.AddConnection(t.ID, 1)
	writeFrame(conn, Frame{Type: frameProxyOK, TunnelID: t.ID})
	s.logger.Printf("隧道 %d 代理连接加入，当前池 %d", t.ID, len(t.idle))
}

// handlePublicConn 处理公网连接并转发到 frpc 代理连接。
func (s *Server) handlePublicConn(t *Tunnel, publicConn net.Conn, proto string) {
	defer publicConn.Close()
	if proto == "tcp" {
		s.agent.AddConnection(t.ID, 1)
		defer s.agent.AddConnection(t.ID, -1)
	}

	// 获取空闲代理连接
	proxyConn := s.acquireProxyConn(t)
	if proxyConn == nil {
		// 无可用代理连接，等待短暂后重试
		time.Sleep(200 * time.Millisecond)
		proxyConn = s.acquireProxyConn(t)
		if proxyConn == nil {
			writeFrame(publicConn, Frame{Type: frameProxyErr, Error: "no proxy available"})
			return
		}
	}

	// 转发（双向拷贝 + 限速）
	done := make(chan struct{}, 2)
	go func() {
		limitWriter := s.limiter.Writer(proxyConn, t.UserToken)
		n, _ := io.Copy(limitWriter, publicConn)
		s.agent.RecordTraffic(t.ID, int64(n), 0)
		done <- struct{}{}
	}()
	go func() {
		limitWriter := s.limiter.Writer(publicConn, t.UserToken)
		n, _ := io.Copy(limitWriter, proxyConn)
		s.agent.RecordTraffic(t.ID, 0, int64(n))
		done <- struct{}{}
	}()
	<-done
	proxyConn.Close()
}

// acquireProxyConn 从连接池获取空闲代理连接。
func (s *Server) acquireProxyConn(t *Tunnel) net.Conn {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.idle) == 0 {
		return nil
	}
	conn := t.idle[len(t.idle)-1]
	t.idle = t.idle[:len(t.idle)-1]
	return conn
}

// serveUDPTunnel UDP 隧道服务。
func (s *Server) serveUDPTunnel(t *Tunnel, addr string) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		s.logger.Printf("UDP 隧道 %d 解析失败: %v", t.ID, err)
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		s.logger.Printf("UDP 隧道 %d 监听失败: %v", t.ID, err)
		return
	}
	s.logger.Printf("UDP 隧道 %d 监听 %s", t.ID, addr)

	// UDP 简化实现：将数据包封装为 TCP 流发给 frpc 代理连接
	buf := make([]byte, 65535)
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		proxyConn := s.acquireProxyConn(t)
		if proxyConn == nil {
			continue
		}
		// 长度前缀帧
		frameData := append([]byte{'D'}, buf[:n]...)
		if _, err := proxyConn.Write(frameData); err != nil {
			proxyConn.Close()
			continue
		}
		s.agent.RecordTraffic(t.ID, int64(n), 0)
		// 读取回包
		proxyConn.SetReadDeadline(time.Now().Add(5 * time.Second))
		resp := make([]byte, 65535)
		rn, err := proxyConn.Read(resp)
		if err == nil && rn > 0 && resp[0] == 'D' {
			conn.WriteToUDP(resp[1:rn], remote)
			s.agent.RecordTraffic(t.ID, 0, int64(rn-1))
		}
		proxyConn.Close()
	}
}

// serveHTTPVhost HTTP/HTTPS 隧道按 Host 路由。
func (s *Server) serveHTTPVhost(addr string, proto string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Printf("vhost %s 端口监听失败: %v", proto, err)
		return
	}
	s.logger.Printf("vhost %s 监听 %s", proto, addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go s.handleHTTPConn(conn, proto)
	}
}

// handleHTTPConn 解析 HTTP 请求头获取 Host 并路由。
func (s *Server) handleHTTPConn(conn net.Conn, proto string) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))

	reader := bufio.NewReader(conn)
	// 读取请求行与 Host 头
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	host := ""
	// 最多读取 20 行头
	for i := 0; i < 20; i++ {
		hl, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		hl = strings.TrimSpace(hl)
		if hl == "" {
			break
		}
		lower := strings.ToLower(hl)
		if strings.HasPrefix(lower, "host:") {
			host = strings.TrimSpace(hl[5:])
			break
		}
	}
	if host == "" {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}
	host = strings.Split(host, ":")[0]

	tunnelID, ok := s.agent.GetDomainRoute(host)
	if !ok {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\nContent-Length: 9\r\n\r\nnot found"))
		return
	}
	s.mu.RLock()
	t := s.tunnels[tunnelID]
	s.mu.RUnlock()
	if t == nil {
		conn.Write([]byte("HTTP/1.1 503 Service Unavailable\r\n\r\n"))
		return
	}

	// 取出代理连接，把已读的字节连同请求一起转发
	proxyConn := s.acquireProxyConn(t)
	if proxyConn == nil {
		conn.Write([]byte("HTTP/1.1 503 Service Unavailable\r\n\r\n"))
		return
	}
	// 将请求行与已读头部回灌
	peeked := make([]byte, reader.Buffered())
	reader.Read(peeked)
	proxyConn.Write([]byte(line))
	proxyConn.Write(peeked)

	done := make(chan struct{}, 2)
	go func() {
		n, _ := io.Copy(proxyConn, conn)
		s.agent.RecordTraffic(t.ID, int64(n), 0)
		done <- struct{}{}
	}()
	go func() {
		n, _ := io.Copy(conn, proxyConn)
		s.agent.RecordTraffic(t.ID, 0, int64(n))
		done <- struct{}{}
	}()
	<-done
	proxyConn.Close()
}

// ---------- 帧编解码 ----------

func readFrame(reader *bufio.Reader) (Frame, error) {
	var f Frame
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return f, err
	}
	if err := json.Unmarshal(data, &f); err != nil {
		return f, err
	}
	return f, nil
}

func writeFrame(conn net.Conn, f Frame) error {
	data, err := json.Marshal(f)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = conn.Write(data)
	return err
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}
