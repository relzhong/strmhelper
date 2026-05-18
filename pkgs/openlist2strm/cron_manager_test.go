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
		runIDs:      make(map[string]uint64),
	}

	if !m.IsTaskRunning("task-a") {
		t.Fatal("expected task-a to be running")
	}
	if m.IsTaskRunning("task-b") {
		t.Fatal("expected task-b not to be running")
	}
}

func TestFinishRunKeepsNewerCancelFunc(t *testing.T) {
	_, newCancel := context.WithCancel(context.Background())
	m := &CronManager{
		cancelFuncs: map[string]context.CancelFunc{"task-a": newCancel},
		activeRuns:  map[string]int{"task-a": 2},
		runIDs:      map[string]uint64{"task-a": 2},
	}

	m.finishRun("task-a", 1)

	if got := m.activeRuns["task-a"]; got != 1 {
		t.Fatalf("activeRuns = %d", got)
	}
	if _, ok := m.cancelFuncs["task-a"]; !ok {
		t.Fatal("expected newer cancel func to remain")
	}
	if _, ok := m.runIDs["task-a"]; !ok {
		t.Fatal("expected newer run id to remain")
	}
}
