package delivery

import (
	"bytes"
	"context"
	"fmt"
	"github.com/marcel-breuer/webguard-server-agent/internal/config"
	"github.com/marcel-breuer/webguard-server-agent/internal/report"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testClient(server *httptest.Server) *Client {
	return &Client{httpClient: newHTTPClient(server.Client().Transport), reportURL: server.URL, retry: config.Retry{MaxAttempts: 2, BaseDelay: config.Duration(time.Millisecond), MaxDelay: config.Duration(time.Millisecond)}, sleep: func(time.Duration) {}}
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

func TestDeliverRejectsOversizedReportBeforeSending(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests++
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	payload := report.Payload{Agent: report.Agent{Version: strings.Repeat("x", report.MaxPayloadBytes)}}
	if err := testClient(server).Deliver(context.Background(), payload); err == nil || !strings.Contains(err.Error(), "maximum encoded size") {
		t.Fatalf("error = %v, want oversized report error", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want no network request", requests)
	}
}

func TestDeliverRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(bytes.Repeat([]byte("x"), maxResponseBytes+1))
	}))
	defer server.Close()
	err := testClient(server).Deliver(context.Background(), report.Payload{})
	if err == nil || !strings.Contains(err.Error(), "response exceeds maximum size") {
		t.Fatalf("error = %v, want oversized response error", err)
	}
}

func TestDeliverDoesNotFollowRedirects(t *testing.T) {
	var redirected bool
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/target" {
			redirected = true
			writer.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(writer, request, "/target", http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	err := testClient(server).Deliver(context.Background(), report.Payload{})
	if err == nil || !strings.Contains(err.Error(), fmt.Sprint(http.StatusTemporaryRedirect)) {
		t.Fatalf("error = %v, want redirect rejection", err)
	}
	if redirected {
		t.Fatal("client followed a report endpoint redirect")
	}
}

func TestDeliverRetriesTransportFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := server.URL
	server.Close()
	sleeps := 0
	client := &Client{httpClient: &http.Client{Timeout: time.Second}, reportURL: url, retry: config.Retry{MaxAttempts: 3, BaseDelay: config.Duration(time.Millisecond), MaxDelay: config.Duration(time.Millisecond)}, sleep: func(time.Duration) { sleeps++ }}
	if err := client.Deliver(context.Background(), report.Payload{}); err == nil || !IsRetryable(err) {
		t.Fatalf("error = %v, want retryable transport failure", err)
	}
	if sleeps != 2 {
		t.Fatalf("retry delays = %d, want 2", sleeps)
	}
}

func TestBackoffNeverExceedsConfiguredMaximum(t *testing.T) {
	retry := config.Retry{BaseDelay: config.Duration(5 * time.Minute), MaxDelay: config.Duration(5 * time.Minute)}
	for attempt := 1; attempt <= 10; attempt++ {
		if delay := backoff(retry, attempt); delay < 0 || delay > retry.MaxDelay.Value() {
			t.Fatalf("attempt %d delay = %s, want between zero and %s", attempt, delay, retry.MaxDelay.Value())
		}
	}
}

func BenchmarkDeliverReport(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client := testClient(server)
	client.retry.MaxAttempts = 1
	payload := report.Payload{SchemaVersion: report.SchemaVersion, ReportID: "f1f1c5de-45d1-4d74-b5de-4d0c4df415e7", SampledAt: time.Now().UTC(), Agent: report.Agent{Version: "bench"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := client.Deliver(context.Background(), payload); err != nil {
			b.Fatal(err)
		}
	}
}
