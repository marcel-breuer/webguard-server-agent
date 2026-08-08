package delivery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/marcel-breuer/webguard-server-agent/internal/config"
	"github.com/marcel-breuer/webguard-server-agent/internal/report"
)

type Queue struct {
	directory string
	maxBytes  int64
	maxAge    time.Duration
	now       func() time.Time
}

func NewQueue(cfg config.Queue) *Queue {
	return &Queue{directory: cfg.Directory, maxBytes: cfg.MaxBytes, maxAge: cfg.MaxAge.Value(), now: time.Now}
}
func (q *Queue) Enqueue(payload report.Payload) error {
	if err := os.MkdirAll(q.directory, 0o700); err != nil {
		return fmt.Errorf("create report queue: %w", err)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("serialize queued report: %w", err)
	}
	temporary, err := os.CreateTemp(q.directory, ".report-*.tmp")
	if err != nil {
		return fmt.Errorf("create queued report: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(encoded); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write queued report: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync queued report: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close queued report: %w", err)
	}
	path := filepath.Join(q.directory, payload.ReportID+".json")
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("store queued report: %w", err)
	}
	return q.Prune()
}
func (q *Queue) Reports() ([]report.Payload, error) {
	if err := q.Prune(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	entries, err := q.entries()
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	payloads := make([]report.Payload, 0, len(entries))
	for _, entry := range entries {
		content, err := os.ReadFile(entry.path)
		if err != nil {
			return nil, fmt.Errorf("read queued report: %w", err)
		}
		var payload report.Payload
		if err := json.Unmarshal(content, &payload); err != nil {
			_ = os.Remove(entry.path)
			continue
		}
		payloads = append(payloads, payload)
	}
	return payloads, nil
}
func (q *Queue) Delete(reportID string) error {
	if err := os.Remove(filepath.Join(q.directory, reportID+".json")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove delivered report: %w", err)
	}
	return nil
}
func (q *Queue) Prune() error {
	entries, err := q.entries()
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var total int64
	for _, entry := range entries {
		if q.now().Sub(entry.modified) > q.maxAge {
			if err := os.Remove(entry.path); err != nil {
				return fmt.Errorf("prune expired report: %w", err)
			}
			continue
		}
		total += entry.size
	}
	for _, entry := range entries {
		if total <= q.maxBytes {
			break
		}
		if q.now().Sub(entry.modified) > q.maxAge {
			continue
		}
		if err := os.Remove(entry.path); err != nil {
			return fmt.Errorf("prune queued report: %w", err)
		}
		total -= entry.size
	}
	return nil
}

type queueEntry struct {
	path     string
	size     int64
	modified time.Time
}

func (q *Queue) entries() ([]queueEntry, error) {
	entries, err := os.ReadDir(q.directory)
	if err != nil {
		return nil, err
	}
	result := make([]queueEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("inspect queued report: %w", err)
		}
		result = append(result, queueEntry{path: filepath.Join(q.directory, entry.Name()), size: info.Size(), modified: info.ModTime()})
	}
	sort.Slice(result, func(left, right int) bool { return result[left].modified.Before(result[right].modified) })
	return result, nil
}
