package api

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"weavenet/panel/cache"
	"weavenet/panel/config"
	"weavenet/panel/database"
	"weavenet/panel/middleware"
	"weavenet/panel/models"
	"weavenet/panel/utils"
)

func adminLog(c *gin.Context, admin *models.User, action, targetType string, targetID *uint, detail string) {
	db := database.DB
	op := models.OperationLog{
		AdminID: &admin.ID, AdminName: admin.Username, Action: action,
		TargetType: targetType, TargetID: targetID, Detail: detail, IP: utils.ClientIP(c),
	}
	db.Create(&op)
}

func parsePage(c *gin.Context, defSize int) (int, int) {
	page, _ := utils.AtoiSafe(c.DefaultQuery("page", "1"), 1)
	size, _ := utils.AtoiSafe(c.DefaultQuery("page_size", strconv.Itoa(defSize)), defSize)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = defSize
	}
	return page, size
}

// ---------- 用户管理 ----------

// AdminListUsers 用户列表。
func AdminListUsers(c *gin.Context) {
	db := database.DB
	page, pageSize := parsePage(c, 20)
	keyword := c.Query("keyword")
	status := c.Query("status")
	q := db.Model(&models.User{})
	if keyword != "" {
		q = q.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	q.Count(&total)
	var list []models.User
	q.Preload("Plan").Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	items := make([]gin.H, 0, len(list))
	for _, u := range list {
		planName := ""
		if u.Plan.ID > 0 {
			planName = u.Plan.Name
		}
		items = append(items, gin.H{
			"id": u.ID, "username": u.Username, "email": u.Email,
			"email_verified": u.EmailVerified, "status": u.Status, "points": u.Points,
			"plan_name": planName, "plan_expires_at": utils.TimeFmt(u.PlanExpiresAt),
			"created_at": utils.TimeFmtV(u.CreatedAt), "last_login_at": utils.TimeFmt(u.LastLoginAt),
		})
	}
	utils.OK(c, gin.H{"total": total, "page": page, "page_size": pageSize, "items": items}, "ok")
}

// AdminBanUser 封禁用户。
func AdminBanUser(c *gin.Context) {
	admin := middleware.GetUser(c)
	user, be := adminGetUser(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	var p struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&p)
	user.Status = "banned"
	detail := p.Reason
	if detail == "" {
		detail = "管理员封禁"
	}
	adminLog(c, admin, "ban_user", "user", &user.ID, detail)
	database.DB.Save(user)
	utils.OKMsg(c, "已封禁用户 "+user.Username)
}

// AdminUnbanUser 解除封禁。
func AdminUnbanUser(c *gin.Context) {
	admin := middleware.GetUser(c)
	user, be := adminGetUser(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	user.Status = "active"
	adminLog(c, admin, "unban_user", "user", &user.ID, "")
	database.DB.Save(user)
	utils.OKMsg(c, "已解除封禁 "+user.Username)
}

// AdminResetPassword 重置密码（生成随机临时密码）。
func AdminResetPassword(c *gin.Context) {
	admin := middleware.GetUser(c)
	user, be := adminGetUser(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	temp := utils.GenToken()[:12]
	hash, err := utils.HashPassword(temp)
	if err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	user.PasswordHash = hash
	adminLog(c, admin, "reset_password", "user", &user.ID, "重置用户密码")
	database.DB.Save(user)
	utils.OK(c, gin.H{"temp_password": temp}, "密码已重置，请将临时密码告知用户")
}

// AdminUpdateUserPlan 调整用户套餐或积分。
func AdminUpdateUserPlan(c *gin.Context) {
	admin := middleware.GetUser(c)
	user, be := adminGetUser(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	var p struct {
		PlanID *uint `json:"plan_id"`
		Points *int  `json:"points"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.Biz(400, 0, "参数错误"))
		return
	}
	db := database.DB
	if p.PlanID != nil {
		var plan models.Plan
		if err := db.First(&plan, *p.PlanID).Error; err != nil {
			utils.Fail(c, utils.NotFound("套餐不存在"))
			return
		}
		user.PlanID = plan.ID
		exp := time.Now().Add(30 * 24 * time.Hour)
		user.PlanExpiresAt = &exp
		db.Create(&models.UserPlanLog{UserID: user.ID, PlanID: plan.ID, PlanName: plan.Name, Reason: "管理员调整"})
		adminLog(c, admin, "update_user_plan", "user", &user.ID, "调整套餐为 "+plan.Name)
	}
	if p.Points != nil {
		user.Points = *p.Points
		adminLog(c, admin, "update_user_points", "user", &user.ID, "调整积分为 "+strconv.Itoa(*p.Points))
	}
	db.Save(user)
	utils.OKMsg(c, "用户信息已更新")
}

func adminGetUser(c *gin.Context) (*models.User, *utils.BizError) {
	id, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return nil, utils.NotFound("用户不存在")
	}
	var u models.User
	if err := database.DB.First(&u, id).Error; err != nil {
		return nil, utils.NotFound("用户不存在")
	}
	return &u, nil
}

// ---------- 节点管理 ----------

func serializeNode(n *models.Node) gin.H {
	return gin.H{
		"id": n.ID, "name": n.Name, "address": n.Address, "port": n.Port,
		"status": n.Status, "speed_limit_mbps": n.SpeedLimitMbps,
		"agent_token": n.AgentToken, "remark": n.Remark,
		"last_heartbeat_at": utils.TimeFmt(n.LastHeartbeatAt),
		"created_at":        utils.TimeFmtV(n.CreatedAt),
	}
}

// AdminListNodes 节点列表。
func AdminListNodes(c *gin.Context) {
	admin := middleware.GetUser(c)
	var list []models.Node
	database.DB.Order("id").Find(&list)
	items := make([]gin.H, 0, len(list))
	for _, n := range list {
		items = append(items, serializeNode(&n))
	}
	_ = admin
	utils.OK(c, items, "ok")
}

// AdminCreateNode 新增节点。
func AdminCreateNode(c *gin.Context) {
	admin := middleware.GetUser(c)
	var p struct {
		Name           string `json:"name"`
		Address        string `json:"address"`
		Port           int    `json:"port"`
		SpeedLimitMbps int    `json:"speed_limit_mbps"`
		Remark         string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&p); err != nil || p.Name == "" || p.Address == "" {
		utils.Fail(c, utils.Biz(400, 0, "名称和地址不能为空"))
		return
	}
	db := database.DB
	var count int64
	db.Model(&models.Node{}).Where("name = ?", p.Name).Count(&count)
	if count > 0 {
		utils.Fail(c, utils.Conflict(0, "节点名称已存在"))
		return
	}
	speed := p.SpeedLimitMbps
	if speed < 1 {
		speed = 100
	}
	node := models.Node{
		Name: p.Name, Address: p.Address, Port: p.Port, Status: "offline",
		SpeedLimitMbps: speed, AgentToken: utils.GenAgentToken(), Remark: p.Remark,
	}
	if node.Port <= 0 {
		node.Port = 7000
	}
	adminLog(c, admin, "create_node", "node", nil, "新增节点 "+p.Name)
	if err := db.Create(&node).Error; err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	utils.OK(c, serializeNode(&node), "节点创建成功")
}

// AdminUpdateNode 修改节点。
func AdminUpdateNode(c *gin.Context) {
	admin := middleware.GetUser(c)
	node, be := adminGetNode(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	var p struct {
		Name           *string `json:"name"`
		Address        *string `json:"address"`
		Port           *int    `json:"port"`
		SpeedLimitMbps *int    `json:"speed_limit_mbps"`
		Remark         *string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.Biz(400, 0, "参数错误"))
		return
	}
	if p.Name != nil {
		node.Name = *p.Name
	}
	if p.Address != nil {
		node.Address = *p.Address
	}
	if p.Port != nil {
		node.Port = *p.Port
	}
	if p.SpeedLimitMbps != nil {
		node.SpeedLimitMbps = *p.SpeedLimitMbps
	}
	if p.Remark != nil {
		node.Remark = *p.Remark
	}
	adminLog(c, admin, "update_node", "node", &node.ID, "修改节点 "+node.Name)
	database.DB.Save(node)
	utils.OK(c, serializeNode(node), "节点已更新")
}

// AdminDeleteNode 删除节点。
func AdminDeleteNode(c *gin.Context) {
	admin := middleware.GetUser(c)
	node, be := adminGetNode(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	adminLog(c, admin, "delete_node", "node", &node.ID, "删除节点 "+node.Name)
	database.DB.Delete(node)
	utils.NoContent(c)
}

// AdminStartNode 启用节点。
func AdminStartNode(c *gin.Context) {
	admin := middleware.GetUser(c)
	node, be := adminGetNode(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	node.Status = "online"
	adminLog(c, admin, "start_node", "node", &node.ID, "启用节点 "+node.Name)
	database.DB.Save(node)
	utils.OK(c, serializeNode(node), "节点已启用")
}

// AdminStopNode 停用节点。
func AdminStopNode(c *gin.Context) {
	admin := middleware.GetUser(c)
	node, be := adminGetNode(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	node.Status = "maintenance"
	adminLog(c, admin, "stop_node", "node", &node.ID, "停用节点 "+node.Name)
	database.DB.Save(node)
	utils.OK(c, serializeNode(node), "节点已停用")
}

// AdminNodeSpeed 节点限速配置。
func AdminNodeSpeed(c *gin.Context) {
	admin := middleware.GetUser(c)
	node, be := adminGetNode(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	var p struct {
		SpeedLimitMbps int `json:"speed_limit_mbps"`
	}
	if err := c.ShouldBindJSON(&p); err != nil || p.SpeedLimitMbps < 1 {
		utils.Fail(c, utils.Biz(400, 0, "限速值必须为正整数"))
		return
	}
	node.SpeedLimitMbps = p.SpeedLimitMbps
	adminLog(c, admin, "update_node_speed", "node", &node.ID, "节点限速调整为 "+strconv.Itoa(p.SpeedLimitMbps)+"Mbps")
	database.DB.Save(node)
	utils.OK(c, serializeNode(node), "节点限速已更新")
}

func adminGetNode(c *gin.Context) (*models.Node, *utils.BizError) {
	id, err := strconv.Atoi(c.Param("node_id"))
	if err != nil {
		return nil, utils.NotFound("节点不存在")
	}
	var n models.Node
	if err := database.DB.First(&n, id).Error; err != nil {
		return nil, utils.NotFound("节点不存在")
	}
	return &n, nil
}

// ---------- 全局隧道 ----------

// AdminListTunnels 全局隧道列表。
func AdminListTunnels(c *gin.Context) {
	db := database.DB
	page, pageSize := parsePage(c, 20)
	keyword := c.Query("keyword")
	nodeID := c.Query("node_id")
	status := c.Query("status")
	q := db.Model(&models.Tunnel{}).Joins("JOIN users ON users.id = tunnels.user_id")
	if keyword != "" {
		q = q.Where("tunnels.name LIKE ? OR tunnels.local_ip LIKE ? OR users.username LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if nodeID != "" {
		q = q.Where("tunnels.node_id = ?", nodeID)
	}
	if status != "" {
		q = q.Where("tunnels.status = ?", status)
	}
	var total int64
	q.Count(&total)
	var list []models.Tunnel
	q.Preload("User").Preload("Node").Order("tunnels.id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	items := make([]gin.H, 0, len(list))
	for _, t := range list {
		username := ""
		if t.User.ID > 0 {
			username = t.User.Username
		}
		nodeName := ""
		if t.Node.ID > 0 {
			nodeName = t.Node.Name
		}
		items = append(items, gin.H{
			"id": t.ID, "name": t.Name, "type": t.Type, "username": username,
			"node_name": nodeName, "local_ip": t.LocalIP, "local_port": t.LocalPort,
			"remote_port": t.RemotePort, "subdomain": t.Subdomain, "status": t.Status,
			"created_at": utils.TimeFmtV(t.CreatedAt),
		})
	}
	utils.OK(c, gin.H{"total": total, "page": page, "page_size": pageSize, "items": items}, "ok")
}

// AdminForceOfflineTunnel 强制下线隧道。
func AdminForceOfflineTunnel(c *gin.Context) {
	admin := middleware.GetUser(c)
	id, err := strconv.Atoi(c.Param("tunnel_id"))
	if err != nil {
		utils.Fail(c, utils.NotFound("隧道不存在"))
		return
	}
	db := database.DB
	var t models.Tunnel
	if err := db.First(&t, id).Error; err != nil {
		utils.Fail(c, utils.NotFound("隧道不存在"))
		return
	}
	t.Status = "stopped"
	t.StatusDetail = "管理员强制下线"
	cache.C.Set(cache.WantKey(t.ID), gin.H{"action": "stop", "force": true}, time.Hour)
	cache.C.Del(cache.RuntimeKey(t.ID))
	adminLog(c, admin, "force_offline_tunnel", "tunnel", &t.ID, "强制下线隧道 "+t.Name)
	db.Save(&t)
	utils.OKMsg(c, "隧道已强制下线")
}

// ---------- 套餐配置 ----------

// AdminListPlans 套餐列表。
func AdminListPlans(c *gin.Context) {
	admin := middleware.GetUser(c)
	var list []models.Plan
	database.DB.Order("sort").Find(&list)
	items := make([]gin.H, 0, len(list))
	for _, p := range list {
		items = append(items, gin.H{
			"id": p.ID, "name": p.Name, "speed_limit_mbps": p.SpeedLimitMbps,
			"tunnel_limit": p.TunnelLimit, "domain_limit": p.DomainLimit, "sort": p.Sort,
		})
	}
	_ = admin
	utils.OK(c, items, "ok")
}

// AdminUpdatePlan 修改套餐档位。
func AdminUpdatePlan(c *gin.Context) {
	admin := middleware.GetUser(c)
	id, err := strconv.Atoi(c.Param("plan_id"))
	if err != nil {
		utils.Fail(c, utils.NotFound("套餐不存在"))
		return
	}
	db := database.DB
	var plan models.Plan
	if err := db.First(&plan, id).Error; err != nil {
		utils.Fail(c, utils.NotFound("套餐不存在"))
		return
	}
	if plan.IsDefault {
		utils.Fail(c, utils.Conflict(0, "免费版为默认套餐，不可修改额度"))
		return
	}
	var p struct {
		Name           *string `json:"name"`
		SpeedLimitMbps *int    `json:"speed_limit_mbps"`
		TunnelLimit    *int    `json:"tunnel_limit"`
		DomainLimit    *int    `json:"domain_limit"`
		Sort           *int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.Biz(400, 0, "参数错误"))
		return
	}
	if p.Name != nil {
		plan.Name = *p.Name
	}
	if p.SpeedLimitMbps != nil {
		plan.SpeedLimitMbps = *p.SpeedLimitMbps
	}
	if p.TunnelLimit != nil {
		plan.TunnelLimit = *p.TunnelLimit
	}
	if p.DomainLimit != nil {
		plan.DomainLimit = *p.DomainLimit
	}
	if p.Sort != nil {
		plan.Sort = *p.Sort
	}
	adminLog(c, admin, "update_plan", "plan", &plan.ID, "修改套餐 "+plan.Name)
	db.Save(&plan)
	utils.OK(c, gin.H{
		"id": plan.ID, "name": plan.Name, "speed_limit_mbps": plan.SpeedLimitMbps,
		"tunnel_limit": plan.TunnelLimit, "domain_limit": plan.DomainLimit, "sort": plan.Sort,
	}, "套餐已更新")
}

// ---------- 公告 ----------

// AdminCreateAnnouncement 发布公告。
func AdminCreateAnnouncement(c *gin.Context) {
	admin := middleware.GetUser(c)
	var p struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Author  string `json:"author"`
	}
	if err := c.ShouldBindJSON(&p); err != nil || p.Title == "" {
		utils.Fail(c, utils.Biz(400, 0, "标题不能为空"))
		return
	}
	author := p.Author
	if author == "" {
		author = admin.Username
	}
	a := models.Announcement{Title: p.Title, Content: p.Content, Author: author, Status: "active"}
	adminLog(c, admin, "create_announcement", "announcement", nil, "发布公告《"+p.Title+"》")
	database.DB.Create(&a)
	utils.OK(c, gin.H{"id": a.ID, "title": a.Title, "author": a.Author, "created_at": utils.TimeFmtV(a.CreatedAt)}, "公告发布成功")
}

// AdminListAnnouncements 公告列表（含下线）。
func AdminListAnnouncements(c *gin.Context) {
	admin := middleware.GetUser(c)
	page, pageSize := parsePage(c, 20)
	var total int64
	database.DB.Model(&models.Announcement{}).Count(&total)
	var list []models.Announcement
	database.DB.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	items := make([]gin.H, 0, len(list))
	for _, a := range list {
		items = append(items, gin.H{
			"id": a.ID, "title": a.Title, "content": a.Content, "author": a.Author,
			"status": a.Status, "created_at": utils.TimeFmtV(a.CreatedAt),
		})
	}
	_ = admin
	utils.OK(c, gin.H{"total": total, "page": page, "page_size": pageSize, "items": items}, "ok")
}

// AdminUpdateAnnouncement 修改公告。
func AdminUpdateAnnouncement(c *gin.Context) {
	admin := middleware.GetUser(c)
	a, be := adminGetAnnouncement(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	var p struct {
		Title   *string `json:"title"`
		Content *string `json:"content"`
		Author  *string `json:"author"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.Biz(400, 0, "参数错误"))
		return
	}
	if p.Title != nil {
		a.Title = *p.Title
	}
	if p.Content != nil {
		a.Content = *p.Content
	}
	if p.Author != nil {
		a.Author = *p.Author
	}
	adminLog(c, admin, "update_announcement", "announcement", &a.ID, "修改公告《"+a.Title+"》")
	database.DB.Save(a)
	utils.OKMsg(c, "公告已更新")
}

// AdminOfflineAnnouncement 公告下线。
func AdminOfflineAnnouncement(c *gin.Context) {
	admin := middleware.GetUser(c)
	a, be := adminGetAnnouncement(c)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	a.Status = "offline"
	adminLog(c, admin, "offline_announcement", "announcement", &a.ID, "下线公告《"+a.Title+"》")
	database.DB.Save(a)
	utils.OKMsg(c, "公告已下线")
}

func adminGetAnnouncement(c *gin.Context) (*models.Announcement, *utils.BizError) {
	id, err := strconv.Atoi(c.Param("announcement_id"))
	if err != nil {
		return nil, utils.NotFound("公告不存在")
	}
	var a models.Announcement
	if err := database.DB.First(&a, id).Error; err != nil {
		return nil, utils.NotFound("公告不存在")
	}
	return &a, nil
}

// ---------- 系统配置 ----------

// AdminGetConfig 获取系统配置。
func AdminGetConfig(c *gin.Context) {
	admin := middleware.GetUser(c)
	var list []models.SystemConfig
	database.DB.Find(&list)
	m := make(map[string]string, len(list))
	for _, x := range list {
		m[x.Key] = x.Value
	}
	_ = admin
	utils.OK(c, m, "ok")
}

// AdminUpdateConfig 修改系统配置。
func AdminUpdateConfig(c *gin.Context) {
	admin := middleware.GetUser(c)
	var p struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&p); err != nil || p.Key == "" {
		utils.Fail(c, utils.Biz(400, 0, "参数错误"))
		return
	}
	db := database.DB
	var row models.SystemConfig
	if err := db.First(&row, "key = ?", p.Key).Error; err != nil {
		row = models.SystemConfig{Key: p.Key, Value: p.Value}
		db.Create(&row)
	} else {
		row.Value = p.Value
		db.Save(&row)
	}
	adminLog(c, admin, "update_config", "config", nil, "修改配置 "+p.Key+"="+p.Value)
	utils.OKMsg(c, "配置已保存")
}

// ---------- 日志 ----------

// AdminOperationLogs 操作日志。
func AdminOperationLogs(c *gin.Context) {
	admin := middleware.GetUser(c)
	page, pageSize := parsePage(c, 50)
	var total int64
	database.DB.Model(&models.OperationLog{}).Count(&total)
	var list []models.OperationLog
	database.DB.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	items := make([]gin.H, 0, len(list))
	for _, x := range list {
		items = append(items, gin.H{
			"id": x.ID, "admin_name": x.AdminName, "action": x.Action,
			"target_type": x.TargetType, "target_id": x.TargetID, "detail": x.Detail,
			"ip": x.IP, "created_at": utils.TimeFmtV(x.CreatedAt),
		})
	}
	_ = admin
	utils.OK(c, gin.H{"total": total, "page": page, "page_size": pageSize, "items": items}, "ok")
}

// AdminRuntimeLogs 运行日志（从本地日志文件读取尾部）。
func AdminRuntimeLogs(c *gin.Context) {
	admin := middleware.GetUser(c)
	lines, _ := utils.AtoiSafe(c.DefaultQuery("lines", "200"), 200)
	if lines < 0 {
		lines = 200
	}
	logFile := filepath.Join(config.C.LogsDir, "weavenet.log")
	items := []string{}
	if data, err := os.ReadFile(logFile); err == nil {
		all := strings.Split(string(data), "\n")
		if len(all) > lines {
			all = all[len(all)-lines:]
		}
		items = all
	}
	_ = admin
	utils.OK(c, gin.H{"lines": items, "file": logFile}, "ok")
}

// ---------- 统计看板 ----------

// AdminDashboard 统计看板。
func AdminDashboard(c *gin.Context) {
	admin := middleware.GetUser(c)
	db := database.DB
	days, _ := utils.AtoiSafe(c.DefaultQuery("days", "7"), 7)
	if days < 1 {
		days = 1
	}
	if days > 31 {
		days = 31
	}
	start := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")

	var totalUsers, newUsersWeek, runningTunnels, totalTunnels int64
	var onlineNodes, totalNodes int64
	db.Model(&models.User{}).Count(&totalUsers)
	db.Model(&models.User{}).Where("created_at >= ?", start).Count(&newUsersWeek)
	db.Model(&models.Node{}).Where("status = ?", "online").Count(&onlineNodes)
	db.Model(&models.Node{}).Count(&totalNodes)
	db.Model(&models.Tunnel{}).Where("status = ?", "running").Count(&runningTunnels)
	db.Model(&models.Tunnel{}).Count(&totalTunnels)

	type row struct {
		Date     string
		InBytes  int64
		OutBytes int64
	}
	var rows []row
	db.Model(&models.TrafficStat{}).Select("date, SUM(in_bytes) as in_bytes, SUM(out_bytes) as out_bytes").
		Where("date >= ?", start).Group("date").Scan(&rows)
	trafficMap := make(map[string][2]int64)
	for _, r := range rows {
		trafficMap[r.Date] = [2]int64{r.InBytes, r.OutBytes}
	}
	trafficSeries := make([]gin.H, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		v := trafficMap[d]
		trafficSeries = append(trafficSeries, gin.H{"date": d, "in_bytes": v[0], "out_bytes": v[1]})
	}

	var pointsIssued int64
	db.Model(&models.PointsLog{}).Where("change > 0").Select("SUM(change)").Scan(&pointsIssued)
	_ = admin
	utils.OK(c, gin.H{
		"summary": gin.H{
			"total_users": totalUsers, "new_users_week": newUsersWeek,
			"online_nodes": onlineNodes, "total_nodes": totalNodes,
			"running_tunnels": runningTunnels, "total_tunnels": totalTunnels,
			"points_issued": pointsIssued,
		},
		"traffic_series": trafficSeries,
	}, "ok")
}
