package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"weavenet/panel/cache"
	"weavenet/panel/config"
	"weavenet/panel/database"
	"weavenet/panel/models"
	"weavenet/panel/utils"
)

// extractToken 从 Authorization Bearer 提取 token。
func extractToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

// CurrentUser 认证中间件：注入 *models.User。
func CurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			utils.Fail(c, utils.AuthErr(""))
			return
		}
		user, err := loadUserByToken(token)
		if err != nil {
			utils.Fail(c, utils.AsBiz(err))
			return
		}
		c.Set(string(utils.UserKey), user)
		c.Next()
	}
}

// loadUserByToken 按会话 token 加载用户（内存缓存优先，DB 兜底）。
func loadUserByToken(token string) (*models.User, error) {
	cacheKey := "session:" + token
	if v, ok := cache.C.Get(cacheKey); ok {
		if uid, ok2 := v.(uint); ok2 {
			return loadUserByID(uid)
		}
	}
	var sess models.Session
	if err := database.DB.Where("token = ?", token).First(&sess).Error; err != nil {
		return nil, utils.AuthErr("")
	}
	if time.Now().After(sess.ExpiresAt) {
		return nil, utils.AuthErr("")
	}
	ttl := time.Until(sess.ExpiresAt)
	cache.C.Set(cacheKey, sess.UserID, ttl)
	return loadUserByID(sess.UserID)
}

// loadUserByID 加载用户并检查封禁。
func loadUserByID(uid uint) (*models.User, error) {
	var user models.User
	if err := database.DB.Preload("Plan").First(&user, uid).Error; err != nil {
		return nil, utils.AuthErr("")
	}
	if user.Status == "banned" {
		return nil, utils.ErrBiz(403, 1006, "账号已被封禁")
	}
	return &user, nil
}

// CurrentUserOpt 用户对象辅助。
func GetUser(c *gin.Context) *models.User {
	v, _ := c.Get(string(utils.UserKey))
	u, _ := v.(*models.User)
	return u
}

// CurrentAdmin 管理员中间件（用户名等于 ADMIN_USERNAME）。
func CurrentAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUser(c)
		if user == nil {
			utils.Fail(c, utils.AuthErr(""))
			return
		}
		if user.Username != config.C.AdminUsername {
			utils.Fail(c, utils.Forbidden(""))
			return
		}
		c.Next()
	}
}

// AgentNode 中间件：agent_ 前缀 token 匹配节点。
func AgentNode() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if !strings.HasPrefix(token, "agent_") {
			utils.Fail(c, utils.Forbidden("节点鉴权失败"))
			return
		}
		var node models.Node
		if err := database.DB.Where("agent_token = ?", token).First(&node).Error; err != nil {
			utils.Fail(c, utils.Forbidden("节点鉴权失败"))
			return
		}
		c.Set("agentNode", &node)
		c.Next()
	}
}

// GetAgentNode 获取当前 agent 节点。
func GetAgentNode(c *gin.Context) *models.Node {
	v, _ := c.Get("agentNode")
	n, _ := v.(*models.Node)
	return n
}

// RateLimit 内存滑动窗口限流。
func RateLimit(scope string, limit int, window time.Duration) gin.HandlerFunc {
	if !config.C.RateLimitEnabled {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		ip := utils.ClientIP(c)
		key := "ratelimit:" + scope + ":" + ip
		n := cache.C.IncrBy(key, 1, window)
		if n > int64(limit) {
			utils.Fail(c, utils.RateLimited(""))
			return
		}
		c.Next()
	}
}
