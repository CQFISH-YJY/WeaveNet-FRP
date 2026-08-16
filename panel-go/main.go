package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"

	"weavenet/panel/api"
	"weavenet/panel/config"
	"weavenet/panel/database"
	"weavenet/panel/middleware"
	"weavenet/panel/models"
	"weavenet/panel/services"
)

var templates *pongo2.TemplateSet

func main() {
	cfg := config.Load()
	if err := database.Init(cfg.DBPath); err != nil {
		log.Fatalf("[db] 初始化失败: %v", err)
	}
	seedDefaults(database.DB)

	services.StartMailWorker()
	services.StartScheduler()

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())

	// 官网模板渲染初始化（LocalFileSystemLoader 相对 base.html 解析 extends）
	if dir, err := filepath.Abs(cfg.TemplatesDir); err == nil {
		if _, err := os.Stat(dir); err == nil {
			if loader, err := pongo2.NewLocalFileSystemLoader(dir); err == nil {
				templates = pongo2.NewSet("site", loader)
			}
		}
	}

	registerAPIRoutes(r)
	registerSiteRoutes(r)
	registerStaticRoutes(r)

	log.Printf("[main] %s 已启动，监听 %s:%d", cfg.AppName, cfg.PanelHost, cfg.PanelPort)
	addr := cfg.PanelHost + ":" + strconv.Itoa(cfg.PanelPort)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[main] 启动失败: %v", err)
	}
}

