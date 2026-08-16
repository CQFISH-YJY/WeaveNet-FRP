// WeaveNet 织网穿透 应急服务（逃生通道）。
//
// 独立于主面板的轻量应急进程，零 Web 框架依赖（仅标准库 + 纯 Go SQLite 驱动），
// 面板 / 数据库 / Redis 全部故障时仍可响应。
//
// 功能：
//
//	GET  /health           存活检查
//	GET  /status           系统状态（CPU/内存/磁盘/容器状态）
//	GET  /logs             拉取服务日志
//	GET  /data             直连 SQLite 拉取只读数据
//	POST /restart          重启服务
//	POST /stop             停止服务
//	POST /start            启动服务
//	POST /reboot           重启服务器
//	POST /exec             受限白名单命令执行
//
// 安全：
// 强随机预共享密钥（>=32 位）鉴权，连续失败 5 次锁定 10 分钟，
// 危险操作（reboot/exec）需二次确认参数。
//
// 运行：emergency --port 9001
// 文档：CQFISH&喵酱出品
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const (
	maxFails    = 5
	lockSeconds = 600
	maxExecOut  = 500
	maxExecErr  = 200
)

// ---------- 配置 ----------

var (
	baseDir, _ = filepath.Abs(".")
	configFile = filepath.Join(baseDir, "emergency.conf")
	logFile    = filepath.Join(baseDir, "emergency.log")
)

type config struct {
	Port           string              `json:"port"`
	SecretKey      string              `json:"secret_key"`
	IPWhitelist    string              `json:"ip_whitelist"`
	DBPath         string              `json:"db_path"`
	ServiceCommand map[string][]string `json:"service_command"`
}

func defaultConfig() *config {
	return &config{Port: "9001"}
}

func loadConfig() *config {
	cfg := defaultConfig()
	if data, err := os.ReadFile(configFile); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			idx := strings.Index(line, "=")
			if idx < 0 {
				continue
			}
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			switch key {
			case "port":
				cfg.Port = val
			case "secret_key":
				cfg.SecretKey = val
			case "ip_whitelist":
				cfg.IPWhitelist = val
			case "db_path":
				cfg.DBPath = val
			case "service_command":
				_ = json.Unmarshal([]byte(val), &cfg.ServiceCommand)
			}
		}
	}
	if len(cfg.SecretKey) < 32 {
		cfg.SecretKey = genSecretKey()
		saveConfig(cfg)
	}
	return cfg
}

func genSecretKey() string {
	buf := make([]byte, 48)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("emergency_%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(buf)[:48]
}

func saveConfig(cfg *config) {
	scm, _ := json.Marshal(cfg.ServiceCommand)
	lines := []string{
		"# WeaveNet 应急服务配置",
		"port = " + cfg.Port,
		"secret_key = " + cfg.SecretKey,
		"ip_whitelist = " + cfg.IPWhitelist,
		"db_path = " + cfg.DBPath,
		"service_command = " + string(scm),
		"",
	}
	_ = os.WriteFile(configFile, []byte(strings.Join(lines, "\n")), 0o600)
}

// ---------- 审计日志 ----------

func audit(action, detail string) {
	ts := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("%s [%s] %s\n", ts, action, detail)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line)
}

// ---------- 鉴权 ----------

type authGuard struct {
	mu          sync.Mutex
	secret      []byte
	fails       int
	lockedUntil time.Time
}

func (g *authGuard) verify(requestSecret string) bool {
	now := time.Now()
	g.mu.Lock()
	if now.Before(g.lockedUntil) {
		g.mu.Unlock()
		return false
	}
	if g.fails >= maxFails {
		g.lockedUntil = now.Add(lockSeconds * time.Second)
		g.fails = 0
		g.mu.Unlock()
		audit("lock", fmt.Sprintf("连续失败 %d 次，锁定 %d 秒", maxFails, lockSeconds))
		return false
	}
	g.mu.Unlock()

	ok := subtle.ConstantTimeCompare([]byte(requestSecret), g.secret) == 1
	g.mu.Lock()
	if ok {
		g.fails = 0
	} else {
		g.fails++
	}
	g.mu.Unlock()
	return ok
}

func (g *authGuard) locked() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return time.Now().Before(g.lockedUntil)
}

// ---------- 系统状态 ----------

type systemInfo struct {
	Time       string         `json:"time"`
	Platform   string         `json:"platform"`
	GoVersion  string         `json:"go_version"`
	CPUCount   int            `json:"cpu_count"`
	Loadavg    []float64      `json:"loadavg"`
	Memory     map[string]any `json:"memory"`
	Disk       map[string]any `json:"disk"`
	Containers map[string]any `json:"containers"`
}

func systemStatus() *systemInfo {
	st := &systemInfo{
		Time:       time.Now().Format("2006-01-02 15:04:05"),
		Platform:   runtime.GOOS + "/" + runtime.GOARCH,
		GoVersion:  runtime.Version(),
		CPUCount:   runtime.NumCPU(),
		Memory:     map[string]any{},
		Disk:       map[string]any{},
		Containers: map[string]any{},
	}
	st.Loadavg = readLoadavg()
	st.Memory = readMemory()
	st.Disk = readDisk()
	st.Containers = readContainers()
	return st
}

