package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"weavenet/panel/database"
	"weavenet/panel/middleware"
	"weavenet/panel/models"
	"weavenet/panel/utils"
)

// CreateTicket 创建工单。
func CreateTicket(c *gin.Context) {
	var p struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&p); err != nil || p.Title == "" || p.Content == "" {
		utils.Fail(c, utils.Biz(400, 0, "标题和内容不能为空"))
		return
	}
	user := middleware.GetUser(c)
	t := models.Ticket{UserID: user.ID, Title: p.Title, Content: p.Content, Status: "open"}
	if err := database.DB.Create(&t).Error; err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	utils.OK(c, serializeTicket(&t), "工单已提交")
}

// ListTickets 工单列表。
func ListTickets(c *gin.Context) {
	user := middleware.GetUser(c)
	db := database.DB
	page, _ := utils.AtoiSafe(c.DefaultQuery("page", "1"), 1)
	pageSize, _ := utils.AtoiSafe(c.DefaultQuery("page_size", "20"), 20)
	var total int64
	db.Model(&models.Ticket{}).Where("user_id = ?", user.ID).Count(&total)
	var list []models.Ticket
	db.Where("user_id = ?", user.ID).Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	items := make([]gin.H, 0, len(list))
	for _, t := range list {
		items = append(items, serializeTicket(&t))
	}
	utils.OK(c, gin.H{"total": total, "page": page, "page_size": pageSize, "items": items}, "ok")
}

// TicketDetail 工单详情。
func TicketDetail(c *gin.Context) {
	user := middleware.GetUser(c)
	t, be := getOwnedTicket(user.ID, c.Param("ticket_id"))
	if be != nil {
		utils.Fail(c, be)
		return
	}
	utils.OK(c, serializeTicket(t), "ok")
}

// ReplyTicket 回复工单（用户补充信息）。
func ReplyTicket(c *gin.Context) {
	var p struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&p); err != nil || p.Content == "" {
		utils.Fail(c, utils.Biz(400, 0, "回复内容不能为空"))
		return
	}
	user := middleware.GetUser(c)
	t, be := getOwnedTicket(user.ID, c.Param("ticket_id"))
	if be != nil {
		utils.Fail(c, be)
		return
	}
	if t.Status != "open" {
		utils.Fail(c, utils.Biz(400, 5001, "工单已关闭，无法回复"))
		return
	}
	t.Content += "\n\n【用户回复】\n" + p.Content
	if err := database.DB.Save(t).Error; err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	utils.OK(c, serializeTicket(t), "回复成功")
}

// CloseTicket 关闭工单。
func CloseTicket(c *gin.Context) {
	user := middleware.GetUser(c)
	t, be := getOwnedTicket(user.ID, c.Param("ticket_id"))
	if be != nil {
		utils.Fail(c, be)
		return
	}
	if t.Status != "open" {
		utils.Fail(c, utils.Biz(400, 5001, "工单已关闭"))
		return
	}
	t.Status = "closed"
	if err := database.DB.Save(t).Error; err != nil {
		utils.Fail(c, utils.AsBiz(err))
		return
	}
	utils.OK(c, serializeTicket(t), "工单已关闭")
}

func serializeTicket(t *models.Ticket) gin.H {
	return gin.H{
		"id": t.ID, "user_id": t.UserID, "title": t.Title, "content": t.Content,
		"status": t.Status, "admin_reply": t.AdminReply,
		"created_at": utils.TimeFmtV(t.CreatedAt), "updated_at": utils.TimeFmtV(t.UpdatedAt),
	}
}

func getOwnedTicket(userID uint, idStr string) (*models.Ticket, *utils.BizError) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, utils.NotFound("工单不存在")
	}
	var t models.Ticket
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&t).Error; err != nil {
		return nil, utils.NotFound("工单不存在")
	}
	return &t, nil
}
