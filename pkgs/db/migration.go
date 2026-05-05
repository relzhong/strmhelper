package db

import (
	"log/slog"

	"github.com/relzhong/strmhelper/pkgs/config"
	"github.com/relzhong/strmhelper/pkgs/models"
)

func MigrateFromConfig() {
	var count int64
	DB.Model(&models.StrmTask{}).Count(&count)
	if count > 0 {
		return // Already migrated or has data
	}

	if len(config.Settings.Config.OpenList2StrmList) == 0 {
		return
	}

	slog.Info("Migrating tasks from config.yaml to database...")

	for _, item := range config.Settings.Config.OpenList2StrmList {
		task := models.StrmTask{
			TaskID:         getString(item, "id"),
			Cron:           getString(item, "cron"),
			URL:            getString(item, "url"),
			PublicURL:      getString(item, "public_url"),
			Username:       getString(item, "username"),
			Password:       getString(item, "password"),
			Token:          getString(item, "token"),
			SourceDir:      getString(item, "source_dir"),
			TargetDir:      getString(item, "target_dir"),
			FlattenMode:    getBool(item, "flatten_mode"),
			Subtitle:       getBool(item, "subtitle"),
			Image:          getBool(item, "image"),
			NFO:            getBool(item, "nfo"),
			Mode:           getString(item, "mode"),
			Overwrite:      getBool(item, "overwrite"),
			SyncServer:     getBool(item, "sync_server"),
			SyncIgnore:     getString(item, "sync_ignore"),
			OtherExt:       getString(item, "other_ext"),
			MaxWorkers:     getInt(item, "max_workers", 50),
			MaxDownloaders: getInt(item, "max_downloaders", 5),
			WaitTime:       getInt(item, "wait_time", 0),
			CheckModTime:    getBool(item, "check_mod_time"),
		}

		// Handle nested smart_protection
		if sp, ok := item["smart_protection"].(map[string]interface{}); ok {
			task.ProtectEnabled = getBool(sp, "enabled")
			task.ProtectThreshold = getInt(sp, "threshold", 100)
			task.ProtectGrace = getInt(sp, "grace_scans", 3)
		}

		if err := DB.Create(&task).Error; err != nil {
			slog.Error("Failed to migrate task", "id", task.TaskID, "error", err)
		} else {
			slog.Info("Migrated task", "id", task.TaskID)
		}
	}
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

func getInt(m map[string]interface{}, key string, def int) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	return def
}
