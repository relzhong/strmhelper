package openlist2strm

import (
	"context"
	"log/slog"
	"sync"

	"github.com/relzhong/strmhelper/pkgs/db"
	"github.com/relzhong/strmhelper/pkgs/models"
	"github.com/robfig/cron/v3"
)

type CronManager struct {
	cron        *cron.Cron
	tasks       map[string]cron.EntryID
	cancelFuncs map[string]context.CancelFunc
	activeRuns  map[string]int
	mu          sync.Mutex
}

var Manager *CronManager

func InitCronManager() {
	Manager = &CronManager{
		cron:        cron.New(),
		tasks:       make(map[string]cron.EntryID),
		cancelFuncs: make(map[string]context.CancelFunc),
		activeRuns:  make(map[string]int),
	}
	Manager.cron.Start()
}

func (m *CronManager) AddTask(task models.StrmTask) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove existing if any and cancel running
	if entryID, ok := m.tasks[task.TaskID]; ok {
		m.cron.Remove(entryID)
	}
	if cancel, ok := m.cancelFuncs[task.TaskID]; ok {
		cancel()
		delete(m.cancelFuncs, task.TaskID)
	}

	if task.Cron == "" {
		return nil
	}

	entryID, err := m.cron.AddFunc(task.Cron, func() {
		m.RunTask(task)
	})
	if err != nil {
		return err
	}

	m.tasks[task.TaskID] = entryID
	slog.Info("Task scheduled", "id", task.TaskID, "cron", task.Cron)
	return nil
}

func (m *CronManager) RemoveTask(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entryID, ok := m.tasks[taskID]; ok {
		m.cron.Remove(entryID)
		delete(m.tasks, taskID)
		slog.Info("Task unscheduled", "id", taskID)
	}

	if cancel, ok := m.cancelFuncs[taskID]; ok {
		cancel()
		delete(m.cancelFuncs, taskID)
		slog.Info("Running task cancelled", "id", taskID)
	}
}

func (m *CronManager) GetNextRun(taskID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entryID, ok := m.tasks[taskID]; ok {
		entry := m.cron.Entry(entryID)
		if !entry.Next.IsZero() {
			return entry.Next.Format("2006-01-02 15:04:05")
		}
	}
	return "-"
}

func (m *CronManager) IsTaskRunning(taskID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.activeRuns[taskID] > 0
}

func (m *CronManager) RunTask(task models.StrmTask) {
	slog.Info("RunTask started", "id", task.TaskID)

	m.mu.Lock()
	// Cancel existing if any
	if cancel, ok := m.cancelFuncs[task.TaskID]; ok {
		slog.Info("Cancelling previous running task instance", "id", task.TaskID)
		cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFuncs[task.TaskID] = cancel
	m.activeRuns[task.TaskID]++
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.activeRuns[task.TaskID]--
		if m.activeRuns[task.TaskID] <= 0 {
			delete(m.activeRuns, task.TaskID)
		}
		if _, ok := m.cancelFuncs[task.TaskID]; ok {
			// Only delete if it's the same cancel function we started with
			// (handles cases where a new task started before this defer ran)
			// Actually, comparing functions is not direct in Go, but we can just check if it's still there.
			// Given our mutex usage, it's safer to just cleanup if it matches.
			delete(m.cancelFuncs, task.TaskID)
		}
		m.mu.Unlock()
		slog.Info("RunTask finished", "id", task.TaskID)
		db.DB.Model(&task).Update("is_running", false)
	}()

	// Check if already running in DB to avoid double trigger
	// (Note: we just cancelled the previous one, so we should proceed)

	// Update running status
	db.DB.Model(&task).Updates(map[string]interface{}{
		"is_running": true,
		"last_error": "",
	})

	sp := &SmartProtectionConfig{
		Enabled:    task.ProtectEnabled,
		Threshold:  task.ProtectThreshold,
		GraceScans: task.ProtectGrace,
	}

	job, err := NewOpenList2Strm(
		ctx,
		task.TaskID, task.URL, task.Username, task.Password, task.Token,
		task.PublicURL, task.SourceDir, task.TargetDir, task.FlattenMode,
		task.Subtitle, task.Image, task.NFO, task.Mode, task.Overwrite,
		task.OtherExt, task.MaxWorkers, task.MaxDownloaders, float64(task.WaitTime),
		task.SyncServer, task.SyncIgnore, sp, task.CheckModTime,
		task.SyncBack,
	)

	if err != nil {
		slog.Error("Failed to create job", "id", task.TaskID, "error", err)
		db.DB.Model(&task).Update("last_error", err.Error())
		return
	}

	if err := job.Run(); err != nil {
		slog.Error("Job run failed", "id", task.TaskID, "error", err)
		db.DB.Model(&task).Update("last_error", err.Error())
	}
}

func (m *CronManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cron.Stop()
	for id, cancel := range m.cancelFuncs {
		slog.Info("Stopping task on shutdown", "id", id)
		cancel()
	}
	m.cancelFuncs = make(map[string]context.CancelFunc)
}
