package services

import (
	"strconv"
	"time"

	"gorm.io/gorm"

	"weavenet/panel/config"
	"weavenet/panel/models"
	"weavenet/panel/utils"
)

// AddPoints 增加/扣除积分并写流水。
func AddPoints(db *gorm.DB, user *models.User, change int, reason string) error {
	if user.Points+change < 0 {
		return utils.ErrBiz(400, 4002, "积分不足")
	}
	user.Points += change
	if err := db.Save(user).Error; err != nil {
		return err
	}
	pl := models.PointsLog{UserID: user.ID, Change: change, Balance: user.Points, Reason: reason}
	return db.Create(&pl).Error
}

// SigninStatus 签到状态。
func SigninStatus(db *gorm.DB, user *models.User) map[string]any {
	today := utils.Today()
	var count int64
	db.Model(&models.SigninLog{}).Where("user_id = ? AND signin_date = ?", user.ID, today).Count(&count)
	continuous := 0
	var last models.SigninLog
	if err := db.Where("user_id = ?", user.ID).Order("signin_date desc").First(&last).Error; err == nil {
		continuous = last.ContinuousDays
	}
	nextBonus := config.C.SigninStreakDays - continuous
	if nextBonus < 0 {
		nextBonus = 0
	}
	return map[string]any{
		"today_signed":    count > 0,
		"continuous_days": continuous,
		"next_bonus_in":   nextBonus,
	}
}

// DoSignin 每日签到。
func DoSignin(db *gorm.DB, user *models.User) (map[string]any, error) {
	today := utils.Today()
	var count int64
	db.Model(&models.SigninLog{}).Where("user_id = ? AND signin_date = ?", user.ID, today).Count(&count)
	if count > 0 {
		return nil, utils.ErrBiz(400, 4001, "今天已经签到过了")
	}
	continuous := 1
	var last models.SigninLog
	if err := db.Where("user_id = ?", user.ID).Order("signin_date desc").First(&last).Error; err == nil {
		if last.SigninDate == utils.Yesterday() {
			continuous = last.ContinuousDays + 1
		}
	}
	base := config.C.SigninPoints
	bonus := 0
	if continuous >= config.C.SigninStreakDays {
		bonus = config.C.SigninStreakBonus
	}
	total := base + bonus
	if err := AddPoints(db, user, total, "每日签到"); err != nil {
		return nil, err
	}
	sl := models.SigninLog{UserID: user.ID, SigninDate: today, Points: total, ContinuousDays: continuous}
	if err := db.Create(&sl).Error; err != nil {
		return nil, err
	}
	return map[string]any{
		"points":          total,
		"base_points":     base,
		"bonus":           bonus,
		"continuous_days": continuous,
		"signin_date":     today,
	}, nil
}

// ExchangeMembership 积分兑换普通会员。
func ExchangeMembership(db *gorm.DB, user *models.User) (map[string]any, error) {
	points := GetConfigInt(db, "exchange_points", config.C.ExchangePoints)
	days := GetConfigInt(db, "exchange_plan_days", config.C.ExchangePlanDays)
	if user.Points < points {
		return nil, utils.ErrBiz(400, 4002, "积分不足")
	}
	var plan models.Plan
	err := db.Where("name IN ?", []string{"普通会员", "普通"}).First(&plan).Error
	if err != nil {
		err = db.Where("id > 1").Order("sort").First(&plan).Error
		if err != nil {
			return nil, utils.ErrBiz(400, 9001, "套餐配置异常，请联系管理员")
		}
	}
	if err := AddPoints(db, user, -points, "积分兑换会员"); err != nil {
		return nil, err
	}
	now := time.Now()
	exp := now
	if user.PlanExpiresAt != nil && user.PlanExpiresAt.After(now) {
		exp = *user.PlanExpiresAt
	}
	exp = exp.AddDate(0, 0, days)
	user.PlanID = plan.ID
	user.PlanExpiresAt = &exp
	if err := db.Save(user).Error; err != nil {
		return nil, err
	}
	db.Create(&models.UserPlanLog{UserID: user.ID, PlanID: plan.ID, PlanName: plan.Name, Reason: "积分兑换会员"})
	return map[string]any{
		"plan_name":       plan.Name,
		"plan_expires_at": utils.TimeFmt(&exp),
		"points":          user.Points,
	}, nil
}

// GetConfig 读取系统配置（缺省回退默认值）。
func GetConfig(db *gorm.DB, key, def string) string {
	var sc models.SystemConfig
	if err := db.First(&sc, "key = ?", key).Error; err == nil {
		return sc.Value
	}
	return def
}

// GetConfigInt 读取整数配置。
func GetConfigInt(db *gorm.DB, key string, def int) int {
	v := GetConfig(db, key, "")
	if v == "" {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil {
		return n
	}
	return def
}

// SetConfig 写入系统配置。
func SetConfig(db *gorm.DB, key, value string) error {
	var sc models.SystemConfig
	err := db.First(&sc, "key = ?", key).Error
	if err == nil {
		sc.Value = value
		return db.Save(&sc).Error
	}
	return db.Create(&models.SystemConfig{Key: key, Value: value}).Error
}
