package delivery

import (
	"context"
	"github.com/marcel-breuer/webguard-server-agent/internal/config"
	"github.com/marcel-breuer/webguard-server-agent/internal/report"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testClient(server *httptest.Server) *Client {
	return &Client{httpClient: server.Client(), reportURL: server.URL, retry: config.Retry{MaxAttempts: 2, BaseDelay: config.Duration(time.Millisecond), MaxDelay: config.Duration(time.Millisecond)}, sleep: func(time.Duration) {}}
}
func TestDeliverAcceptsSuccessfulCoreResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Content-Type") != "application/json" || request.Header.Get("Accept") != "application/json" {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	if err := testClient(server).Deliver(context.Background(), report.Payload{}); err != nil {
		t.Fatal(err)
	}
}
func TestDeliverDoesNotRetryAuthenticationFailure(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests++
		writer.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	if err := testClient(server).Deliver(context.Background(), report.Payload{}); err == nil {
		t.Fatal("expected error")
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
}
func TestDeliverRetriesServerFailure(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests++
		if requests == 1 {
			writer.WriteHeader(http.StatusBadGateway)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	if err := testClient(server).Deliver(context.Background(), report.Payload{}); err != nil {
		t.Fatal(err)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}
