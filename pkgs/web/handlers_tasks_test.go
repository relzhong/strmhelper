package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/relzhong/strmhelper/pkgs/db"
	"github.com/relzhong/strmhelper/pkgs/models"
	"github.com/relzhong/strmhelper/pkgs/openlist2strm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTaskTestDB(t *testing.T) {
	t.Helper()

	var err error
	db.DB, err = gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.DB.AutoMigrate(&models.StrmTask{}, &models.SyncedFile{}, &models.PendingRemoteDelete{}); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	openlist2strm.InitCronManager()
	t.Cleanup(func() { openlist2strm.Manager.Stop() })
}

func TestDeleteTaskHandlerDeletesIdleTask(t *testing.T) {
	setupTaskTestDB(t)
	task := models.StrmTask{TaskID: "idle-task"}
	db.DB.Create(&task)
	db.DB.Create(&models.SyncedFile{TaskID: task.TaskID, LocalPath: "a", RemotePath: "/a"})

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/tasks?id=%d", task.ID), nil)
	res := httptest.NewRecorder()
	DeleteTaskHandler(res, req)

	if got := res.Code; got != http.StatusOK {
		t.Fatalf("status = %d", got)
	}
	var count int64
	db.DB.Model(&models.StrmTask{}).Where("task_id = ?", task.TaskID).Count(&count)
	if count != 0 {
		t.Fatal("expected task to be deleted")
	}
	db.DB.Model(&models.SyncedFile{}).Where("task_id = ?", task.TaskID).Count(&count)
	if count != 0 {
		t.Fatal("expected synced state to be deleted")
	}
}

func TestDeleteTaskHandlerRejectsRunningTask(t *testing.T) {
	setupTaskTestDB(t)

	listStarted := make(chan struct{})
	releaseList := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/me":
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": map[string]any{}})
		case "/api/fs/list":
			close(listStarted)
			<-releaseList
			_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": map[string]any{"content": []any{}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	task := models.StrmTask{TaskID: "running-task", URL: server.URL, SourceDir: "/", TargetDir: t.TempDir()}
	db.DB.Create(&task)
	db.DB.Create(&models.SyncedFile{TaskID: task.TaskID, LocalPath: "a", RemotePath: "/a"})

	done := make(chan struct{})
	go func() {
		openlist2strm.Manager.RunTask(task)
		close(done)
	}()

	select {
	case <-listStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("task did not start")
	}

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/tasks?id=%d", task.ID), nil)
	res := httptest.NewRecorder()
	DeleteTaskHandler(res, req)

	if got := res.Code; got != http.StatusConflict {
		t.Fatalf("status = %d", got)
	}
	var count int64
	db.DB.Model(&models.StrmTask{}).Where("task_id = ?", task.TaskID).Count(&count)
	if count != 1 {
		t.Fatal("expected running task to remain")
	}
	db.DB.Model(&models.SyncedFile{}).Where("task_id = ?", task.TaskID).Count(&count)
	if count != 1 {
		t.Fatal("expected synced state to remain")
	}

	close(releaseList)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("task did not finish")
	}
}
