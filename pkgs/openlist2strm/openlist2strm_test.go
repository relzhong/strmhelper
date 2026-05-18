package openlist2strm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/OpenListTeam/OpenList/v4/drivers/openlist"
	"github.com/go-resty/resty/v2"
)

func TestRemoveUsesCanceledContext(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &openListClient{
		Addition: openlist.Addition{Address: server.URL},
		resty:    resty.New(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := client.Remove(ctx, "/movies/file.mkv"); err == nil {
		t.Fatal("expected canceled context to abort remove request")
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("expected no remote requests, got %d", got)
	}
}
