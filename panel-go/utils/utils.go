package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"weavenet/panel/config"
)

// BizError 业务错误。
type BizError struct {
	HTTPCode     int
	BusinessCode int
	Message      string
}

func (e *BizError) Error() string { return e.Message }

// 常用错误构造。
func AuthErr(msg string) *BizError {
	if msg == "" {
		msg = "未登录或登录已过期"
	}
	return &BizError{http.StatusUnauthorized, 0, msg}
}
func Forbidden(msg string) *BizError {
	if msg == "" {
		msg = "没有权限执行此操作"
	}
	return &BizError{http.StatusForbidden, 0, msg}
}
func NotFound(msg string) *BizError {
	if msg == "" {
		msg = "资源不存在"
	}
	return &BizError{http.StatusNotFound, 0, msg}
}
func Conflict(bc int, msg string) *BizError { return &BizError{http.StatusConflict, bc, msg} }
func RateLimited(msg string) *BizError {
	if msg == "" {
		msg = "请求过于频繁，请稍后再试"
	}
	return &BizError{http.StatusTooManyRequests, 0, msg}
}
func Biz(httpCode, bc int, msg string) *BizError { return &BizError{httpCode, bc, msg} }

// OK 成功响应：{code:0, message, data}。
func OK(c *gin.Context, data any, msg string) {
	if msg == "" {
		msg = "ok"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg, "data": data})
}

// OKMsg 无数据成功。
func OKMsg(c *gin.Context, msg string) { OK(c, nil, msg) }

// Fail 错误响应：{code:http, business_code, message, data:null}。
func Fail(c *gin.Context, e *BizError) {
	c.AbortWithStatusJSON(e.HTTPCode, gin.H{
		"code":          e.HTTPCode,
		"business_code": e.BusinessCode,
		"message":       e.Message,
		"data":          nil,
	})
}

// NoContent 204 无响应体。
func NoContent(c *gin.Context) { c.Status(http.StatusNoContent) }

// TimeFmt 时间转 isoformat（YYYY-MM-DDTHH:MM:SS）。
func TimeFmt(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02T15:04:05")
}

// TimeFmtV 值类型时间转 isoformat。
func TimeFmtV(t time.Time) string { return t.Format("2006-01-02T15:04:05") }

// Today 今日日期 YYYY-MM-DD。
func Today() string { return time.Now().Format("2006-01-02") }

// Yesterday 昨日日期。
func Yesterday() string { return time.Now().AddDate(0, 0, -1).Format("2006-01-02") }

// HashPassword bcrypt 哈希。
func HashPassword(pwd string) (string, error) {
	rounds := config.C.BcryptRounds
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), rounds)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// VerifyPassword 校验密码。
func VerifyPassword(hash, pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)) == nil
}

// GenToken 生成 64 位随机 token（字母+数字）。
func GenToken() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return randStr(chars, 64)
}

// GenAgentToken agent_ 前缀 48 位随机。
func GenAgentToken() string { return "agent_" + randStr("abcdefghijklmnopqrstuvwxyz0123456789", 48) }

// GenEmailCode 6 位数字验证码。
func GenEmailCode() string { return randStr("0123456789", 6) }

// GenSecretKey 12 位（stcp/xtcp 访问密钥）。
func GenSecretKey() string {
	return randStr("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 12)
}

func randStr(chars string, n int) string {
	var sb strings.Builder
	cb := big.NewInt(int64(len(chars)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, cb)
		if err != nil {
			sb.WriteByte(chars[0])
			continue
		}
		sb.WriteByte(chars[idx.Int64()])
	}
	return sb.String()
}

// UserTokenForFrpc 用户 frpc 鉴权 Token：u_ + sha256(id:password_hash:secret_key)。
func UserTokenForFrpc(userID uint, passwordHash string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s", userID, passwordHash, config.C.SecretKey)))
	return "u_" + hex.EncodeToString(h[:])
}

// MaskToken 打码 token 用于日志。
func MaskToken(t string) string {
	if len(t) <= 8 {
		return "***"
	}
	return t[:4] + "***" + t[len(t)-4:]
}

// ClientIP 获取客户端 IP。
func ClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	return c.ClientIP()
}

// IsAdmin 判断是否为管理员。
func IsAdmin(username string) bool { return username == config.C.AdminUsername }

// ErrBiz 构造 BizError 快捷方式。
func ErrBiz(code int, bc int, msg string) *BizError {
	return &BizError{code, bc, msg}
}

// AsBiz 将 error 转为 BizError。
func AsBiz(err error) *BizError {
	var be *BizError
	if errors.As(err, &be) {
		return be
	}
	return ErrBiz(http.StatusInternalServerError, 9001, "服务器内部错误，请稍后重试")
}

// RegisterBiz 全局错误处理中间件使用。
type CtxKey string

const UserKey CtxKey = "user"

// AtoiSafe 安全字符串转整数。
func AtoiSafe(s string, def int) (int, bool) {
	if s == "" {
		return def, false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def, false
	}
	return n, true
}
