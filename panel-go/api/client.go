package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"weavenet/panel/config"
	"weavenet/panel/database"
	"weavenet/panel/middleware"
	"weavenet/panel/models"
	"weavenet/panel/services"
	"weavenet/panel/utils"
)

// ClientLogin 客户端登录：返回专用 Token 与用户基本信息。
func ClientLogin(c *gin.Context) {
	var p struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.Biz(400, 0, "参数错误"))
		return
	}
	db := database.DB
	var user models.User
	if err := db.Preload("Plan").Where("username = ? OR email = ?", p.Username, p.Username).First(&user).Error; err != nil {
		utils.Fail(c, utils.Biz(401, 0, "用户名或密码错误"))
		return
	}
	if !utils.VerifyPassword(user.PasswordHash, p.Password) {
		utils.Fail(c, utils.Biz(401, 0, "用户名或密码错误"))
		return
	}
	if !user.EmailVerified {
		utils.Fail(c, utils.Biz(403, 1005, "邮箱未验证，请先在网页端激活账号"))
		return
	}
	if user.Status == "banned" {
		utils.Fail(c, utils.Biz(403, 1006, "账号已被封禁，请联系管理员"))
		return
	}
	token := utils.GenToken()
	expiresAt := time.Now().Add(time.Duration(config.C.SessionDays) * 24 * time.Hour)
	sess := models.Session{Token: token, UserID: user.ID, ExpiresAt: expiresAt}
	if err := db.Create(&sess).Error; err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	planName := ""
	speed := 8
	if user.Plan.ID > 0 {
		planName = user.Plan.Name
		speed = user.Plan.SpeedLimitMbps
	}
	utils.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id": user.ID, "username": user.Username, "email": user.Email,
			"plan_name": planName,
			"plan_expires_at": utils.TimeFmt(user.PlanExpiresAt),
			"speed_limit_mbps": speed,
		},
	}, "登录成功")
}

// ClientTunnels 拉取用户隧道 + 节点信息。
func ClientTunnels(c *gin.Context) {
	user := middleware.GetUser(c)
	db := database.DB
	var tunnels []models.Tunnel
	db.Preload("Node").Where("user_id = ?", user.ID).Order("id desc").Find(&tunnels)
	speed := 8
	if user.Plan.ID > 0 {
		speed = user.Plan.SpeedLimitMbps
	}
	items := make([]gin.H, 0, len(tunnels))
	for _, t := range tunnels {
		items = append(items, gin.H{
			"id": t.ID, "name": t.Name, "type": t.Type, "node_id": t.NodeID,
			"node_name": t.Node.Name, "node_address": t.Node.Address,
			"node_port": t.Node.Port,
			"local_ip": t.LocalIP, "local_port": t.LocalPort,
			"remote_port": t.RemotePort, "subdomain": t.Subdomain,
			"custom_domain": t.CustomDomain, "kcp": t.KCP,
			"encryption": t.Encryption, "compression": t.Compression,
			"secret_key": t.SecretKey, "load_balancers": services.ParseLB(t.LoadBalancers),
			"status": t.Status,
			"bandwidth_limit_kbps": speed * 1000,
			"public_address":       clientPublicAddress(&t),
		})
	}
	utils.OK(c, gin.H{
		"user": gin.H{
			"id": user.ID, "username": user.Username,
			"plan_name": user.Plan.Name, "tunnel_limit": user.Plan.TunnelLimit,
			"domain_limit": user.Plan.DomainLimit, "speed_limit_mbps": speed,
		},
		"tunnels": items,
	}, "ok")
}

// ClientConfig 生成 frpc.toml 配置。
func ClientConfig(c *gin.Context) {
	var p struct {
		TunnelID uint `json:"tunnel_id"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.Biz(400, 0, "参数错误"))
		return
	}
	user := middleware.GetUser(c)
	db := database.DB
	var t models.Tunnel
	if err := db.Where("id = ? AND user_id = ?", p.TunnelID, user.ID).First(&t).Error; err != nil {
		utils.Fail(c, utils.NotFound("隧道不存在"))
		return
	}
	var node models.Node
	if err := db.First(&node, t.NodeID).Error; err != nil {
		utils.Fail(c, utils.NotFound("节点不存在"))
		return
	}
	utils.OK(c, gin.H{"config": services.GenerateFrpcConfig(db, &t, user, &node), "tunnel_id": t.ID}, "ok")
}

// clientPublicAddress 客户端视角的公网地址。
func clientPublicAddress(t *models.Tunnel) string {
	if t.Type == "http" || t.Type == "https" {
		domain := ""
		if t.CustomDomain != nil && *t.CustomDomain != "" {
			domain = *t.CustomDomain
		} else if t.Subdomain != nil && *t.Subdomain != "" {
			domain = *t.Subdomain
		}
		if domain != "" {
			return t.Type + "://" + domain
		}
		if t.RemotePort != nil {
			return t.Type + "://" + t.Node.Address + ":" + itoaInt(*t.RemotePort)
		}
	}
	if t.RemotePort != nil {
		return t.Node.Address + ":" + itoaInt(*t.RemotePort)
	}
	if t.Node.Address != "" {
		return t.Node.Address
	}
	return config.C.PanelBaseURL
}

func itoaInt(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
