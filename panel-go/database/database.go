package database

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"weavenet/panel/models"
)

var DB *gorm.DB

// Init 打开 SQLite（WAL 模式，对齐原实现）并自动迁移。
func Init(path string) error {
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(30000)&_pragma=foreign_keys(1)"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}
	DB = db
	if err := db.AutoMigrate(
		&models.User{}, &models.Plan{}, &models.Node{}, &models.Tunnel{},
		&models.Domain{}, &models.SigninLog{}, &models.PointsLog{}, &models.Ticket{},
		&models.Announcement{}, &models.TrafficStat{}, &models.SystemConfig{},
		&models.Session{}, &models.EmailCode{}, &models.OperationLog{}, &models.UserPlanLog{},
	); err != nil {
		return err
	}
	log.Println("[db] SQLite 已连接")
	return nil
}

// Close 关闭数据库。
func Close() {
	if DB != nil {
		if sqlDB, err := DB.DB(); err == nil {
			sqlDB.Close()
		}
	}
}

// FileExists 判断文件是否存在。
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
