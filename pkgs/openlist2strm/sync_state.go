package openlist2strm

import (
	"log/slog"
	"path/filepath"
	"github.com/relzhong/strmhelper/pkgs/db"
	"github.com/relzhong/strmhelper/pkgs/models"
)

type SyncStateManager struct {
	TaskID     string
	TargetDir  string
	GraceScans int
}

func NewSyncStateManager(taskID string, targetDir string, graceScans int) *SyncStateManager {
	return &SyncStateManager{
		TaskID:     taskID,
		TargetDir:  targetDir,
		GraceScans: graceScans,
	}
}

func (s *SyncStateManager) toRelative(path string) string {
	rel, err := filepath.Rel(s.TargetDir, path)
	if err != nil {
		return path
	}
	return rel
}

// IsKnown checks if the file was previously synced
func (s *SyncStateManager) IsKnown(localPath string) bool {
	localPath = s.toRelative(localPath)
	var synced models.SyncedFile
	err := db.DB.Where("task_id = ? AND local_path = ?", s.TaskID, localPath).First(&synced).Error
	return err == nil
}

// MarkSynced records a successful sync
func (s *SyncStateManager) MarkSynced(localPath, remotePath string) {
	localPath = s.toRelative(localPath)
	synced := models.SyncedFile{
		TaskID:     s.TaskID,
		LocalPath:  localPath,
		RemotePath: remotePath,
	}
	// Use Clause OnConflict to update if exists
	db.DB.Where(models.SyncedFile{TaskID: s.TaskID, LocalPath: localPath}).
		Assign(models.SyncedFile{RemotePath: remotePath}).
		FirstOrCreate(&synced)
}

// Unmark removes a record after deletion
func (s *SyncStateManager) Unmark(localPath string) {
	localPath = s.toRelative(localPath)
	var remotePath string
	db.DB.Model(&models.SyncedFile{}).Where("task_id = ? AND local_path = ?", s.TaskID, localPath).Pluck("remote_path", &remotePath)

	db.DB.Where("task_id = ? AND local_path = ?", s.TaskID, localPath).Delete(&models.SyncedFile{})
	if remotePath != "" {
		db.DB.Where("task_id = ? AND remote_path = ?", s.TaskID, remotePath).Delete(&models.PendingRemoteDelete{})
	}
}

// HandleMissingLocal handles a remote file that is missing locally
// returns true if the remote file should be deleted
func (s *SyncStateManager) HandleMissingLocal(localPath, remotePath string) bool {
	localPath = s.toRelative(localPath)
	if s.GraceScans <= 0 {
		return true // No protection, delete immediately
	}

	var pending models.PendingRemoteDelete
	err := db.DB.Where("task_id = ? AND remote_path = ?", s.TaskID, remotePath).First(&pending).Error
	if err != nil {
		// First time missing
		pending = models.PendingRemoteDelete{
			TaskID:       s.TaskID,
			RemotePath:   remotePath,
			MissingCount: 1,
		}
		db.DB.Create(&pending)
		slog.Info("Remote file missing locally, marked for pending delete", "path", remotePath)
		return false
	}

	pending.MissingCount++
	db.DB.Save(&pending)

	if pending.MissingCount >= s.GraceScans {
		slog.Warn("Remote file missing locally for grace period, triggering delete", "path", remotePath, "count", pending.MissingCount)
		return true
	}

	slog.Info("Remote file missing locally, pending delete count incremented", "path", remotePath, "count", pending.MissingCount)
	return false
}

// ClearPending removes a file from the pending delete list (it reappeared)
func (s *SyncStateManager) ClearPending(remotePath string) {
	db.DB.Where("task_id = ? AND remote_path = ?", s.TaskID, remotePath).Delete(&models.PendingRemoteDelete{})
}

// GetSyncedCount returns the number of synced files for this task
func (s *SyncStateManager) GetSyncedCount() int64 {
	var count int64
	db.DB.Model(&models.SyncedFile{}).Where("task_id = ?", s.TaskID).Count(&count)
	return count
}
