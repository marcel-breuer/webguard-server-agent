package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadAppliesSafeDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.json")
	content := `{"report_url":"https://app.example.test/api/v1/server-health/token","service_checks":[{"id":"health","type":"http","target":"http://127.0.0.1:8080/health"}]}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Interval.Value() != DefaultInterval || cfg.Retry.MaxAttempts != 3 || cfg.Queue.MaxBytes == 0 {
		t.Fatalf("safe defaults were not applied: %+v", cfg)
	}
}
func TestLoadRejectsInsecureConfigurationPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.json")
	if err := os.WriteFile(path, []byte(`{"report_url":"https://app.example.test/api/v1/server-health/token"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "owner-only") {
		t.Fatalf("error = %v", err)
	}
}
func TestConfigRejectsNonLocalServiceTarget(t *testing.T) {
	cfg := Config{ReportURL: "https://app.example.test/report", Interval: Duration(DefaultInterval), Retry: Retry{MaxAttempts: 1, BaseDelay: Duration(time.Second), MaxDelay: Duration(time.Second)}, Queue: Queue{Directory: "/tmp/queue", MaxBytes: 1024, MaxAge: Duration(time.Minute)}, ServiceChecks: []ServiceCheck{{ID: "external", Type: "http", Target: "https://example.com", Timeout: Duration(time.Second)}}}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "loopback") {
		t.Fatalf("error = %v", err)
	}
}
