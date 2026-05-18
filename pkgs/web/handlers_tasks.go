package web

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/relzhong/strmhelper/pkgs/db"
	"github.com/relzhong/strmhelper/pkgs/models"
	"github.com/relzhong/strmhelper/pkgs/openlist2strm"
	"github.com/robfig/cron/v3"
)

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	var tasks []models.StrmTask
	db.DB.Find(&tasks)

	for i := range tasks {
		tasks[i].NextRun = openlist2strm.Manager.GetNextRun(tasks[i].TaskID)
	}

	tmpl := template.Must(template.ParseFiles("ui/admin.html"))
	tmpl.Execute(w, tasks)
}

func CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	task := models.StrmTask{
		TaskID:           r.FormValue("task_id"),
		Cron:             r.FormValue("cron"),
		URL:              r.FormValue("url"),
		PublicURL:        r.FormValue("public_url"),
		Username:         r.FormValue("username"),
		Password:         r.FormValue("password"),
		Token:            r.FormValue("token"),
		SourceDir:        r.FormValue("source_dir"),
		TargetDir:        r.FormValue("target_dir"),
		FlattenMode:      r.FormValue("flatten_mode") == "on",
		Subtitle:         r.FormValue("subtitle") == "on",
		Image:            r.FormValue("image") == "on",
		NFO:              r.FormValue("nfo") == "on",
		Mode:             r.FormValue("mode"),
		Overwrite:        r.FormValue("overwrite") == "on",
		SyncServer:       r.FormValue("sync_server") == "on",
		CheckModTime:     r.FormValue("check_mod_time") == "on",
		SyncBack:         r.FormValue("sync_back") == "on",
		SyncIgnore:       r.FormValue("sync_ignore"),
		ProtectEnabled:   r.FormValue("protect_enabled") == "on",
		ProtectThreshold: getFormInt(r, "protect_threshold", 100),
		ProtectGrace:     getFormInt(r, "protect_grace", 3),
		OtherExt:         r.FormValue("other_ext"),
		MaxWorkers:       getFormInt(r, "max_workers", 50),
		MaxDownloaders:   getFormInt(r, "max_downloaders", 5),
		WaitTime:         getFormInt(r, "wait_time", 0),
	}

	if task.Cron != "" {
		if _, err := cron.ParseStandard(task.Cron); err != nil {
			http.Error(w, "Invalid cron expression: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if err := db.DB.Create(&task).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	openlist2strm.Manager.AddTask(task)
	w.Header().Set("HX-Redirect", "/ui/")
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	var task models.StrmTask
	if err := db.DB.First(&task, id).Error; err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodPost {
		task.TaskID = r.FormValue("task_id")
		task.Cron = r.FormValue("cron")
		task.URL = r.FormValue("url")
		task.PublicURL = r.FormValue("public_url")
		task.Username = r.FormValue("username")
		task.Password = r.FormValue("password")
		task.Token = r.FormValue("token")
		task.SourceDir = r.FormValue("source_dir")
		task.TargetDir = r.FormValue("target_dir")
		task.FlattenMode = r.FormValue("flatten_mode") == "on"
		task.Subtitle = r.FormValue("subtitle") == "on"
		task.Image = r.FormValue("image") == "on"
		task.NFO = r.FormValue("nfo") == "on"
		task.Mode = r.FormValue("mode")
		task.Overwrite = r.FormValue("overwrite") == "on"
		task.SyncServer = r.FormValue("sync_server") == "on"
		task.CheckModTime = r.FormValue("check_mod_time") == "on"
		task.SyncBack = r.FormValue("sync_back") == "on"
		task.SyncIgnore = r.FormValue("sync_ignore")
		task.ProtectEnabled = r.FormValue("protect_enabled") == "on"
		task.ProtectThreshold = getFormInt(r, "protect_threshold", 100)
		task.ProtectGrace = getFormInt(r, "protect_grace", 3)
		task.OtherExt = r.FormValue("other_ext")
		task.MaxWorkers = getFormInt(r, "max_workers", 50)
		task.MaxDownloaders = getFormInt(r, "max_downloaders", 5)
		task.WaitTime = getFormInt(r, "wait_time", 0)

		if task.Cron != "" {
			if _, err := cron.ParseStandard(task.Cron); err != nil {
				http.Error(w, "Invalid cron expression: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		db.DB.Save(&task)
		openlist2strm.Manager.AddTask(task)
		w.Header().Set("HX-Redirect", "/ui/")
		return
	}
}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	var task models.StrmTask
	if err := db.DB.First(&task, id).Error; err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if openlist2strm.Manager.IsTaskRunning(task.TaskID) {
		http.Error(w, "Task is still running; try again after it finishes.", http.StatusConflict)
		return
	}

	openlist2strm.Manager.RemoveTask(task.TaskID)

	// Clean up related data (Hard delete)
	db.DB.Unscoped().Where("task_id = ?", task.TaskID).Delete(&models.SyncedFile{})
	db.DB.Unscoped().Where("task_id = ?", task.TaskID).Delete(&models.PendingRemoteDelete{})

	db.DB.Unscoped().Delete(&models.StrmTask{}, id)

	w.Header().Set("HX-Redirect", "/ui/")
}

func RunTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.FormValue("id")
	}
	slog.Info("RunTaskHandler called", "idStr", idStr)

	if idStr == "" {
		slog.Warn("RunTaskHandler: missing id parameter")
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("RunTaskHandler: invalid id parameter", "idStr", idStr, "error", err)
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var task models.StrmTask
	if err := db.DB.First(&task, id).Error; err != nil {
		slog.Error("RunTaskHandler: task not found", "id", id, "error", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	slog.Info("Manual task trigger received", "id", task.TaskID)
	go openlist2strm.Manager.RunTask(task)

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func getFormInt(r *http.Request, key string, def int) int {
	valStr := r.FormValue(key)
	if valStr == "" {
		return def
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return def
	}
	return val
}
