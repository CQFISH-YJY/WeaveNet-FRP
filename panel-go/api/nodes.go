package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"weavenet/panel/cache"
	"weavenet/panel/database"
	"weavenet/panel/models"
	"weavenet/panel/utils"
)

// ListNodes 节点列表（不返回 agent_token）。
func ListNodes(c *gin.Context) {
	db := database.DB
	var nodes []models.Node
	db.Order("id").Find(&nodes)
	items := make([]gin.H, 0, len(nodes))
	for _, n := range nodes {
		_, online := cache.C.Get(cache.NodeKey(n.ID))
		if !online && n.Status == "online" {
			n.Status = "offline"
		}
		items = append(items, gin.H{
			"id": n.ID, "name": n.Name, "address": n.Address, "port": n.Port,
			"status": n.Status, "speed_limit_mbps": n.SpeedLimitMbps, "remark": n.Remark,
			"last_heartbeat_at": utils.TimeFmt(n.LastHeartbeatAt),
		})
	}
	utils.OK(c, items, "ok")
}

// GetNodeStatus 单节点状态。
func GetNodeStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("node_id"))
	db := database.DB
	var n models.Node
	if err := db.First(&n, id).Error; err != nil {
		utils.Fail(c, utils.NotFound("节点不存在"))
		return
	}
	utils.OK(c, gin.H{
		"id": n.ID, "name": n.Name, "status": n.Status,
		"speed_limit_mbps": n.SpeedLimitMbps,
		"last_heartbeat_at": utils.TimeFmt(n.LastHeartbeatAt),
	}, "ok")
}

// ListAnnouncements 公告列表。
func ListAnnouncements(c *gin.Context) {
	db := database.DB
	page, _ := utils.AtoiSafe(c.DefaultQuery("page", "1"), 1)
	pageSize, _ := utils.AtoiSafe(c.DefaultQuery("page_size", "10"), 10)
	var total int64
	db.Model(&models.Announcement{}).Where("status = ?", "active").Count(&total)
	var items []models.Announcement
	db.Where("status = ?", "active").Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items)
	list := make([]gin.H, 0, len(items))
	for _, a := range items {
		content := a.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		list = append(list, gin.H{
			"id": a.ID, "title": a.Title, "content": content, "author": a.Author,
			"created_at": utils.TimeFmtV(a.CreatedAt),
		})
	}
	utils.OK(c, gin.H{"total": total, "page": page, "page_size": pageSize, "items": list}, "ok")
}

// GetAnnouncement 公告详情。
func GetAnnouncement(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("announcement_id"))
	db := database.DB
	var a models.Announcement
	if err := db.First(&a, id).Error; err != nil {
		utils.Fail(c, utils.NotFound("公告不存在"))
		return
	}
	if a.Status != "active" {
		utils.Fail(c, utils.ErrBiz(400, 5002, "公告已下线"))
		return
	}
	utils.OK(c, gin.H{
		"id": a.ID, "title": a.Title, "content": a.Content, "author": a.Author,
		"created_at": utils.TimeFmtV(a.CreatedAt), "updated_at": utils.TimeFmtV(a.UpdatedAt),
	}, "ok")
}
