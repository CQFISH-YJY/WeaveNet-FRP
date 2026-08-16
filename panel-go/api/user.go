package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"weavenet/panel/cache"
	"weavenet/panel/database"
	"weavenet/panel/middleware"
	"weavenet/panel/models"
	"weavenet/panel/services"
	"weavenet/panel/utils"
)

// GetProfile 个人资料。
func GetProfile(c *gin.Context) {
	user := middleware.GetUser(c)
	utils.OK(c, gin.H{
		"id":              user.ID,
		"username":        user.Username,
		"email":           user.Email,
		"email_verified":  user.EmailVerified,
		"status":          user.Status,
		"points":          user.Points,
		"plan":            planMap(&user.Plan),
		"plan_expires_at": utils.TimeFmt(user.PlanExpiresAt),
		"created_at":      utils.TimeFmtV(user.CreatedAt),
		"last_login_at":   utils.TimeFmt(user.LastLoginAt),
	}, "ok")
}

func planMap(p *models.Plan) gin.H {
	if p == nil || p.ID == 0 {
		return gin.H{"id": 1, "name": "免费版", "speed_limit_mbps": 8, "tunnel_limit": 3, "domain_limit": 1}
	}
	return gin.H{
		"id": p.ID, "name": p.Name, "speed_limit_mbps": p.SpeedLimitMbps,
		"tunnel_limit": p.TunnelLimit, "domain_limit": p.DomainLimit,
	}
}

// UpdateProfile 修改资料（仅邮箱预填）。
func UpdateProfile(c *gin.Context) {
	var p struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.ErrBiz(422, 0, "参数错误"))
		return
	}
	db := database.DB
	user := middleware.GetUser(c)
	if p.Email != "" && p.Email != user.Email {
		var count int64
		db.Model(&models.User{}).Where("email = ?", p.Email).Count(&count)
		if count > 0 {
			utils.Fail(c, utils.ErrBiz(409, 1003, "该邮箱已被其他账号使用"))
			return
		}
		user.Email = p.Email
		user.EmailVerified = false
		db.Save(user)
	}
	utils.OKMsg(c, "资料已更新")
}

// SendEmailCode 占位接口。
func SendEmailCode(c *gin.Context) {
	utils.OKMsg(c, "请在修改邮箱时提供新邮箱与验证码")
}

// ChangeEmail 修改邮箱。
func ChangeEmail(c *gin.Context) {
	var p struct {
		NewEmail string `json:"new_email" binding:"required"`
		Code     string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.ErrBiz(422, 0, "参数错误"))
		return
	}
	db := database.DB
	user := middleware.GetUser(c)
	var count int64
	db.Model(&models.User{}).Where("email = ?", p.NewEmail).Count(&count)
	if count > 0 {
		utils.Fail(c, utils.ErrBiz(409, 1003, "该邮箱已被其他账号使用"))
		return
	}
	// 无有效验证码则先发送
	var codeRow models.EmailCode
	err := db.Where("email = ? AND purpose = ? AND used = ?", p.NewEmail, "change_email", false).Order("id desc").First(&codeRow).Error
	if err != nil || codeRow.Code != p.Code {
		code := utils.GenEmailCode()
		db.Create(&models.EmailCode{Email: p.NewEmail, Code: code, Purpose: "change_email", ExpiresAt: time.Now().Add(emailCodeTTL)})
		services.SendVerificationCode(p.NewEmail, code, "change_email")
		utils.Fail(c, utils.ErrBiz(400, 1001, "验证码已发送至新邮箱，请查收后重新提交"))
		return
	}
	if time.Now().After(codeRow.ExpiresAt) {
		utils.Fail(c, utils.ErrBiz(400, 1002, "验证码已过期"))
		return
	}
	codeRow.Used = true
	db.Save(&codeRow)
	user.Email = p.NewEmail
	user.EmailVerified = true
	db.Save(user)
	utils.OKMsg(c, "邮箱修改成功")
}

// ChangePassword 修改密码。
func ChangePassword(c *gin.Context) {
	var p struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.ErrBiz(422, 0, "参数错误"))
		return
	}
	db := database.DB
	user := middleware.GetUser(c)
	if !utils.VerifyPassword(user.PasswordHash, p.OldPassword) {
		utils.Fail(c, utils.ErrBiz(400, 0, "原密码错误"))
		return
	}
	hash, _ := utils.HashPassword(p.NewPassword)
	user.PasswordHash = hash
	db.Save(user)
	// 作废全部会话
	var sessions []models.Session
	db.Where("user_id = ?", user.ID).Find(&sessions)
	for _, s := range sessions {
		cache.C.Del("session:" + s.Token)
	}
	db.Where("user_id = ?", user.ID).Delete(&models.Session{})
	utils.OKMsg(c, "密码修改成功")
}

// GetQuota 套餐额度。
func GetQuota(c *gin.Context) {
	user := middleware.GetUser(c)
	db := database.DB
	var tunnelCount int64
	db.Model(&models.Tunnel{}).Where("user_id = ?", user.ID).Count(&tunnelCount)
	plan := user.Plan
	utils.OK(c, gin.H{
		"plan":             planMap(&plan),
		"plan_expires_at":  utils.TimeFmt(user.PlanExpiresAt),
		"tunnel_count":     tunnelCount,
		"tunnel_limit":     plan.TunnelLimit,
		"domain_count":     0,
	}, "ok")
}

// GetUserLogs 用户操作日志。
func GetUserLogs(c *gin.Context) {
	user := middleware.GetUser(c)
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	pg, _ := utils.AtoiSafe(page, 1)
	ps, _ := utils.AtoiSafe(pageSize, 10)
	db := database.DB
	var total int64
	db.Model(&models.OperationLog{}).Where("target_type = ?", "user").Where("target_id = ?", user.ID).Count(&total)
	var logs []models.OperationLog
	db.Where("target_type = ?", "user").Where("target_id = ?", user.ID).
		Order("id desc").Offset((pg - 1) * ps).Limit(ps).Find(&logs)
	items := make([]gin.H, 0, len(logs))
	for _, l := range logs {
		items = append(items, gin.H{
			"id": l.ID, "action": l.Action, "detail": l.Detail,
			"created_at": utils.TimeFmtV(l.CreatedAt),
		})
	}
	utils.OK(c, gin.H{"total": total, "page": pg, "page_size": ps, "items": items}, "ok")
}
