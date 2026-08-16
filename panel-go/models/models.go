package models

import (
	"time"

	"gorm.io/gorm"
)

// 时间格式：API 输出 isoformat（YYYY-MM-DDTHH:MM:SS），日期 YYYY-MM-DD。

// User 用户。
type User struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Username      string     `gorm:"size:64;uniqueIndex" json:"username"`
	Email         string     `gorm:"size:255;uniqueIndex" json:"email"`
	PasswordHash  string     `gorm:"size:255" json:"-"`
	EmailVerified bool       `gorm:"default:false" json:"email_verified"`
	Status        string     `gorm:"size:16;default:active" json:"status"`
	PlanID        uint       `gorm:"index;default:1" json:"plan_id"`
	PlanExpiresAt *time.Time `json:"plan_expires_at"`
	Points        int        `gorm:"default:0" json:"points"`
	CreatedAt     time.Time  `gorm:"index" json:"created_at"`
	LastLoginAt   *time.Time `json:"last_login_at"`

	Plan    Plan      `gorm:"foreignKey:PlanID" json:"plan"`
	Tunnels []Tunnel  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// Plan 套餐。
type Plan struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	Name           string `gorm:"size:64;uniqueIndex" json:"name"`
	SpeedLimitMbps int    `gorm:"default:8" json:"speed_limit_mbps"`
	TunnelLimit    int    `gorm:"default:3" json:"tunnel_limit"`
	DomainLimit    int    `gorm:"default:1" json:"domain_limit"`
	Sort           int    `gorm:"default:0" json:"sort"`
	IsDefault      bool   `gorm:"default:false" json:"is_default"`
}

// Node 节点。
type Node struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"size:64;uniqueIndex" json:"name"`
	Address         string     `gorm:"size:255" json:"address"`
	Port            int        `gorm:"default:7000" json:"port"`
	Status          string     `gorm:"size:16;default:offline" json:"status"`
	SpeedLimitMbps  int        `gorm:"default:100" json:"speed_limit_mbps"`
	AgentToken      string     `gorm:"size:128;default:''" json:"-"`
	Remark          string     `gorm:"size:512;default:''" json:"remark"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// Tunnel 隧道。
type Tunnel struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	UserID       uint   `gorm:"index;constraint:OnDelete:CASCADE" json:"user_id"`
	NodeID       uint   `gorm:"index;constraint:OnDelete:RESTRICT" json:"node_id"`
	Name         string `gorm:"size:64" json:"name"`
	Type         string `gorm:"size:20" json:"type"`
	LocalIP      string `gorm:"size:255;default:127.0.0.1" json:"local_ip"`
	LocalPort    int    `gorm:"default:0" json:"local_port"`
	RemotePort   *int   `gorm:"index" json:"remote_port"`
	Subdomain    *string `gorm:"size:128;index" json:"subdomain"`
	CustomDomain *string `gorm:"size:255" json:"custom_domain"`
	KCP          bool   `gorm:"default:false" json:"kcp"`
	Encryption   bool   `gorm:"default:true" json:"encryption"`
	Compression  bool   `gorm:"default:false" json:"compression"`
	SecretKey    *string `gorm:"size:128" json:"secret_key"`
	LoadBalancers string `gorm:"type:text" json:"load_balancers"`
	Status       string `gorm:"size:16;default:stopped" json:"status"`
	StatusDetail string `gorm:"size:512;default:''" json:"status_detail"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
	Node Node `gorm:"foreignKey:NodeID" json:"node"`
}

// Domain 免费域名。
type Domain struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;constraint:OnDelete:CASCADE" json:"user_id"`
	TunnelID   *uint     `gorm:"constraint:OnDelete:SET NULL" json:"tunnel_id"`
	Subdomain  string    `gorm:"size:128;uniqueIndex" json:"subdomain"`
	FullDomain string    `gorm:"size:255;index" json:"full_domain"`
	Status     string    `gorm:"size:16;default:active" json:"status"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// SigninLog 签到记录。
type SigninLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"index;constraint:OnDelete:CASCADE" json:"user_id"`
	SigninDate     string    `gorm:"size:16;index" json:"signin_date"`
	Points         int       `gorm:"default:0" json:"points"`
	ContinuousDays int       `gorm:"default:1" json:"continuous_days"`
	CreatedAt      time.Time `json:"created_at"`
}

// PointsLog 积分流水。
type PointsLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;constraint:OnDelete:CASCADE" json:"user_id"`
	Change    int       `json:"change"`
	Balance   int       `gorm:"default:0" json:"balance"`
	Reason    string    `gorm:"size:128" json:"reason"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// Ticket 工单。
type Ticket struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;constraint:OnDelete:CASCADE" json:"user_id"`
	Title      string    `gorm:"size:128" json:"title"`
	Content    string    `gorm:"type:text" json:"content"`
	Status     string    `gorm:"size:16;default:open" json:"status"`
	AdminReply *string   `gorm:"type:text" json:"admin_reply"`
	AdminID    *uint     `json:"admin_id"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Announcement 公告。
type Announcement struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:128" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	Author    string    `gorm:"size:64;default:''" json:"author"`
	Status    string    `gorm:"size:16;default:active" json:"status"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TrafficStat 按日流量聚合。
type TrafficStat struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	TunnelID uint   `gorm:"index;constraint:OnDelete:CASCADE" json:"tunnel_id"`
	Date     string `gorm:"size:16;index" json:"date"`
	InBytes  int64  `gorm:"default:0" json:"in_bytes"`
	OutBytes int64  `gorm:"default:0" json:"out_bytes"`
}

// SystemConfig 系统配置 key-value。
type SystemConfig struct {
	Key   string `gorm:"primaryKey;size:64" json:"key"`
	Value string `gorm:"type:text;default:''" json:"value"`
}

// Session 会话。
type Session struct {
	Token     string    `gorm:"primaryKey;size:128" json:"-"`
	UserID    uint      `gorm:"index;constraint:OnDelete:CASCADE" json:"user_id"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// EmailCode 邮箱验证码。
type EmailCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"size:255;index" json:"email"`
	Code      string    `gorm:"size:16" json:"code"`
	Purpose   string    `gorm:"size:32" json:"purpose"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	Attempts  int       `gorm:"default:0" json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
}

// OperationLog 操作审计。
type OperationLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	AdminID    *uint     `gorm:"index" json:"admin_id"`
	AdminName  string    `gorm:"size:64;default:''" json:"admin_name"`
	Action     string    `gorm:"size:64" json:"action"`
	TargetType string    `gorm:"size:32" json:"target_type"`
	TargetID   *uint     `json:"target_id"`
	Detail     string    `gorm:"type:text;default:''" json:"detail"`
	IP         string    `gorm:"size:64;default:''" json:"ip"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// UserPlanLog 套餐变更记录。
type UserPlanLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;constraint:OnDelete:CASCADE" json:"user_id"`
	PlanID    uint      `gorm:"index" json:"plan_id"`
	PlanName  string    `gorm:"size:64;default:''" json:"plan_name"`
	Reason    string    `gorm:"size:128;default:''" json:"reason"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// BeforeCreate 钩子：无特殊逻辑，占位。
func (u *User) BeforeCreate(tx *gorm.DB) error { return nil }
