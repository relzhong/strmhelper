package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/relzhong/strmhelper/pkgs/config"
	"github.com/relzhong/strmhelper/pkgs/db"
	"github.com/relzhong/strmhelper/pkgs/models"
	"github.com/relzhong/strmhelper/pkgs/openlist2strm"
	"github.com/relzhong/strmhelper/pkgs/utils"
	"github.com/relzhong/strmhelper/pkgs/web"
	"log/slog"
)

func printLogo() {
	fmt.Print(utils.LOGO)
	padding := (65 - len(config.Settings.AppName) - len(config.Settings.AppVersion) - 2) / 2
	line := ""
	for i := 0; i < padding; i++ {
		line += "="
	}
	fmt.Printf("%s %s %s %s\n\n", line, config.Settings.AppName, config.Settings.AppVersion, line)
}

func main() {
	printLogo()

	slog.Info(fmt.Sprintf("StrmHelper %s 启动中...", config.Settings.AppVersion))
	slog.Debug(fmt.Sprintf("是否开启 DEBUG 模式: %v", config.Settings.Debug))

	// Initialize Database
	db.InitDB()
	db.MigrateFromConfig()

	// Initialize Cron Manager
	openlist2strm.InitCronManager()

	// Load tasks from DB and add to cron
	var tasks []models.StrmTask
	db.DB.Find(&tasks)
	for _, task := range tasks {
		openlist2strm.Manager.AddTask(task)
	}

	// Setup Web Server
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))))

	// API Handlers
	mux.HandleFunc("/api/login", web.LoginHandler)
	mux.HandleFunc("/api/logout", web.LogoutHandler)
	mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			web.DeleteTaskHandler(w, r)
		} else if r.Method == http.MethodPost {
			web.CreateTaskHandler(w, r)
		}
	})
	mux.HandleFunc("/api/tasks/update", web.UpdateTaskHandler)
	mux.HandleFunc("/api/tasks/run", web.RunTaskHandler)

	// Admin UI (uses template)
	mux.HandleFunc("/ui/admin.html", web.TasksHandler)
	mux.HandleFunc("/ui/content", web.ContentHandler)

	// Root redirect
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/ui/", http.StatusMovedPermanently)
			return
		}
		http.NotFound(w, r)
	})

	// Apply Auth Middleware
	handler := web.AuthMiddleware(mux)

	go func() {
		slog.Info("Web server is listening on http://localhost:8080/ui/")
		if err := http.ListenAndServe(":8080", handler); err != nil {
			slog.Error(fmt.Sprintf("Web server failed: %v", err))
		}
	}()

	slog.Info("StrmHelper 启动完成")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("StrmHelper 程序退出！")
	openlist2strm.Manager.Stop()
}
