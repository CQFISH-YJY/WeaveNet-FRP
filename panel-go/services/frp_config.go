package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"weavenet/panel/config"
	"weavenet/panel/database"
	"weavenet/panel/models"
	"weavenet/panel/utils"
)

// RESERVED_PORTS 保留端口。
var RESERVED_PORTS = func() map[int]bool {
	m := make(map[int]bool)
	for i := 1; i < 1024; i++ {
		m[i] = true
	}
	for _, p := range []int{7000, 8000, 9001, 5432, 3306, 6379, 27017} {
		m[p] = true
	}
	return m
}()

// CheckTunnelLimit 检查隧道数量额度。
func CheckTunnelLimit(db *gorm.DB, user *models.User, plan *models.Plan) error {
	var count int64
	db.Model(&models.Tunnel{}).Where("user_id = ?", user.ID).Count(&count)
	if count >= int64(plan.TunnelLimit) {
		return utils.ErrBiz(400, 2001, fmt.Sprintf("套餐最多可创建 %d 条隧道，请升级套餐", plan.TunnelLimit))
	}
	return nil
}

// CheckDomainLimit 检查域名额度。
func CheckDomainLimit(db *gorm.DB, user *models.User, plan *models.Plan) error {
	var count int64
	db.Model(&models.Domain{}).Where("user_id = ? AND status = ?", user.ID, "active").Count(&count)
	if count >= int64(plan.DomainLimit) {
		return utils.ErrBiz(400, 3003, "免费域名额度不足，请升级套餐")
	}
	return nil
}

// AllocateRemotePort 分配远程端口（四段扫描）。
func AllocateRemotePort(db *gorm.DB, nodeID uint) (int, error) {
	var occupied []int
	db.Model(&models.Tunnel{}).Where("node_id = ? AND remote_port IS NOT NULL", nodeID).Pluck("remote_port", &occupied)
	oc := make(map[int]bool)
	for _, p := range occupied {
		oc[p] = true
	}
	starts := []int{20000, 40000, 1024, 10000}
	for _, start := range starts {
		for i := 0; i < 1000; i++ {
			p := start + i
			if p > 65535 {
				break
			}
			if RESERVED_PORTS[p] || oc[p] {
				continue
			}
			return p, nil
		}
	}
	return 0, utils.ErrBiz(409, 3001, "节点远程端口已用尽")
}

// CheckRemotePortAvailable 检查远程端口是否可用。
func CheckRemotePortAvailable(db *gorm.DB, nodeID uint, port int, excludeID uint) bool {
	if RESERVED_PORTS[port] {
		return false
	}
	q := database.DB.Model(&models.Tunnel{}).Where("node_id = ? AND remote_port = ?", nodeID, port)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	q.Count(&count)
	return count == 0
}

// CheckSubdomainAvailable 检查子域名是否可用。
func CheckSubdomainAvailable(db *gorm.DB, subdomain string, excludeID uint) bool {
	q := db.Model(&models.Domain{}).Where("subdomain = ? AND status = ?", subdomain, "active")
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	var count int64
	q.Count(&count)
	return count == 0
}

// BuildPublicAddress 构建公网访问地址。
func BuildPublicAddress(t *models.Tunnel, node *models.Node) string {
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
		return fmt.Sprintf("%s://%s:%d", t.Type, node.Address, t.RemotePort)
	}
	if t.RemotePort != nil {
		return fmt.Sprintf("%s:%d", node.Address, *t.RemotePort)
	}
	if node.Address != "" {
		return node.Address
	}
	return config.C.PanelBaseURL
}

// TunnelBandwidth 带宽字符串（KB）。
func TunnelBandwidth(mbps int) string {
	return fmt.Sprintf("%dKB", mbps*1000)
}

// GenerateFrpcConfig 生成 frpc.toml 配置文本。
func GenerateFrpcConfig(db *gorm.DB, t *models.Tunnel, user *models.User, node *models.Node) string {
	var sb strings.Builder
	token := utils.UserTokenForFrpc(user.ID, user.PasswordHash)
	sb.WriteString("serverAddr = \"" + node.Address + "\"\n")
	sb.WriteString(fmt.Sprintf("serverPort = %d\n", node.Port))
	sb.WriteString("auth.method = \"token\"\n")
	sb.WriteString(fmt.Sprintf("auth.token = \"%s\"\n", token))
	sb.WriteString("[[proxies]]\n")
	sb.WriteString(fmt.Sprintf("name = \"%s-%d\"\n", t.Name, t.ID))
	sb.WriteString(fmt.Sprintf("type = \"%s\"\n", t.Type))
	sb.WriteString(fmt.Sprintf("localIP = \"%s\"\n", t.LocalIP))
	sb.WriteString(fmt.Sprintf("localPort = %d\n", t.LocalPort))
	if t.RemotePort != nil {
		sb.WriteString(fmt.Sprintf("remotePort = %d\n", *t.RemotePort))
	}
	if t.Subdomain != nil && *t.Subdomain != "" {
		sb.WriteString(fmt.Sprintf("subdomain = \"%s\"\n", *t.Subdomain))
	}
	if t.CustomDomain != nil && *t.CustomDomain != "" {
		sb.WriteString(fmt.Sprintf("customDomains = [\"%s\"]\n", *t.CustomDomain))
	}
	sb.WriteString(fmt.Sprintf("transport.kcp = %v\n", t.KCP))
	if user.Plan != (models.Plan{}) {
		sb.WriteString(fmt.Sprintf("transport.bandwidthLimit = \"%s\"\n", TunnelBandwidth(user.Plan.SpeedLimitMbps)))
	}
	sb.WriteString("transport.bandwidthLimitMode = \"client\"\n")
	sb.WriteString(fmt.Sprintf("transport.encryption.enable = %v\n", t.Encryption))
	if t.Compression {
		sb.WriteString("transport.compression.enable = true\n")
	}
	if t.SecretKey != nil && *t.SecretKey != "" {
		sb.WriteString(fmt.Sprintf("secretKey = \"%s\"\n", *t.SecretKey))
	}
	if t.Type == "loadbalance" {
		var lbs []struct {
			IP   string `json:"ip"`
			Port int    `json:"port"`
		}
		if json.Unmarshal([]byte(t.LoadBalancers), &lbs) == nil {
			for i, lb := range lbs {
				sb.WriteString("[[proxies]]\n")
				sb.WriteString(fmt.Sprintf("name = \"%s-lb%d-%d\"\n", t.Name, i+1, t.ID))
				sb.WriteString("type = \"tcp\"\n")
				sb.WriteString(fmt.Sprintf("localIP = \"%s\"\n", lb.IP))
				sb.WriteString(fmt.Sprintf("localPort = %d\n", lb.Port))
				if t.RemotePort != nil {
					sb.WriteString(fmt.Sprintf("remotePort = %d\n", *t.RemotePort))
				}
				sb.WriteString("loadBalancer.group = \"weavenet-lb\"\n")
				sb.WriteString(fmt.Sprintf("loadBalancer.groupKey = \"%d\"\n", t.ID))
			}
		}
	}
	return sb.String()
}

// ParseLB 解析负载均衡后端列表。
func ParseLB(s string) []map[string]any {
	var out []map[string]any
	if s == "" {
		return out
	}
	var raw []struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	}
	if json.Unmarshal([]byte(s), &raw) != nil {
		return out
	}
	for _, r := range raw {
		out = append(out, map[string]any{"ip": r.IP, "port": r.Port})
	}
	return out
}
