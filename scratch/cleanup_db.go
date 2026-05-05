package main

import (
	"log"
	"path/filepath"
	"os"

	"github.com/relzhong/strmhelper/pkgs/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Manual config load if needed, but we can just guess the path
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".strmhelper", "strmhelper.db")
	// If the above is wrong, let's try current dir or common locations
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dbPath = "strmhelper.db" // check local
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Printf("Cleaning up soft-deleted tasks from %s...", dbPath)

	result := db.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.StrmTask{})
	log.Printf("Hard-deleted %d soft-deleted tasks.", result.RowsAffected)

	result = db.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.SyncedFile{})
	log.Printf("Hard-deleted %d soft-deleted synced files.", result.RowsAffected)

	result = db.Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.PendingRemoteDelete{})
	log.Printf("Hard-deleted %d soft-deleted pending deletes.", result.RowsAffected)

	log.Println("Cleanup complete.")
}
