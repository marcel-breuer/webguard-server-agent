package checks

import (
	"context"
	"github.com/marcel-breuer/webguard-server-agent/internal/config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRunReportsHTTPResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) { writer.WriteHeader(http.StatusNoContent) }))
	defer server.Close()
	target := strings.Replace(server.URL, "127.0.0.1", "localhost", 1)
	result := Run(context.Background(), []config.ServiceCheck{{ID: "health", Type: "http", Target: target, Timeout: config.Duration(time.Second)}})
	if len(result) != 1 || !result[0].Success || result[0].StatusCode == nil || *result[0].StatusCode != http.StatusNoContent || result[0].ResponseTimeMS < 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}
