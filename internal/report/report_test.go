package report

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestPayloadSerializesCoreContract(t *testing.T) {
	status := 200
	payload := Payload{SchemaVersion: SchemaVersion, ReportID: "f1f1c5de-45d1-4d74-b5de-4d0c4df415e7", SampledAt: time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC), Host: Host{CPUUsagePercent: 42.5, LogicalCPUCount: 4, LoadAverage1m: 1.42, LoadAverage5m: 1.1, LoadAverage15m: 0.94, RAMUsagePercent: 68.2, SwapUsagePercent: 4.1, UptimeSeconds: 86400}, ServiceChecks: []ServiceCheck{{ID: "app-health", Success: true, ResponseTimeMS: 38.4, StatusCode: &status}}, Agent: Agent{Version: "1.0.0"}}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var actual map[string]any
	if err := json.Unmarshal(encoded, &actual); err != nil {
		t.Fatal(err)
	}
	if actual["schema_version"] != float64(1) || actual["report_id"] != payload.ReportID || actual["sampled_at"] != "2026-08-08T12:00:00Z" {
		t.Fatalf("unexpected contract fields: %s", encoded)
	}
}

func TestCoreVersionOneFixtureCanBeDecoded(t *testing.T) {
	content, err := os.ReadFile("../../testdata/core-v1-report.json")
	if err != nil {
		t.Fatal(err)
	}
	var payload Payload
	if err := json.Unmarshal(content, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.SchemaVersion != SchemaVersion || payload.ReportID == "" || payload.Host.LogicalCPUCount == 0 {
		t.Fatalf("invalid fixture payload: %+v", payload)
	}
}
