package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/relzhong/strmhelper/pkgs/config"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login api and static assets
		if r.URL.Path == "/api/login" || (len(r.URL.Path) > 10 && r.URL.Path[:11] == "/ui/static/") {
			next.ServeHTTP(w, r)
			return
		}

		// Allow /ui/ to serve index.html (which is the shell)
		if r.URL.Path == "/ui/" || r.URL.Path == "/ui/index.html" {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != "authenticated" {
			if r.Header.Get("HX-Request") != "" {
				// For HTMX requests, we can't just redirect to a partial if they expect a full page
				// but here we actually want to show the login partial.
				// However, a 401 might be better handled by the client.
				// For now, let's just return the login partial if they are requesting content.
				if r.URL.Path == "/ui/content" {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Redirect(w, r, "/ui/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ContentHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value != "authenticated" {
		http.ServeFile(w, r, "ui/login.html")
		return
	}
	TasksHandler(w, r)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == config.Settings.Config.Web.Username && password == config.Settings.Config.Web.Password {
		http.SetCookie(w, &http.Cookie{
			Name:    "session",
			Value:   "authenticated",
			Path:    "/",
			Expires: time.Now().Add(24 * time.Hour),
		})
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprint(w, "<small style='color: var(--pico-secondary)'>Invalid credentials</small>")
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-1 * time.Hour),
	})
	http.Redirect(w, r, "/ui/login.html", http.StatusSeeOther)
}
