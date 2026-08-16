package services

import (
	"log"
	"strconv"
	"strings"
	"time"

	"weavenet/panel/cache"
	"weavenet/panel/config"
	"weavenet/panel/database"
	"weavenet/panel/models"
)

// StartScheduler 启动全部定时任务。
func StartScheduler() {
	go runEvery(time.Hour, checkPlanExpiry, "会员到期检查")
	go runEvery(5*time.Minute, aggregateDailyTraffic, "流量聚合")
	go runEvery(time.Minute, markOfflineNodes, "节点离线检测")
	log.Println("[scheduler] 定时任务已启动")
}

func runEvery(interval time.Duration, fn func(), name string) {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[scheduler] %s 异常: %v", name, r)
				}
			}()
			fn()
		}()
		time.Sleep(interval)
	}
}

// checkPlanExpiry 会员到期自动降级为免费版。
func checkPlanExpiry() {
	now := time.Now()
	var users []models.User
	database.DB.Where("plan_id != 1 AND plan_expires_at IS NOT NULL AND plan_expires_at < ?", now).Find(&users)
	for _, u := range users {
		u.PlanID = 1
		database.DB.Save(&u)
		database.DB.Create(&models.UserPlanLog{UserID: u.ID, PlanID: 1, PlanName: "免费版", Reason: "会员到期自动降级"})
	}
}

// aggregateDailyTraffic 将内存当日流量入库 traffic_stats。
func aggregateDailyTraffic() {
	db := database.DB
	today := time.Now().Format("2006-01-02")
	prefix := "traffic:" + today + ":"
	for _, key := range cache.C.KeysPrefix(prefix) {
		idStr := strings.TrimPrefix(key, prefix)
		tid, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			cache.C.Del(key)
			continue
		}
		v, ok := cache.C.GetAndDel(key)
		if !ok {
			continue
		}
		t, ok := v.(cache.TrafficEntry)
		if !ok {
			continue
		}
		var stat models.TrafficStat
		err = db.Where("tunnel_id = ? AND date = ?", tid, today).First(&stat).Error
		if err != nil {
			stat = models.TrafficStat{TunnelID: uint(tid), Date: today, InBytes: t.In, OutBytes: t.Out}
			db.Create(&stat)
		} else {
			stat.InBytes += t.In
			stat.OutBytes += t.Out
			db.Save(&stat)
		}
	}
}

// markOfflineNodes 节点心跳超时置 offline。
func markOfflineNodes() {
	timeout := time.Duration(config.C.NodeHeartbeatTO) * time.Second
	var nodes []models.Node
	database.DB.Where("status IN ?", []string{"online", "maintenance"}).Find(&nodes)
	for _, n := range nodes {
		_, ok := cache.C.Get(cache.NodeKey(n.ID))
		if !ok {
			// Redis 降级兜底：用 last_heartbeat_at
			if n.LastHeartbeatAt == nil || time.Since(*n.LastHeartbeatAt) > timeout {
				database.DB.Model(&models.Node{}).Where("id = ?", n.ID).Update("status", "offline")
			}
		}
	}
}
