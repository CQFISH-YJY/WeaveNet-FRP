package main

import (
	"log"
	"strconv"

	"gorm.io/gorm"

	"weavenet/panel/config"
	"weavenet/panel/models"
	"weavenet/panel/utils"
)

// seedDefaults 写入种子数据（幂等）。
func seedDefaults(db *gorm.DB) {
	seedPlans(db)
	seedAdmin(db)
	seedConfigs(db)
	seedDemoNode(db)
}

func seedPlans(db *gorm.DB) {
	var count int64
	db.Model(&models.Plan{}).Count(&count)
	if count > 0 {
		return
	}
	plans := []models.Plan{
		{ID: 1, Name: "免费版", SpeedLimitMbps: 8, TunnelLimit: 3, DomainLimit: 1, Sort: 1, IsDefault: true},
		{ID: 2, Name: "普通会员", SpeedLimitMbps: 16, TunnelLimit: 6, DomainLimit: 4, Sort: 2},
		{ID: 3, Name: "高级会员", SpeedLimitMbps: 24, TunnelLimit: 10, DomainLimit: 8, Sort: 3},
		{ID: 4, Name: "超级会员", SpeedLimitMbps: 32, TunnelLimit: 14, DomainLimit: 16, Sort: 4},
	}
	db.Create(&plans)
	log.Println("[seed] 已写入四档套餐")
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Where("username = ?", config.C.AdminUsername).Count(&count)
	if count > 0 {
		return
	}
	hash, err := utils.HashPassword(config.C.AdminPassword)
	if err != nil {
		log.Printf("[seed] 管理员密码哈希失败: %v", err)
		return
	}
	admin := models.User{
		Username: config.C.AdminUsername, Email: config.C.AdminEmail,
		PasswordHash: hash, EmailVerified: true, Status: "active",
		PlanID: 4, Points: 99999,
	}
	if err := db.Create(&admin).Error; err != nil {
		log.Printf("[seed] 创建管理员失败: %v", err)
		return
	}
	log.Printf("[seed] 已创建管理员账号 %s", config.C.AdminUsername)
}

func seedConfigs(db *gorm.DB) {
	defaults := map[string]string{
		"signin_points":       str(config.C.SigninPoints),
		"signin_streak_bonus": str(config.C.SigninStreakBonus),
		"signin_streak_days":  str(config.C.SigninStreakDays),
		"exchange_points":     str(config.C.ExchangePoints),
		"exchange_plan_days":  str(config.C.ExchangePlanDays),
		"exchange_plan_name":  "普通会员",
		"domain_suffix":       config.C.DomainSuffix,
		"smtp_host":           config.C.SmtpHost,
		"smtp_port":           str(config.C.SmtpPort),
		"smtp_user":           config.C.SmtpUser,
		"smtp_from":           config.C.SmtpFrom,
		"smtp_use_ssl":        boolStr(config.C.SmtpUseSSL),
		"smtp_use_tls":        boolStr(config.C.SmtpUseTLS),
	}
	for key, value := range defaults {
		var count int64
		db.Model(&models.SystemConfig{}).Where("key = ?", key).Count(&count)
		if count == 0 {
			db.Create(&models.SystemConfig{Key: key, Value: value})
		}
	}
}

func seedDemoNode(db *gorm.DB) {
	var node models.Node
	err := db.Order("id").First(&node).Error
	if err != nil {
		node = models.Node{
			Name: "上海主节点", Address: "127.0.0.1", Port: 7000,
			Status: "offline", SpeedLimitMbps: 100,
			Remark: "默认演示节点，配置 Agent Token 后 frps 接入即上线",
		}
		if err := db.Create(&node).Error; err != nil {
			log.Printf("[seed] 创建演示节点失败: %v", err)
			return
		}
		log.Println("[seed] 已创建演示节点")
	}
	if node.AgentToken == "" {
		node.AgentToken = utils.GenAgentToken()
		db.Save(&node)
		log.Println("[seed] 已生成演示节点 Agent Token")
	}
}

func str(n int) string { return strconv.Itoa(n) }

func boolStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
