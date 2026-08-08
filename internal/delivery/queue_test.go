package delivery

import (
	"github.com/marcel-breuer/webguard-server-agent/internal/config"
	"github.com/marcel-breuer/webguard-server-agent/internal/report"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestQueuePrunesReportsPastAgeLimit(t *testing.T) {
	directory := t.TempDir()
	queue := NewQueue(config.Queue{Directory: directory, MaxBytes: 1024, MaxAge: config.Duration(time.Hour)})
	base := time.Now()
	queue.now = func() time.Time { return base }
	if err := queue.Enqueue(report.Payload{ReportID: "expired"}); err != nil {
		t.Fatal(err)
	}
	expired := base.Add(-2 * time.Hour)
	if err := os.Chtimes(filepath.Join(directory, "expired.json"), expired, expired); err != nil {
		t.Fatal(err)
	}
	if err := queue.Prune(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(directory, "expired.json")); !os.IsNotExist(err) {
		t.Fatalf("expired report was not removed: %v", err)
	}
}
func TestQueuePrunesOldestReportsPastSizeLimit(t *testing.T) {
	directory := t.TempDir()
	queue := NewQueue(config.Queue{Directory: directory, MaxBytes: 500, MaxAge: config.Duration(24 * time.Hour)})
	base := time.Now()
	queue.now = func() time.Time { return base }
	payload := report.Payload{Agent: report.Agent{Version: "a long value that fills the queue"}}
	payload.ReportID = "first"
	if err := queue.Enqueue(payload); err != nil {
		t.Fatal(err)
	}
	older := base.Add(-time.Minute)
	if err := os.Chtimes(filepath.Join(directory, "first.json"), older, older); err != nil {
		t.Fatal(err)
	}
	payload.ReportID = "second"
	if err := queue.Enqueue(payload); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(directory, "first.json")); !os.IsNotExist(err) {
		t.Fatalf("oldest report was not removed: %v", err)
	}
}