func registerAPIRoutes(r *gin.Engine) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", middleware.RateLimit("register", 5, time.Minute), api.Register)
		auth.POST("/email-verify", api.EmailVerify)
		auth.POST("/resend-code", middleware.RateLimit("resend_code", 3, time.Minute), api.ResendCode)
		auth.POST("/login", middleware.RateLimit("login", 10, time.Minute), api.Login)
		auth.POST("/logout", middleware.CurrentUser(), api.Logout)
		auth.POST("/forgot-password", middleware.RateLimit("forgot_password", 5, time.Minute), api.ForgotPassword)
		auth.POST("/reset-password", api.ResetPassword)
	}

	user := r.Group("/api/user", middleware.CurrentUser())
	{
		user.GET("/profile", api.GetProfile)
		user.PUT("/profile", api.UpdateProfile)
		user.POST("/send-email-code", api.SendEmailCode)
		user.POST("/change-email", api.ChangeEmail)
		user.POST("/change-password", api.ChangePassword)
		user.GET("/quota", api.GetQuota)
		user.GET("/logs", api.GetUserLogs)
	}

	tunnels := r.Group("/api/tunnels", middleware.CurrentUser())
	{
		tunnels.GET("", api.ListTunnels)
		tunnels.POST("", api.CreateTunnel)
		tunnels.GET("/:tunnel_id", api.GetTunnel)
		tunnels.PUT("/:tunnel_id", api.UpdateTunnel)
		tunnels.DELETE("/:tunnel_id", api.DeleteTunnel)
		tunnels.POST("/:tunnel_id/start", api.StartTunnel)
		tunnels.POST("/:tunnel_id/stop", api.StopTunnel)
		tunnels.GET("/:tunnel_id/config", api.GetTunnelConfig)
		tunnels.GET("/:tunnel_id/status", api.GetTunnelStatus)
	}

	nodes := r.Group("/api/nodes")
	{
		nodes.GET("", api.ListNodes)
		nodes.GET("/:node_id/status", api.GetNodeStatus)
	}

	domains := r.Group("/api/domains", middleware.CurrentUser())
	{
		domains.GET("", api.ListDomains)
		domains.POST("", api.CreateDomain)
		domains.DELETE("/:domain_id", api.ReleaseDomain)
	}

	signin := r.Group("/api/signin", middleware.CurrentUser())
	{
		signin.POST("", middleware.RateLimit("signin", 3, time.Minute), api.DoSignin)
		signin.GET("/status", api.GetSigninStatus)
	}

	points := r.Group("/api/points", middleware.CurrentUser())
	{
		points.GET("/logs", api.ListPointsLogs)
		points.POST("/exchange", api.ExchangePoints)
		points.GET("/rules", api.GetPointsRules)
	}

	stats := r.Group("/api/stats", middleware.CurrentUser())
	{
		stats.GET("/traffic", api.GetTrafficStats)
		stats.GET("/overview", api.GetOverview)
	}

	announcements := r.Group("/api/announcements")
	{
		announcements.GET("", api.ListAnnouncements)
		announcements.GET("/:announcement_id", api.GetAnnouncement)
	}

	tickets := r.Group("/api/tickets", middleware.CurrentUser())
	{
		tickets.POST("", api.CreateTicket)
		tickets.GET("", api.ListTickets)
		tickets.GET("/:ticket_id", api.TicketDetail)
		tickets.POST("/:ticket_id/reply", api.ReplyTicket)
		tickets.POST("/:ticket_id/close", api.CloseTicket)
	}

	agent := r.Group("/api/agent")
	{
		agent.POST("/register", api.AgentRegister)
		agent.POST("/heartbeat", middleware.AgentNode(), api.AgentHeartbeat)
		agent.GET("/tunnels", middleware.AgentNode(), api.AgentTunnels)
	}

	client := r.Group("/api/client")
	{
		client.POST("/login", middleware.RateLimit("client_login", 10, time.Minute), api.ClientLogin)
		client.GET("/tunnels", middleware.CurrentUser(), api.ClientTunnels)
		client.POST("/config", middleware.CurrentUser(), api.ClientConfig)
	}

	adminUsers := r.Group("/api/admin/users", middleware.CurrentUser(), middleware.CurrentAdmin())
	{
		adminUsers.GET("", api.AdminListUsers)
		adminUsers.POST("/:user_id/ban", api.AdminBanUser)
		adminUsers.POST("/:user_id/unban", api.AdminUnbanUser)
		adminUsers.POST("/:user_id/reset-password", api.AdminResetPassword)
		adminUsers.PUT("/:user_id/plan", api.AdminUpdateUserPlan)
	}

	adminNodes := r.Group("/api/admin/nodes", middleware.CurrentUser(), middleware.CurrentAdmin())
	{
		adminNodes.GET("", api.AdminListNodes)
		adminNodes.POST("", api.AdminCreateNode)
		adminNodes.PUT("/:node_id", api.AdminUpdateNode)
		adminNodes.DELETE("/:node_id", api.AdminDeleteNode)
		adminNodes.POST("/:node_id/start", api.AdminStartNode)
		adminNodes.POST("/:node_id/stop", api.AdminStopNode)
		adminNodes.PUT("/:node_id/speed", api.AdminNodeSpeed)
	}

	admin := r.Group("/api/admin", middleware.CurrentUser(), middleware.CurrentAdmin())
	{
		admin.GET("/tunnels", api.AdminListTunnels)
		admin.POST("/tunnels/:tunnel_id/offline", api.AdminForceOfflineTunnel)
		admin.GET("/plans", api.AdminListPlans)
		admin.PUT("/plans/:plan_id", api.AdminUpdatePlan)
		admin.GET("/announcements", api.AdminListAnnouncements)
		admin.POST("/announcements", api.AdminCreateAnnouncement)
		admin.PUT("/announcements/:announcement_id", api.AdminUpdateAnnouncement)
		admin.POST("/announcements/:announcement_id/offline", api.AdminOfflineAnnouncement)
		admin.GET("/config", api.AdminGetConfig)
		admin.PUT("/config", api.AdminUpdateConfig)
		admin.GET("/logs/operation", api.AdminOperationLogs)
		admin.GET("/logs/runtime", api.AdminRuntimeLogs)
		admin.GET("/dashboard", api.AdminDashboard)
	}

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"engine": "go", "cache": "memory"}})
	})
}

// ---------- 官网页面 ----------

func registerSiteRoutes(r *gin.Engine) {
	r.GET("/", siteHandler("index.html"))
	r.GET("/download", siteHandler("download.html"))
	r.GET("/docs", siteHandler("docs.html"))
	r.GET("/pricing", siteHandler("pricing.html"))
	r.GET("/about", siteHandler("about.html"))
	r.GET("/terms", siteHandler("terms.html"))
	r.GET("/privacy", siteHandler("privacy.html"))
	r.GET("/changelog", siteHandler("changelog.html"))
	r.GET("/help", siteHandler("help.html"))
	r.GET("/status", siteHandler("status.html"))
	r.GET("/announcements", siteHandler("announcements.html"))
	r.GET("/announcements/:announcement_id", siteAnnouncementDetail)
}

