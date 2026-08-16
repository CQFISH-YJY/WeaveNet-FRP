package api

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"weavenet/panel/cache"
	"weavenet/panel/database"
	"weavenet/panel/middleware"
	"weavenet/panel/models"
	"weavenet/panel/services"
	"weavenet/panel/utils"
)

type tunnelPayload struct {
	Name          string         `json:"name"`
	NodeID        uint           `json:"node_id"`
	Type          string         `json:"type"`
	LocalIP       string         `json:"local_ip"`
	LocalPort     int            `json:"local_port"`
	RemotePort    *int           `json:"remote_port"`
	Subdomain     *string        `json:"subdomain"`
	CustomDomain  *string        `json:"custom_domain"`
	KCP           bool           `json:"kcp"`
	Encryption    bool           `json:"encryption"`
	Compression   bool           `json:"compression"`
	SecretKey     *string        `json:"secret_key"`
	LoadBalancers []map[string]any `json:"load_balancers"`
}

// serializeTunnel 序列化隧道（含实时状态与公网地址）。
func serializeTunnel(t *models.Tunnel, node *models.Node) gin.H {
	var nodeInfo gin.H
	if node != nil && node.ID > 0 {
		nodeInfo = gin.H{
			"id": node.ID, "name": node.Name, "address": node.Address, "port": node.Port,
			"status": node.Status, "speed_limit_mbps": node.SpeedLimitMbps,
			"last_heartbeat_at": utils.TimeFmt(node.LastHeartbeatAt),
		}
	} else {
		nodeInfo = nil
	}
	runtime, _ := cache.C.Get(cache.RuntimeKey(t.ID))
	online := false
	connections := 0
	if rt, ok := runtime.(cache.TunnelRuntime); ok {
		online = rt.Online
		connections = rt.Connections
	}
	today := utils.Today()
	tr, _ := cache.GetTraffic(today, t.ID)
	addr := ""
	if node != nil {
		addr = services.BuildPublicAddress(t, node)
	}
	return gin.H{
		"id": t.ID, "user_id": t.UserID, "node_id": t.NodeID, "name": t.Name, "type": t.Type,
		"local_ip": t.LocalIP, "local_port": t.LocalPort, "remote_port": t.RemotePort,
		"subdomain": t.Subdomain, "custom_domain": t.CustomDomain, "kcp": t.KCP,
		"encryption": t.Encryption, "compression": t.Compression, "secret_key": t.SecretKey,
		"load_balancers": services.ParseLB(t.LoadBalancers), "status": t.Status,
		"status_detail": t.StatusDetail, "created_at": utils.TimeFmtV(t.CreatedAt),
		"node": nodeInfo, "online": online, "connections": connections,
		"today_in": tr.In, "today_out": tr.Out, "public_address": addr,
	}
}

// ListTunnels 隧道列表。
func ListTunnels(c *gin.Context) {
	user := middleware.GetUser(c)
	db := database.DB
	q := db.Where("user_id = ?", user.ID)
	if t := c.Query("type"); t != "" {
		q = q.Where("type = ?", t)
	}
	var tunnels []models.Tunnel
	q.Order("id desc").Find(&tunnels)
	items := make([]gin.H, 0, len(tunnels))
	for _, t := range tunnels {
		var node models.Node
		db.First(&node, t.NodeID)
		items = append(items, serializeTunnel(&t, &node))
	}
	utils.OK(c, items, "ok")
}

// getOwnedTunnel 获取用户所属隧道。
func getOwnedTunnel(c *gin.Context) (*models.Tunnel, *models.Node, *utils.BizError) {
	user := middleware.GetUser(c)
	id, err := strconv.Atoi(c.Param("tunnel_id"))
	if err != nil {
		return nil, nil, utils.NotFound("隧道不存在")
	}
	db := database.DB
	var t models.Tunnel
	if err := db.Where("id = ? AND user_id = ?", id, user.ID).First(&t).Error; err != nil {
		return nil, nil, utils.NotFound("隧道不存在")
	}
	var node models.Node
	db.First(&node, t.NodeID)
	return &t, &node, nil
}

