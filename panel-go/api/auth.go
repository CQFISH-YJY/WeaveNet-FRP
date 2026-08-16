package api

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"weavenet/panel/cache"
	"weavenet/panel/config"
	"weavenet/panel/database"
	"weavenet/panel/middleware"
	"weavenet/panel/models"
	"weavenet/panel/services"
	"weavenet/panel/utils"
)

const emailCodeTTL = 5 * time.Minute
const emailCodeMaxAttempts = 5

// Register 注册。
func Register(c *gin.Context) {
	var p struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.ErrBiz(422, 0, "参数错误: "+err.Error()))
		return
	}
	db := database.DB
	var count int64
	db.Model(&models.User{}).Where("username = ?", p.Username).Count(&count)
	if count > 0 {
		utils.Fail(c, utils.ErrBiz(409, 1004, "用户名已被注册"))
		return
	}
	db.Model(&models.User{}).Where("email = ?", p.Email).Count(&count)
	if count > 0 {
		utils.Fail(c, utils.ErrBiz(409, 1003, "该邮箱已被注册"))
		return
	}
	hash, err := utils.HashPassword(p.Password)
	if err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	user := models.User{Username: p.Username, Email: p.Email, PasswordHash: hash, PlanID: 1}
	if err := db.Create(&user).Error; err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	code := utils.GenEmailCode()
	db.Create(&models.EmailCode{Email: p.Email, Code: code, Purpose: "register", ExpiresAt: time.Now().Add(emailCodeTTL)})
	services.SendVerificationCode(p.Email, code, "register")
	c.JSON(201, gin.H{"code": 0, "message": "注册成功，请查收验证邮件激活账号", "data": gin.H{"id": user.ID, "email": user.Email}})
}

// EmailVerify 邮箱验证。
func EmailVerify(c *gin.Context) {
	var p struct {
		Email   string `json:"email" binding:"required"`
		Code    string `json:"code" binding:"required"`
		Purpose string `json:"purpose" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.ErrBiz(422, 0, "参数错误"))
		return
	}
	db := database.DB
	var codeRow models.EmailCode
	err := db.Where("email = ? AND code = ? AND purpose = ? AND used = ?",
		p.Email, p.Code, p.Purpose, false).Order("id desc").First(&codeRow).Error
	if err != nil {
		utils.Fail(c, utils.ErrBiz(400, 1001, "验证码错误"))
		return
	}
	if time.Now().After(codeRow.ExpiresAt) {
		utils.Fail(c, utils.ErrBiz(400, 1002, "验证码已过期"))
		return
	}
	codeRow.Used = true
	db.Save(&codeRow)
	if p.Purpose == "register" || p.Purpose == "change_email" {
		db.Model(&models.User{}).Where("email = ?", p.Email).Updates(map[string]any{"email_verified": true})
	}
	utils.OK(c, gin.H{"email": p.Email}, "邮箱验证成功")
}

// ResendCode 重发验证码。
func ResendCode(c *gin.Context) {
	var p struct {
		Email   string `json:"email" binding:"required"`
		Purpose string `json:"purpose" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.ErrBiz(422, 0, "参数错误"))
		return
	}
	db := database.DB
	var last models.EmailCode
	if err := db.Where("email = ? AND purpose = ?", p.Email, p.Purpose).Order("id desc").First(&last).Error; err == nil {
		if time.Since(last.CreatedAt) < 60*time.Second {
			utils.Fail(c, utils.RateLimited("发送过于频繁，请稍后再试"))
			return
		}
	}
	code := utils.GenEmailCode()
	db.Create(&models.EmailCode{Email: p.Email, Code: code, Purpose: p.Purpose, ExpiresAt: time.Now().Add(emailCodeTTL)})
	services.SendVerificationCode(p.Email, code, p.Purpose)
	utils.OKMsg(c, "验证码已发送")
}