// siteHandler 静态官网页面（公共上下文：站点名 + 最新公告）。
func siteHandler(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := pongo2.Context{
			"site_name": config.C.AppName,
			"current":   c.Request.URL.Path,
		}
		switch name {
		case "index.html":
			ctx["announcements"] = announcementList(5)
			nodes, _ := nodeList()
			ctx["node_count"] = len(nodes)
		case "announcements.html":
			ctx["announcement_items"] = announcementList(100)
		case "status.html":
			nodes, online := nodeList()
			ctx["nodes"] = nodes
			ctx["node_count"] = len(nodes)
			ctx["online_count"] = online
			uptime := "99.9"
			if len(nodes) > 0 {
				uptime = strconv.FormatFloat(float64(online)/float64(len(nodes))*100, 'f', 1, 64)
			}
			ctx["uptime"] = uptime
		}
		renderSite(c, name, ctx)
	}
}

// siteAnnouncementDetail 公告详情页。
func siteAnnouncementDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("announcement_id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/announcements")
		return
	}
	var a models.Announcement
	if err := database.DB.First(&a, id).Error; err != nil || a.Status != "active" {
		c.Redirect(http.StatusFound, "/announcements")
		return
	}
	renderSite(c, "announcement.html", pongo2.Context{
		"site_name": config.C.AppName,
		"current":   c.Request.URL.Path,
		"announcement": map[string]any{
			"title":           a.Title,
			"author":          a.Author,
			"content":         a.Content,
			"created_at_text": a.CreatedAt.Format("2006-01-02 15:04"),
		},
	})
}

func announcementList(limit int) []map[string]any {
	var list []models.Announcement
	database.DB.Where("status = ?", "active").Order("id desc").Limit(limit).Find(&list)
	out := make([]map[string]any, 0, len(list))
	for _, a := range list {
		excerpt := a.Content
		if len(excerpt) > 120 {
			excerpt = excerpt[:120] + "..."
		}
		out = append(out, map[string]any{
			"id":              a.ID,
			"title":           a.Title,
			"author":          a.Author,
			"content":         a.Content,
			"excerpt":         excerpt,
			"created_at_text": a.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	return out
}

func nodeList() ([]map[string]any, int) {
	var list []models.Node
	database.DB.Order("id").Find(&list)
	online := 0
	out := make([]map[string]any, 0, len(list))
	for _, n := range list {
		if n.Status == "online" {
			online++
		}
		hb := ""
		if n.LastHeartbeatAt != nil {
			hb = n.LastHeartbeatAt.Format("2006-01-02 15:04")
		}
		out = append(out, map[string]any{
			"name":                n.Name,
			"address":             n.Address,
			"port":                n.Port,
			"status":              n.Status,
			"speed_limit_mbps":    n.SpeedLimitMbps,
			"last_heartbeat_text": hb,
		})
	}
	return out, online
}

func renderSite(c *gin.Context, name string, ctx pongo2.Context) {
	tmpl, err := templates.FromFile(name)
	if err != nil {
		log.Printf("[site] 模板 %s 加载失败: %v", name, err)
		c.String(http.StatusInternalServerError, "模板加载失败")
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteWriter(ctx, c.Writer); err != nil {
		log.Printf("[site] 模板 %s 渲染失败: %v", name, err)
	}
}

// ---------- 静态资源与 SPA ----------

func registerStaticRoutes(r *gin.Engine) {
	if dir, err := filepath.Abs(config.C.StaticDir); err == nil {
		if _, err := os.Stat(dir); err == nil {
			r.Static("/static", dir)
		}
	}
	spaDir, err := filepath.Abs(config.C.WebDistDir)
	if err != nil || !dirExists(spaDir) {
		return
	}
	// /panel 前缀 SPA 托管：未命中回退 index.html（支持 hash 路由深链）
	r.GET("/panel/*any", func(c *gin.Context) {
		path := strings.TrimPrefix(c.Param("any"), "/")
		if path == "" || path == "index.html" {
			c.File(filepath.Join(spaDir, "index.html"))
			return
		}
		full := filepath.Join(spaDir, filepath.FromSlash(path))
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			c.File(full)
			return
		}
		c.File(filepath.Join(spaDir, "index.html"))
	})
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
