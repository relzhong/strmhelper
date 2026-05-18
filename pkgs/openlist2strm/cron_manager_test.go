package openlist2strm

import (
	"context"
	"testing"

	"github.com/robfig/cron/v3"
)

func TestIsTaskRunningTracksActiveRuns(t *testing.T) {
	m := &CronManager{
		cron:        cron.New(),
		tasks:       make(map[string]cron.EntryID),
		cancelFuncs: make(map[string]context.CancelFunc),
		activeRuns:  map[string]int{"task-a": 1},
	}

	if !m.IsTaskRunning("task-a") {
		t.Fatal("expected task-a to be running")
	}
	if m.IsTaskRunning("task-b") {
		t.Fatal("expected task-b not to be running")
	}
}
