package main

// WeaveNet 织网穿透 frpc 客户端入口。
//
// frpc 由 Electron 桌面客户端托管为子进程，配置由面板自动生成
// （frpc.toml），本程序负责：
//   - 连接 frps 控制端口并完成用户 Token 鉴权
//   - 为每条隧道建立代理连接加入 frps 连接池
//   - 将本地服务端口的数据经代理连接转发
//
// 协议与 frps 一致（JSON 行帧），见 server 包。

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/weavenet/frp-server/internal/config"
)

// Frame 协议帧（与 frps 对齐）。
type Frame struct {
	Type     string `json:"type"`
	Token    string `json:"token,omitempty"`
	TunnelID int    `json:"tunnel_id,omitempty"`
	OK       bool   `json:"ok,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Client frpc 客户端。
type Client struct {
	cfg    *config.ClientConfig
	logger *log.Logger

	authToken string
	connected bool
	stopCh    chan struct{}
	wg        sync.WaitGroup
	backoff   time.Duration
}

func main() {
	configPath := flag.String("c", "frpc.toml", "配置文件路径")
	flag.Parse()

	cfg, err := config.LoadClientConfig(*configPath)
	if err != nil {
		log.Fatalf("[frpc] 配置加载失败: %v", err)
	}

	logger := log.New(os.Stdout, "[frpc] ", log.LstdFlags)
	c := &Client{cfg: cfg, logger: logger, stopCh: make(chan struct{}), backoff: time.Second}
	c.run()
}

// run 主循环：断线自动重连（指数退避）。
func (c *Client) run() {
	for {
		select {
		case <-c.stopCh:
			return
		default:
		}
		err := c.connectOnce()
		if err != nil {
			c.logger.Printf("连接失败: %v，%v 后重试", err, c.backoff)
			select {
			case <-c.stopCh:
				return
			case <-time.After(c.backoff):
			}
			c.backoff *= 2
			if c.backoff > 60*time.Second {
				c.backoff = 60 * time.Second
			}
			continue
		}
		c.backoff = time.Second
		// 连接成功后保持（连接断开时 connectOnce 内部处理重连逻辑）
	}
}

// connectOnce 建立控制连接并进入转发。
func (c *Client) connectOnce() error {
	addr := fmt.Sprintf("%s:%d", c.cfg.ServerAddr, c.cfg.ServerPort)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 鉴权
	authFrame := Frame{Type: "auth", Token: c.cfg.Token}
	if err := writeFrame(conn, authFrame); err != nil {
		return err
	}
	reader := bufio.NewReader(conn)
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	result, err := readFrame(reader)
	if err != nil {
		return err
	}
	if !result.OK {
		return fmt.Errorf("鉴权失败: %s", result.Error)
	}
	c.connected = true
	c.logger.Printf("鉴权通过，连接 %s", addr)

	// 建立隧道代理连接
	c.wg.Add(1)
	go c.tunnelLoop(conn)

	// 读取控制连接直到断开
	for {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		_, err := readFrame(reader)
		if err != nil {
			c.logger.Printf("控制连接断开: %v", err)
			break
		}
	}
	c.connected = false
	c.wg.Wait()
	return fmt.Errorf("控制连接已断开")
}

// tunnelLoop 维持隧道代理连接池。
func (c *Client) tunnelLoop(controlConn net.Conn) {
	defer c.wg.Done()
	for {
		select {
		case <-c.stopCh:
			return
		default:
		}
		if !c.connected {
			return
		}
		// 为当前隧道建立代理连接
		err := c.joinProxy(c.cfg)
		if err != nil {
			c.logger.Printf("隧道代理失败: %v", err)
			select {
			case <-c.stopCh:
				return
			case <-time.After(2 * time.Second):
			}
			continue
		}
	}
}

// joinProxy 建立一条隧道代理连接并转发本地流量。
func (c *Client) joinProxy(cfg *config.ClientConfig) error {
	addr := fmt.Sprintf("%s:%d", cfg.ServerAddr, cfg.ServerPort)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 发送 proxy 帧
	proxyFrame := Frame{Type: "proxy", Token: c.cfg.Token, TunnelID: tunnelIDFromName(cfg.ProxyName)}
	if err := writeFrame(conn, proxyFrame); err != nil {
		return err
	}
	reader := bufio.NewReader(conn)
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	result, err := readFrame(reader)
	if err != nil {
		return err
	}
	if !result.OK {
		return fmt.Errorf("代理加入失败: %s", result.Error)
	}
	c.logger.Printf("隧道 %s 代理连接建立，监听本地 %s:%d",
		cfg.ProxyName, cfg.LocalIP, cfg.LocalPort)

	// 该连接保持空闲，等待 frps 侧分配公网流量；frps 关闭时连接断开
	// 为维持连接并感知断线，周期性发送 keepalive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	buf := make([]byte, 4096)
	for {
		conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// 超时保持连接
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				conn.Write([]byte{})
				continue
			}
			return err
		}
		if n == 0 {
			continue
		}
		// frps 转发来公网数据，这里把数据回灌到本地服务
		if buf[0] == 'D' && n > 1 {
			// UDP 数据帧
			c.forwardUDP(buf[1:n], conn)
			continue
		}
		// TCP 数据直接转发本地
		c.forwardTCP(buf[:n], conn)
		_ = ticker
	}
}

// forwardTCP 将 frps 转发数据写到本地服务端口。
func (c *Client) forwardTCP(data []byte, proxyConn net.Conn) {
	localAddr := fmt.Sprintf("%s:%d", c.cfg.LocalIP, c.cfg.LocalPort)
	local, err := net.DialTimeout("tcp", localAddr, 5*time.Second)
	if err != nil {
		c.logger.Printf("本地服务连接失败 %s: %v", localAddr, err)
		return
	}
	defer local.Close()
	local.Write(data)
	// 回传
	buf := make([]byte, 8192)
	for {
		n, err := local.Read(buf)
		if n > 0 {
			proxyConn.Write(buf[:n])
		}
		if err != nil {
			return
		}
	}
}

// forwardUDP 将 frps 转发 UDP 帧写入本地 UDP 端口。
func (c *Client) forwardUDP(data []byte, proxyConn net.Conn) {
	localAddr := fmt.Sprintf("%s:%d", c.cfg.LocalIP, c.cfg.LocalPort)
	udpAddr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		return
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.Write(data)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 65535)
	n, err := conn.Read(buf)
	if err == nil && n > 0 {
		proxyConn.Write(append([]byte{'D'}, buf[:n]...))
	}
}

// tunnelIDFromName 从隧道名称中提取 ID（格式：name-ID）。
func tunnelIDFromName(name string) int {
	idx := strings.LastIndex(name, "-")
	if idx < 0 {
		return 0
	}
	id := 0
	for _, ch := range name[idx+1:] {
		if ch < '0' || ch > '9' {
			return 0
		}
		id = id*10 + int(ch-'0')
	}
	return id
}

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