// ---------- HTTP 服务 ----------

type server struct {
	cfg   *config
	guard *authGuard
}

func (s *server) send(w http.ResponseWriter, code int, payload map[string]any) {
	body, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(code)
	_, _ = w.Write(body)
}

func (s *server) ok(w http.ResponseWriter, data any, message string) {
	s.send(w, http.StatusOK, map[string]any{"code": 0, "message": message, "data": data})
}

func (s *server) fail(w http.ResponseWriter, code int, message string) {
	s.send(w, code, map[string]any{"code": code, "message": message, "data": nil})
}

func (s *server) auth(w http.ResponseWriter, r *http.Request) bool {
	// IP 白名单
	if wl := strings.TrimSpace(s.cfg.IPWhitelist); wl != "" {
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			clientIP = r.RemoteAddr
		}
		allowed := false
		for _, ip := range strings.Split(wl, ",") {
			if strings.TrimSpace(ip) == clientIP {
				allowed = true
				break
			}
		}
		if !allowed {
			audit("deny_ip", clientIP+" 不在白名单")
			s.fail(w, http.StatusForbidden, "IP 不在白名单内")
			return false
		}
	}
	secret := r.Header.Get("X-Emergency-Key")
	if !s.guard.verify(secret) {
		if s.guard.locked() {
			s.fail(w, 423, "已锁定，请稍后再试")
		} else {
			s.fail(w, http.StatusUnauthorized, "密钥错误")
		}
		return false
	}
	return true
}

func queryValue(r *http.Request, name, def string) string {
	vals, ok := r.URL.Query()[name]
	if !ok || len(vals) == 0 {
		return def
	}
	return vals[0]
}

func confirmToken(action string) string {
	sum := sha256.Sum256([]byte("weavenet:" + action))
	return hex.EncodeToString(sum[:])[:16]
}

func (s *server) confirm(w http.ResponseWriter, r *http.Request, action string) bool {
	expected := confirmToken(action)
	if queryValue(r, "confirm", "") != expected {
		s.fail(w, http.StatusBadRequest, "危险操作需二次确认，请携带 confirm="+expected)
		return false
	}
	return true
}

func (s *server) handle(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			audit("error", fmt.Sprintf("%s %s: %v", r.Method, r.URL.Path, rec))
			s.fail(w, http.StatusInternalServerError, "内部错误")
		}
	}()

	path := r.URL.Path

	// GET /health 免鉴权
	if r.Method == http.MethodGet && path == "/health" {
		s.ok(w, map[string]any{"status": "alive", "service": "emergency"}, "存活")
		return
	}
	if !s.auth(w, r) {
		return
	}

	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	switch r.Method {
	case http.MethodGet:
		switch path {
		case "/status":
			audit("status", "来自 "+clientIP)
			s.ok(w, systemStatus(), "系统状态")
		case "/logs":
			service := queryValue(r, "service", "panel")
			lines := 200
			if v := queryValue(r, "lines", ""); v != "" {
				if n, err := strconv.Atoi(v); err == nil {
					lines = n
				}
			}
			s.ok(w, map[string]any{"service": service, "log": tailLog(service, lines)}, "日志")
		case "/data":
			s.readData(w, r)
		default:
			s.fail(w, http.StatusNotFound, "接口不存在")
		}
	case http.MethodPost:
		switch path {
		case "/restart":
			if !s.confirm(w, r, "restart") {
				return
			}
			service := queryValue(r, "service", "panel")
			audit("restart", service+" 重启（二次确认通过）")
			s.ok(w, map[string]any{"service": service}, "重启指令已执行")
			go s.restartService(service)
		case "/stop":
			service := queryValue(r, "service", "all")
			audit("stop", service+" 停止")
			s.ok(w, map[string]any{"service": service}, "停止指令已执行")
			go s.stopService(service)
		case "/start":
			service := queryValue(r, "service", "panel")
			audit("start", service+" 启动")
			s.ok(w, map[string]any{"service": service}, "启动指令已执行")
			go s.startService(service)
		case "/reboot":
			if !s.confirm(w, r, "reboot") {
				return
			}
			audit("reboot", "服务器重启（二次确认通过）")
			s.ok(w, map[string]any{}, "服务器重启指令已执行")
			go reboot()
		case "/exec":
			if !s.confirm(w, r, "exec") {
				return
			}
			cmd := queryValue(r, "cmd", "")
			if !allowedCommand(cmd) {
				s.fail(w, http.StatusBadRequest, "命令不在白名单内")
				return
			}
			audit("exec", "执行: "+cmd)
			go execCmd(cmd)
			s.ok(w, map[string]any{"cmd": cmd}, "命令已执行")
		default:
			s.fail(w, http.StatusNotFound, "接口不存在")
		}
	default:
		s.fail(w, http.StatusMethodNotAllowed, "仅支持 GET/POST")
	}
}

