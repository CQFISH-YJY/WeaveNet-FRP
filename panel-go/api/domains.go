package api

import (
	"strconv"

	"weavenet/panel/config"
	"weavenet/panel/database"
	"weavenet/panel/middleware"
	"weavenet/panel/models"
	"weavenet/panel/services"
	"weavenet/panel/utils"

	"github.com/gin-gonic/gin"
)

// ListDomains 免费域名列表。
func ListDomains(c *gin.Context) {
	user := middleware.GetUser(c)
	var list []models.Domain
	database.DB.Where("user_id = ?", user.ID).Order("id desc").Find(&list)
	items := make([]gin.H, 0, len(list))
	for _, d := range list {
		items = append(items, serializeDomain(&d))
	}
	utils.OK(c, items, "ok")
}

// CreateDomain 申请免费域名。
func CreateDomain(c *gin.Context) {
	var p struct {
		Subdomain string `json:"subdomain"`
		TunnelID  *uint  `json:"tunnel_id"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.Biz(400, 0, "参数错误"))
		return
	}
	if len(p.Subdomain) < 3 || len(p.Subdomain) > 128 {
		utils.Fail(c, utils.Biz(400, 0, "子域名长度需在 3-128 位之间"))
		return
	}
	user := middleware.GetUser(c)
	db := database.DB
	var plan models.Plan
	if err := db.First(&plan, user.PlanID).Error; err != nil {
		utils.Fail(c, utils.Biz(400, 2001, "套餐配置异常，请联系管理员"))
		return
	}
	if err := services.CheckDomainLimit(db, user, &plan); err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	if !services.CheckSubdomainAvailable(db, p.Subdomain, 0) {
		utils.Fail(c, utils.Conflict(3002, "子域名已被占用，请更换"))
		return
	}
	fullDomain := p.Subdomain + "." + config.C.DomainSuffix
	d := models.Domain{
		UserID: user.ID, TunnelID: p.TunnelID, Subdomain: p.Subdomain,
		FullDomain: fullDomain, Status: "active",
	}
	if p.TunnelID != nil {
		var t models.Tunnel
		if err := db.Where("id = ? AND user_id = ?", *p.TunnelID, user.ID).First(&t).Error; err != nil {
			utils.Fail(c, utils.NotFound("隧道不存在"))
			return
		}
		if t.Type != "http" && t.Type != "https" {
			utils.Fail(c, utils.Biz(400, 0, "仅 HTTP/HTTPS 隧道可绑定免费域名"))
			return
		}
		sub := p.Subdomain
		t.Subdomain = &sub
		db.Save(&t)
	}
	if err := db.Create(&d).Error; err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	utils.OK(c, gin.H{"id": d.ID, "subdomain": d.Subdomain, "full_domain": d.FullDomain}, "免费域名申请成功")
}

// ReleaseDomain 释放免费域名。
func ReleaseDomain(c *gin.Context) {
	user := middleware.GetUser(c)
	id, err := strconv.Atoi(c.Param("domain_id"))
	if err != nil {
		utils.Fail(c, utils.NotFound("域名不存在"))
		return
	}
	db := database.DB
	var d models.Domain
	if err := db.Where("id = ? AND user_id = ?", id, user.ID).First(&d).Error; err != nil {
		utils.Fail(c, utils.NotFound("域名不存在"))
		return
	}
	d.Status = "released"
	if d.TunnelID != nil {
		var t models.Tunnel
		if err := db.First(&t, *d.TunnelID).Error; err == nil && t.Subdomain != nil && *t.Subdomain == d.Subdomain {
			t.Subdomain = nil
			db.Save(&t)
		}
	}
	db.Save(&d)
	utils.NoContent(c)
}

func serializeDomain(d *models.Domain) gin.H {
	return gin.H{
		"id": d.ID, "user_id": d.UserID, "tunnel_id": d.TunnelID, "subdomain": d.Subdomain,
		"full_domain": d.FullDomain, "status": d.Status,
		"created_at": utils.TimeFmtV(d.CreatedAt),
	}
}