// GetTunnel 隧道详情。
func GetTunnel(c *gin.Context) {
	t, node, be := getOwnedTunnel(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	utils.OK(c, serializeTunnel(t, node), "ok")
}

// CreateTunnel 创建隧道。
func CreateTunnel(c *gin.Context) {
	var p tunnelPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.ErrBiz(422, 0, "参数错误"))
		return
	}
	db := database.DB
	user := middleware.GetUser(c)
	if p.Name == "" || p.NodeID == 0 || p.LocalPort < 1 {
		utils.Fail(c, utils.ErrBiz(400, 0, "请填写名称、节点与本地端口"))
		return
	}
	var node models.Node
	if err := db.First(&node, p.NodeID).Error; err != nil {
		utils.Fail(c, utils.NotFound("节点不存在"))
		return
	}
	if node.Status != "online" {
		utils.Fail(c, utils.ErrBiz(400, 3004, "节点离线或维护中，无法创建隧道"))
		return
	}
	if err := services.CheckTunnelLimit(db, user, &user.Plan); err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	remotePort := p.RemotePort
	if p.Type == "tcp" || p.Type == "udp" || p.Type == "kcp" || p.Type == "loadbalance" {
		if remotePort == nil {
			port, err := services.AllocateRemotePort(db, p.NodeID)
			if err != nil {
				utils.Fail(c, utils.AsBiz(err))
				return
			}
			remotePort = &port
		} else if !services.CheckRemotePortAvailable(db, p.NodeID, *remotePort, 0) {
			utils.Fail(c, utils.ErrBiz(409, 3001, "远程端口已被占用"))
			return
		}
	}
	if p.Subdomain != nil && *p.Subdomain != "" {
		if !services.CheckSubdomainAvailable(db, *p.Subdomain, 0) {
			utils.Fail(c, utils.ErrBiz(409, 3002, "子域名已被占用"))
			return
		}
	}
	lbStr := ""
	if len(p.LoadBalancers) > 0 {
		b, _ := json.Marshal(p.LoadBalancers)
		lbStr = string(b)
	}
	t := models.Tunnel{
		UserID: user.ID, NodeID: p.NodeID, Name: p.Name, Type: p.Type,
		LocalIP: orDefault(p.LocalIP, "127.0.0.1"), LocalPort: p.LocalPort,
		RemotePort: remotePort, Subdomain: p.Subdomain, CustomDomain: p.CustomDomain,
		KCP: p.KCP, Encryption: p.Encryption, Compression: p.Compression,
		SecretKey: p.SecretKey, LoadBalancers: lbStr, Status: "stopped",
	}
	if t.Type == "stcp" || t.Type == "xtcp" {
		sk := utils.GenSecretKey()
		t.SecretKey = &sk
	}
	if err := db.Create(&t).Error; err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	utils.OK(c, serializeTunnel(&t, &node), "隧道创建成功")
}