func tailLog(service string, lines int) string {
	path := filepath.Join(baseDir, service+".log")
	data, err := os.ReadFile(path)
	if err != nil {
		return "(日志文件不存在)"
	}
	all := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if lines < 1 {
		lines = 1
	}
	if lines > 2000 {
		lines = 2000
	}
	if len(all) > lines {
		all = all[len(all)-lines:]
	}
	return strings.Join(all, "\n")
}

var allowedPrefixes = []string{
	"systemctl status", "docker ps", "docker logs", "free -m",
	"df -h", "uptime", "ps aux", "netstat -tlnp", "cat /etc/os-release",
}

func allowedCommand(cmd string) bool {
	for _, p := range allowedPrefixes {
		if strings.HasPrefix(cmd, p) {
			return true
		}
	}
	return false
}

func (s *server) readData(w http.ResponseWriter, r *http.Request) {
	dataType := queryValue(r, "type", "users")
	if s.cfg.DBPath == "" {
		s.fail(w, http.StatusBadRequest, "未配置数据库路径")
		return
	}
	if _, err := os.Stat(s.cfg.DBPath); err != nil {
		s.fail(w, http.StatusBadRequest, "未配置数据库路径")
		return
	}
	if dataType != "users" && dataType != "tunnels" && dataType != "nodes" {
		s.fail(w, http.StatusBadRequest, "type 仅支持 users|tunnels|nodes")
		return
	}
	limit := 100
	if v := queryValue(r, "limit", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > 1000 {
		limit = 1000
	}

	dsn := "file:" + filepath.ToSlash(s.cfg.DBPath) + "?mode=ro"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "数据库打开失败: "+err.Error())
		return
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s LIMIT ?", dataType), limit)
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "数据库读取失败: "+err.Error())
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		s.fail(w, http.StatusInternalServerError, "数据库读取失败: "+err.Error())
		return
	}
	var out []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			break
		}
		row := map[string]any{}
		for i, c := range cols {
			switch v := vals[i].(type) {
			case []byte:
				row[c] = string(v)
			default:
				row[c] = v
			}
		}
		out = append(out, row)
	}
	audit("data", fmt.Sprintf("读取 %s %d 行", dataType, len(out)))
	s.ok(w, map[string]any{"type": dataType, "count": len(out), "rows": out}, "数据")
}

// ---------- 服务操作 ----------

func (s *server) serviceCommands(service string) []string {
	cmds := s.cfg.ServiceCommand
	if cmds == nil {
		return nil
	}
	return cmds[service]
}

func runCommands(cmds []string) {
	for _, cmd := range cmds {
		if cmd == "" {
			continue
		}
		shell := "sh"
		if runtime.GOOS == "windows" {
			shell = "cmd"
		}
		c := exec.Command(shell, "-c", cmd)
		if err := c.Start(); err != nil {
			audit("cmd_error", cmd+": "+err.Error())
		}
	}
}

func (s *server) restartService(service string) {
	cmds := s.serviceCommands(service)
	if len(cmds) == 0 {
		audit("restart_skip", service+" 未配置命令，跳过")
		return
	}
	runCommands(cmds)
}

func (s *server) stopService(service string) {
	cmds := s.serviceCommands(service)
	if len(cmds) == 0 {
		audit("stop_skip", service+" 未配置命令，跳过")
		return
	}
	runCommands(cmds)
}

func (s *server) startService(service string) {
	cmds := s.serviceCommands(service)
	if len(cmds) == 0 {
		audit("start_skip", service+" 未配置命令，跳过")
		return
	}
	runCommands(cmds)
}

func execCmd(cmd string) {
	shell := "sh"
	if runtime.GOOS == "windows" {
		shell = "cmd"
	}
	c := exec.Command(shell, "-c", cmd)
	out, err := c.CombinedOutput()
	if err != nil {
		audit("exec_error", cmd+": "+err.Error())
		return
	}
	audit("exec_result", "exit=0 out="+truncate(string(out), maxExecOut))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func reboot() {
	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command("shutdown", "-r", "now")
	} else {
		cmd = exec.Command("shutdown", "/r", "/t", "5")
	}
	if err := cmd.Start(); err != nil {
		audit("reboot_error", err.Error())
	}
}

// ---------- 入口 ----------

func main() {
	cfg := loadConfig()
	guard := &authGuard{secret: []byte(cfg.SecretKey)}
	srv := &server{cfg: cfg, guard: guard}

	port := cfg.Port
	if len(os.Args) > 2 && os.Args[1] == "--port" {
		port = os.Args[2]
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handle)

	addr := "0.0.0.0:" + port
	audit("start", "应急服务启动，端口 "+port)
	fmt.Printf("WeaveNet 应急服务运行于 http://%s\n", addr)
	fmt.Printf("密钥: %s\n", cfg.SecretKey)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
