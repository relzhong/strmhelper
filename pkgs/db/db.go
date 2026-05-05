package db

import (
	"log"
	"path/filepath"

	"github.com/relzhong/strmhelper/pkgs/config"
	"github.com/relzhong/strmhelper/pkgs/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dbPath := filepath.Join(config.Settings.ConfigDir, "strmhelper.db")
	var err error

	logLevel := logger.Silent
	if config.Settings.Debug {
		logLevel = logger.Info
	}

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate
	err = DB.AutoMigrate(&models.StrmTask{}, &models.User{}, &models.SyncedFile{}, &models.PendingRemoteDelete{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
}