// UpdateTunnel 更新隧道。
func UpdateTunnel(c *gin.Context) {
	t, _, be := getOwnedTunnel(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	var p struct {
		Name         string         `json:"name"`
		LocalIP      string         `json:"local_ip"`
		LocalPort    int            `json:"local_port"`
		RemotePort   *int           `json:"remote_port"`
		Subdomain    *string        `json:"subdomain"`
		CustomDomain *string        `json:"custom_domain"`
		KCP          *bool          `json:"kcp"`
		Encryption   *bool          `json:"encryption"`
		Compression  *bool          `json:"compression"`
		SecretKey    *string        `json:"secret_key"`
		LoadBalancers []map[string]any `json:"load_balancers"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.ErrBiz(422, 0, "参数错误"))
		return
	}
	db := database.DB
	if p.Name != "" {
		t.Name = p.Name
	}
	if p.LocalIP != "" {
		t.LocalIP = p.LocalIP
	}
	if p.LocalPort > 0 {
		t.LocalPort = p.LocalPort
	}
	if p.RemotePort != nil {
		if !services.CheckRemotePortAvailable(db, t.NodeID, *p.RemotePort, t.ID) {
			utils.Fail(c, utils.ErrBiz(409, 3001, "远程端口已被占用"))
			return
		}
		t.RemotePort = p.RemotePort
	}
	if p.Subdomain != nil && *p.Subdomain != "" {
		if !services.CheckSubdomainAvailable(db, *p.Subdomain, 0) {
			utils.Fail(c, utils.ErrBiz(409, 3002, "子域名已被占用"))
			return
		}
		t.Subdomain = p.Subdomain
	}
	if p.CustomDomain != nil {
		t.CustomDomain = p.CustomDomain
	}
	if p.KCP != nil {
		t.KCP = *p.KCP
	}
	if p.Encryption != nil {
		t.Encryption = *p.Encryption
	}
	if p.Compression != nil {
		t.Compression = *p.Compression
	}
	if p.SecretKey != nil {
		t.SecretKey = p.SecretKey
	}
	if len(p.LoadBalancers) > 0 {
		b, _ := json.Marshal(p.LoadBalancers)
		t.LoadBalancers = string(b)
	}
	db.Save(t)
	utils.OK(c, serializeTunnel(t, nil), "隧道更新成功")
}

// DeleteTunnel 删除隧道。
func DeleteTunnel(c *gin.Context) {
	t, _, be := getOwnedTunnel(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	db := database.DB
	db.Model(&models.Domain{}).Where("tunnel_id = ?", t.ID).Update("tunnel_id", nil)
	db.Delete(t)
	cache.C.Del(cache.RuntimeKey(t.ID))
	cache.C.Del(cache.WantKey(t.ID))
	utils.NoContent(c)
}

// StartTunnel 启动隧道。
func StartTunnel(c *gin.Context) {
	t, _, be := getOwnedTunnel(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	var node models.Node
	database.DB.First(&node, t.NodeID)
	if node.Status != "online" {
		utils.Fail(c, utils.ErrBiz(400, 3004, "节点离线或维护中"))
		return
	}
	t.Status = "running"
	database.DB.Save(t)
	// 写启停指令（frps 轮询感知）
	cmd, _ := json.Marshal(gin.H{"action": "start", "force": false})
	cache.C.Set(cache.WantKey(t.ID), cmd, time.Hour)
	utils.OKMsg(c, "隧道启动指令已下发")
}

// StopTunnel 停止隧道。
func StopTunnel(c *gin.Context) {
	t, _, be := getOwnedTunnel(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	t.Status = "stopped"
	database.DB.Save(t)
	cmd, _ := json.Marshal(gin.H{"action": "stop", "force": false})
	cache.C.Set(cache.WantKey(t.ID), cmd, time.Hour)
	utils.OKMsg(c, "隧道停止指令已下发")
}

// GetTunnelConfig 生成 frpc.toml 配置。
func GetTunnelConfig(c *gin.Context) {
	t, node, be := getOwnedTunnel(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	user := middleware.GetUser(c)
	cfg := services.GenerateFrpcConfig(database.DB, t, user, node)
	utils.OK(c, gin.H{"config": cfg}, "ok")
}

// GetTunnelStatus 隧道实时状态。
func GetTunnelStatus(c *gin.Context) {
	t, _, be := getOwnedTunnel(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	rt, _ := cache.C.Get(cache.RuntimeKey(t.ID))
	online, connections := false, 0
	var in, out int64
	ts := ""
	if r, ok := rt.(cache.TunnelRuntime); ok {
		online, connections = r.Online, r.Connections
		in, out, ts = r.In, r.Out, r.Ts
	}
	tr, _ := cache.GetTraffic(utils.Today(), t.ID)
	utils.OK(c, gin.H{
		"status": t.Status, "online": online, "connections": connections,
		"in_bytes": in, "out_bytes": out, "today_in": tr.In, "today_out": tr.Out, "ts": ts,
	}, "ok")
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
