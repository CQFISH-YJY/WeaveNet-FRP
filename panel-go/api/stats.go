package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"weavenet/panel/cache"
	"weavenet/panel/database"
	"weavenet/panel/middleware"
	"weavenet/panel/models"
	"weavenet/panel/services"
	"weavenet/panel/utils"
)

// DoSignin 每日签到。
func DoSignin(c *gin.Context) {
	user := middleware.GetUser(c)
	data, err := services.DoSignin(database.DB, user)
	if err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	utils.OK(c, data, "签到成功")
}

// GetSigninStatus 签到状态。
func GetSigninStatus(c *gin.Context) {
	user := middleware.GetUser(c)
	utils.OK(c, services.SigninStatus(database.DB, user), "ok")
}

// ListPointsLogs 积分流水。
func ListPointsLogs(c *gin.Context) {
	user := middleware.GetUser(c)
	db := database.DB
	page, _ := utils.AtoiSafe(c.DefaultQuery("page", "1"), 1)
	pageSize, _ := utils.AtoiSafe(c.DefaultQuery("page_size", "10"), 10)
	var total int64
	db.Model(&models.PointsLog{}).Where("user_id = ?", user.ID).Count(&total)
	var logs []models.PointsLog
	db.Where("user_id = ?", user.ID).Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)
	items := make([]gin.H, 0, len(logs))
	for _, l := range logs {
		items = append(items, gin.H{
			"id": l.ID, "change": l.Change, "balance": l.Balance, "reason": l.Reason,
			"created_at": utils.TimeFmtV(l.CreatedAt),
		})
	}
	utils.OK(c, gin.H{"total": total, "page": page, "page_size": pageSize, "items": items}, "ok")
}

// ExchangePoints 积分兑换会员。
func ExchangePoints(c *gin.Context) {
	user := middleware.GetUser(c)
	data, err := services.ExchangeMembership(database.DB, user)
	if err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	utils.OK(c, data, "兑换成功")
}

// GetPointsRules 积分规则。
func GetPointsRules(c *gin.Context) {
	db := database.DB
	rules := map[string]any{
		"signin_points":       services.GetConfigInt(db, "signin_points", 10),
		"signin_streak_bonus": services.GetConfigInt(db, "signin_streak_bonus", 30),
		"signin_streak_days":  services.GetConfigInt(db, "signin_streak_days", 7),
		"exchange_points":     services.GetConfigInt(db, "exchange_points", 300),
		"exchange_plan_days":  services.GetConfigInt(db, "exchange_plan_days", 30),
		"exchange_plan_name":  services.GetConfig(db, "exchange_plan_name", "普通会员"),
	}
	utils.OK(c, rules, "ok")
}

// GetTrafficStats 近 N 日流量统计。
func GetTrafficStats(c *gin.Context) {
	days, _ := utils.AtoiSafe(c.DefaultQuery("days", "7"), 7)
	if days < 1 {
		days = 1
	}
	if days > 31 {
		days = 31
	}
	user := middleware.GetUser(c)
	db := database.DB
	var tunnelIDs []uint
	db.Model(&models.Tunnel{}).Where("user_id = ?", user.ID).Pluck("id", &tunnelIDs)
	startDate := dateNDaysAgo(days - 1)
	var rows []struct {
		Date     string
		InBytes  int64
		OutBytes int64
	}
	q := db.Model(&models.TrafficStat{}).Select("date, SUM(in_bytes) as in_bytes, SUM(out_bytes) as out_bytes")
	if len(tunnelIDs) > 0 {
		q = q.Where("tunnel_id IN ?", tunnelIDs)
	} else {
		q = q.Where("tunnel_id = ?", 0)
	}
	q.Where("date >= ?", startDate).Group("date").Scan(&rows)
	byDate := make(map[string][2]int64)
	for _, r := range rows {
		byDate[r.Date] = [2]int64{r.InBytes, r.OutBytes}
	}
	items := make([]gin.H, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := dateNDaysAgo(i)
		v, ok := byDate[d]
		if !ok {
			v = [2]int64{0, 0}
		}
		items = append(items, gin.H{"date": d, "in_bytes": v[0], "out_bytes": v[1]})
	}
	utils.OK(c, items, "ok")
}

// GetOverview 在线概览。
func GetOverview(c *gin.Context) {
	user := middleware.GetUser(c)
	db := database.DB
	var tunnelTotal int64
	db.Model(&models.Tunnel{}).Where("user_id = ?", user.ID).Count(&tunnelTotal)
	var tunnelRunning int64
	db.Model(&models.Tunnel{}).Where("user_id = ? AND status = ?", user.ID, "running").Count(&tunnelRunning)
	todayIn, todayOut := trafficSumForUser(user.ID)
	utils.OK(c, gin.H{
		"tunnel_total":   tunnelTotal,
		"tunnel_running": tunnelRunning,
		"today_in":       todayIn,
		"today_out":      todayOut,
		"connections":    0,
		"points":         user.Points,
	}, "ok")
}

// dateNDaysAgo 返回 N 天前日期 YYYY-MM-DD。
func dateNDaysAgo(n int) string {
	return time.Now().AddDate(0, 0, -n).Format("2006-01-02")
}

// trafficSumForTunnel 单个隧道今日流量。
func trafficSumForTunnel(tid uint) (int64, int64) {
	prefix := "traffic:" + utils.Today() + ":"
	key := prefix + strconv.FormatUint(uint64(tid), 10)
	v, ok := cache.C.Get(key)
	if !ok {
		return 0, 0
	}
	t, ok := v.(cache.TrafficEntry)
	if !ok {
		return 0, 0
	}
	return t.In, t.Out
}

// trafficSumForUser 用户全部隧道今日流量。
func trafficSumForUser(userID uint) (int64, int64) {
	var ids []uint
	database.DB.Model(&models.Tunnel{}).Where("user_id = ?", userID).Pluck("id", &ids)
	var in, out int64
	for _, id := range ids {
		i, o := trafficSumForTunnel(id)
		in += i
		out += o
	}
	return in, out
}

// trafficKeyForTunnel 构造流量缓存 key（兼容 cache 包）。
func trafficKeyForTunnel(tid uint) string {
	var sb strings.Builder
	sb.WriteString("traffic:")
	sb.WriteString(utils.Today())
	sb.WriteString(":")
	sb.WriteString(strconv.FormatUint(uint64(tid), 10))
	return sb.String()
}
