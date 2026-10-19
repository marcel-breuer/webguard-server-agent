package delivery

import (
	"fmt"
	"github.com/marcel-breuer/webguard-server-agent/internal/config"
	"github.com/marcel-breuer/webguard-server-agent/internal/report"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestQueuePrunesReportsPastAgeLimit(t *testing.T) {
	directory := t.TempDir()
	queue := NewQueue(config.Queue{Directory: directory, MaxBytes: 1024, MaxAge: config.Duration(time.Hour)})
	base := time.Now()
	queue.now = func() time.Time { return base }
	id := "f1f1c5de-45d1-4d74-b5de-4d0c4df415e7"
	if err := queue.Enqueue(report.Payload{ReportID: id}); err != nil {
		t.Fatal(err)
	}
	expired := base.Add(-2 * time.Hour)
	if err := os.Chtimes(filepath.Join(directory, id+".json"), expired, expired); err != nil {
		t.Fatal(err)
	}
	if err := queue.Prune(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(directory, id+".json")); !os.IsNotExist(err) {
		t.Fatalf("expired report was not removed: %v", err)
	}
}
func TestQueuePrunesOldestReportsPastSizeLimit(t *testing.T) {
	directory := t.TempDir()
	queue := NewQueue(config.Queue{Directory: directory, MaxBytes: 1024, MaxAge: config.Duration(24 * time.Hour)})
	base := time.Now()
	queue.now = func() time.Time { return base }
	payload := report.Payload{Agent: report.Agent{Version: strings.Repeat("x", 550)}}
	payload.ReportID = "f1f1c5de-45d1-4d74-b5de-4d0c4df415e7"
	if err := queue.Enqueue(payload); err != nil {
		t.Fatal(err)
	}
	older := base.Add(-time.Minute)
	if err := os.Chtimes(filepath.Join(directory, payload.ReportID+".json"), older, older); err != nil {
		t.Fatal(err)
	}
	firstID := payload.ReportID
	payload.ReportID = "f1f1c5de-45d1-4d74-b5de-4d0c4df415e8"
	if err := queue.Enqueue(payload); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(directory, firstID+".json")); !os.IsNotExist(err) {
		t.Fatalf("oldest report was not removed: %v", err)
	}
}

func TestQueueRejectsOversizedReportsAndTraversalIDs(t *testing.T) {
	queue := NewQueue(config.Queue{Directory: t.TempDir(), MaxBytes: 1024, MaxAge: config.Duration(24 * time.Hour)})
	oversized := report.Payload{ReportID: "f1f1c5de-45d1-4d74-b5de-4d0c4df415e7", Agent: report.Agent{Version: strings.Repeat("x", report.MaxPayloadBytes)}}
	if err := queue.Enqueue(oversized); err == nil {
		t.Fatal("expected oversized queued report to be rejected")
	}
	if err := queue.Delete("../../outside"); err == nil {
		t.Fatal("expected traversal report id to be rejected")
	}
}

func TestQueueStatsExposeBoundedDeliveryBacklog(t *testing.T) {
	directory := t.TempDir()
	queue := NewQueue(config.Queue{Directory: directory, MaxBytes: 1024, MaxAge: config.Duration(time.Hour)})
	base := time.Now()
	queue.now = func() time.Time { return base }
	id := "f1f1c5de-45d1-4d74-b5de-4d0c4df415e7"
	if err := queue.Enqueue(report.Payload{ReportID: id}); err != nil {
		t.Fatal(err)
	}
	oldest := base.Add(-2 * time.Minute)
	if err := os.Chtimes(filepath.Join(directory, id+".json"), oldest, oldest); err != nil {
		t.Fatal(err)
	}
	stats, err := queue.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.Count != 1 || stats.Bytes == 0 || stats.MaxBytes != 1024 || stats.MaxCount != MaxQueueReports || stats.OldestAge != 2*time.Minute {
		t.Fatalf("unexpected queue statistics: %+v", stats)
	}
}

func TestQueuePrunesExcessReportCount(t *testing.T) {
	directory := t.TempDir()
	queue := NewQueue(config.Queue{Directory: directory, MaxBytes: config.MaxQueueBytes, MaxAge: config.Duration(24 * time.Hour)})
	base := time.Now()
	queue.now = func() time.Time { return base }
	for i := 0; i <= MaxQueueReports; i++ {
		id := fmt.Sprintf("00000000-0000-4000-8000-%012x", i)
		path := filepath.Join(directory, id+".json")
		if err := os.WriteFile(path, []byte(`{}`), 0o600); err != nil {
			t.Fatal(err)
		}
		modified := base.Add(time.Duration(i) * time.Millisecond)
		if err := os.Chtimes(path, modified, modified); err != nil {
			t.Fatal(err)
		}
	}
	if err := queue.Prune(); err != nil {
		t.Fatal(err)
	}
	stats, err := queue.Stats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.Count != MaxQueueReports {
		t.Fatalf("queued report count = %d, want %d", stats.Count, MaxQueueReports)
	}
}
