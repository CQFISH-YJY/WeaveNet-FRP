package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"weavenet/panel/cache"
	"weavenet/panel/config"
	"weavenet/panel/database"
	"weavenet/panel/middleware"
	"weavenet/panel/models"
	"weavenet/panel/utils"
)

// AgentRegister 节点注册：frps 启动时携带 Agent Token 调用，不存在则自动注册。
func AgentRegister(c *gin.Context) {
	var p struct {
		AgentToken string `json:"agent_token"`
		Name       string `json:"name"`
		Address    string `json:"address"`
		Port       int    `json:"port"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.Biz(400, 0, "参数错误"))
		return
	}
	if p.AgentToken == "" {
		utils.Fail(c, utils.NotFound("缺少 agent_token"))
		return
	}
	db := database.DB
	var node models.Node
	if err := db.Where("agent_token = ?", p.AgentToken).First(&node).Error; err != nil {
		// 允许按名称注册：name + token 组合
		if p.Name != "" {
			if err := db.Where("name = ?", p.Name).First(&node).Error; err == nil {
				node.AgentToken = p.AgentToken
			}
		}
		if node.ID == 0 {
			utils.Fail(c, utils.NotFound("节点未授权，请在管理后台创建节点后配置 Agent Token"))
			return
		}
	}
	if p.Address != "" {
		node.Address = p.Address
	}
	if p.Port > 0 {
		node.Port = p.Port
	}
	now := time.Now()
	node.LastHeartbeatAt = &now
	node.Status = "online"
	cache.C.Set(cache.NodeKey(node.ID), true, time.Duration(config.C.NodeHeartbeatTO)*time.Second*3)
	if err := db.Save(&node).Error; err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	utils.OK(c, gin.H{"node_id": node.ID, "status": "registered"}, "ok")
}

// AgentHeartbeat 心跳上报：隧道状态、连接数、流量增量。
func AgentHeartbeat(c *gin.Context) {
	node := middleware.GetAgentNode(c)
	db := database.DB
	var p struct {
		Tunnels []struct {
			TunnelID    uint  `json:"tunnel_id"`
			Online      bool  `json:"online"`
			Connections int   `json:"connections"`
			InDelta     int64 `json:"in_delta"`
			OutDelta    int64 `json:"out_delta"`
		} `json:"tunnels"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.Biz(400, 0, "参数错误"))
		return
	}
	now := time.Now()
	node.LastHeartbeatAt = &now
	node.Status = "online"
	cache.C.Set(cache.NodeKey(node.ID), true, time.Duration(config.C.NodeHeartbeatTO)*time.Second*3)

	for _, item := range p.Tunnels {
		if item.TunnelID == 0 {
			continue
		}
		// 实时状态入内存缓存
		rt := cache.TunnelRuntime{
			Online:      item.Online,
			Connections: item.Connections,
			In:          item.InDelta,
			Out:         item.OutDelta,
			Ts:          now.Format("2006-01-02T15:04:05"),
		}
		if v, ok := cache.C.Get(cache.RuntimeKey(item.TunnelID)); ok {
			if old, ok2 := v.(cache.TunnelRuntime); ok2 {
				rt.In += old.In
				rt.Out += old.Out
			}
		}
		cache.C.Set(cache.RuntimeKey(item.TunnelID), rt, 24*time.Hour)
		// 流量增量累加当日总量
		if item.InDelta != 0 || item.OutDelta != 0 {
			cache.AddTraffic(utils.Today(), item.TunnelID, item.InDelta, item.OutDelta)
		}
		// 同步持久化状态
		var t models.Tunnel
		if err := db.First(&t, item.TunnelID).Error; err == nil {
			if item.Online {
				t.Status = "running"
				t.StatusDetail = "在线"
			}
			db.Save(&t)
		}
	}
	if err := db.Save(node).Error; err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	utils.OKMsg(c, "ok")
}

// AgentTunnels 拉取本节点隧道与用户限速配置（frps 每 10s 轮询）。
func AgentTunnels(c *gin.Context) {
	node := middleware.GetAgentNode(c)
	db := database.DB
	var tunnels []models.Tunnel
	db.Preload("Node").Where("node_id = ? AND status = ?", node.ID, "running").Find(&tunnels)
	result := make([]gin.H, 0, len(tunnels))
	userIDs := make(map[uint]bool)
	for i := range tunnels {
		t := &tunnels[i]
		var user models.User
		if err := db.Preload("Plan").First(&user, t.UserID).Error; err != nil {
			continue
		}
		bandwidth := 8000
		if user.Plan.ID > 0 {
			bandwidth = user.Plan.SpeedLimitMbps * 1000
		}
		result = append(result, gin.H{
			"tunnel_id":            t.ID,
			"user_token":           utils.UserTokenForFrpc(user.ID, user.PasswordHash),
			"username":             user.Username,
			"type":                 t.Type,
			"name":                 t.Name,
			"local_ip":             t.LocalIP,
			"local_port":           t.LocalPort,
			"remote_port":          t.RemotePort,
			"subdomain":            t.Subdomain,
			"custom_domain":        t.CustomDomain,
			"kcp":                  t.KCP,
			"encryption":           t.Encryption,
			"compression":          t.Compression,
			"secret_key":           t.SecretKey,
			"bandwidth_limit_kbps": bandwidth,
			"load_balancers":       t.LoadBalancers,
		})
		userIDs[user.ID] = true
	}
	// 同时下发所有在线用户的限速配置
	speedLimits := make([]gin.H, 0, len(userIDs))
	for uid := range userIDs {
		var u models.User
		if err := db.Preload("Plan").First(&u, uid).Error; err != nil {
			continue
		}
		bandwidth := 8000
		if u.Plan.ID > 0 {
			bandwidth = u.Plan.SpeedLimitMbps * 1000
		}
		speedLimits = append(speedLimits, gin.H{
			"user_token":           utils.UserTokenForFrpc(u.ID, u.PasswordHash),
			"username":             u.Username,
			"bandwidth_limit_kbps": bandwidth,
			"status":               u.Status,
		})
	}
	// 域名路由表
	var domains []models.Domain
	db.Where("status = ?", "active").Find(&domains)
	domainItems := make([]gin.H, 0, len(domains))
	for _, d := range domains {
		domainItems = append(domainItems, gin.H{
			"full_domain": d.FullDomain,
			"subdomain":   d.Subdomain,
			"tunnel_id":   d.TunnelID,
		})
	}
	utils.OK(c, gin.H{"tunnels": result, "speed_limits": speedLimits, "domains": domainItems}, "ok")
}