// loginUser 登录逻辑（网页端与客户端共用）。
func loginUser(c *gin.Context, account, password string) (map[string]any, *utils.BizError) {
	db := database.DB
	var user models.User
	err := db.Preload("Plan").Where("username = ?", account).First(&user).Error
	if err != nil {
		err = db.Preload("Plan").Where("email = ?", account).First(&user).Error
	}
	if err != nil || !utils.VerifyPassword(user.PasswordHash, password) {
		return nil, utils.ErrBiz(401, 0, "用户名或密码错误")
	}
	if !user.EmailVerified {
		return nil, utils.ErrBiz(403, 1005, "邮箱未验证，请先激活账号")
	}
	if user.Status == "banned" {
		return nil, utils.ErrBiz(403, 1006, "账号已被封禁，请联系管理员")
	}
	now := time.Now()
	user.LastLoginAt = &now
	db.Save(&user)
	token := utils.GenToken()
	expires := now.AddDate(0, 0, config.C.SessionDays)
	db.Create(&models.Session{Token: token, UserID: user.ID, ExpiresAt: expires})
	cache.C.Set("session:"+token, user.ID, time.Until(expires))
	return map[string]any{
		"token": token,
		"user": map[string]any{
			"id":              user.ID,
			"username":        user.Username,
			"email":           user.Email,
			"email_verified":  user.EmailVerified,
			"status":          user.Status,
			"points":          user.Points,
			"is_admin":        utils.IsAdmin(user.Username),
			"plan_name":       user.Plan.Name,
			"plan_expires_at": utils.TimeFmt(user.PlanExpiresAt),
		},
	}, nil
}

// Login 登录。
func Login(c *gin.Context) {
	var p struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.ErrBiz(422, 0, "参数错误"))
		return
	}
	data, be := loginUser(c, p.Username, p.Password)
	if be != nil {
		utils.Fail(c, be)
		return
	}
	utils.OK(c, data, "登录成功")
}

// Logout 登出。
func Logout(c *gin.Context) {
	token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if token != "" {
		database.DB.Where("token = ?", token).Delete(&models.Session{})
		cache.C.Del("session:" + token)
	}
	utils.NoContent(c)
}

// ForgotPassword 找回密码第一步。
func ForgotPassword(c *gin.Context) {
	var p struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.ErrBiz(422, 0, "参数错误"))
		return
	}
	db := database.DB
	var user models.User
	if err := db.Where("email = ?", p.Email).First(&user).Error; err == nil {
		code := utils.GenEmailCode()
		db.Create(&models.EmailCode{Email: p.Email, Code: code, Purpose: "reset_password", ExpiresAt: time.Now().Add(emailCodeTTL)})
		services.SendVerificationCode(p.Email, code, "reset_password")
	}
	utils.OKMsg(c, "如果该邮箱已注册，验证码已发送")
}

// ResetPassword 找回密码第二步。
func ResetPassword(c *gin.Context) {
	var p struct {
		Email       string `json:"email" binding:"required"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.Fail(c, utils.ErrBiz(422, 0, "参数错误"))
		return
	}
	db := database.DB
	var codeRow models.EmailCode
	err := db.Where("email = ? AND code = ? AND purpose = ? AND used = ?",
		p.Email, p.Code, "reset_password", false).Order("id desc").First(&codeRow).Error
	if err != nil {
		utils.Fail(c, utils.ErrBiz(400, 1001, "验证码错误"))
		return
	}
	if time.Now().After(codeRow.ExpiresAt) {
		utils.Fail(c, utils.ErrBiz(400, 1002, "验证码已过期"))
		return
	}
	codeRow.Used = true
	db.Save(&codeRow)
	hash, _ := utils.HashPassword(p.NewPassword)
	db.Model(&models.User{}).Where("email = ?", p.Email).Update("password_hash", hash)
	// 作废该用户全部旧会话
	var user models.User
	if err := db.Where("email = ?", p.Email).First(&user).Error; err == nil {
		var sessions []models.Session
		db.Where("user_id = ?", user.ID).Find(&sessions)
		for _, s := range sessions {
			cache.C.Del("session:" + s.Token)
		}
		db.Where("user_id = ?", user.ID).Delete(&models.Session{})
	}
	utils.OKMsg(c, "密码重置成功，请使用新密码登录")
}

// RequireAuth 别名：供路由使用中间件。
var RequireAuth = middleware.CurrentUser
