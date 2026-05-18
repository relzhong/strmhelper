package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContentHandlerDisablesCaching(t *testing.T) {
	t.Chdir("../..")

	req := httptest.NewRequest(http.MethodGet, "/ui/content", nil)
	res := httptest.NewRecorder()

	ContentHandler(res, req)

	if got := res.Header().Get("Cache-Control"); got != "no-store, no-cache, must-revalidate" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if got := res.Header().Get("Vary"); got != "Cookie" {
		t.Fatalf("Vary = %q", got)
	}
}

func TestLoginHandlerDisablesCaching(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/login", nil)
	res := httptest.NewRecorder()

	LoginHandler(res, req)

	if got := res.Header().Get("Cache-Control"); got != "no-store, no-cache, must-revalidate" {
		t.Fatalf("Cache-Control = %q", got)
	}
}
