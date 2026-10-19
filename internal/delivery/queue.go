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

const MaxQueueReports = 4096

type QueueStats struct {
	Count     int
	Bytes     int64
	OldestAge time.Duration
	MaxBytes  int64
	MaxCount  int
}

func NewQueue(cfg config.Queue) *Queue {
	return &Queue{directory: cfg.Directory, maxBytes: cfg.MaxBytes, maxAge: cfg.MaxAge.Value(), now: time.Now}
}
func (q *Queue) Enqueue(payload report.Payload) error {
	if !validReportID(payload.ReportID) {
		return fmt.Errorf("queued report id is invalid")
	}
	if err := os.MkdirAll(q.directory, 0o700); err != nil {
		return fmt.Errorf("create report queue: %w", err)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("serialize queued report: %w", err)
	}
	if len(encoded) > report.MaxPayloadBytes {
		return fmt.Errorf("queued report exceeds maximum encoded size")
	}
	if err := q.pruneFor(int64(len(encoded))); err != nil {
		return err
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
	return nil
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
		if !validReportID(payload.ReportID) || filepath.Base(entry.path) != payload.ReportID+".json" {
			_ = os.Remove(entry.path)
			continue
		}
		payloads = append(payloads, payload)
	}
	return payloads, nil
}
func (q *Queue) Delete(reportID string) error {
	if !validReportID(reportID) {
		return fmt.Errorf("queued report id is invalid")
	}
	if err := os.Remove(filepath.Join(q.directory, reportID+".json")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove delivered report: %w", err)
	}
	return nil
}
func (q *Queue) Prune() error {
	return q.pruneFor(0)
}

func (q *Queue) Stats() (QueueStats, error) {
	if err := q.Prune(); err != nil {
		return QueueStats{}, err
	}
	entries, err := q.entries()
	if os.IsNotExist(err) {
		return QueueStats{MaxBytes: q.maxBytes, MaxCount: MaxQueueReports}, nil
	}
	if err != nil {
		return QueueStats{}, err
	}
	stats := QueueStats{Count: len(entries), MaxBytes: q.maxBytes, MaxCount: MaxQueueReports}
	if len(entries) == 0 {
		return stats, nil
	}
	stats.OldestAge = q.now().Sub(entries[0].modified)
	if stats.OldestAge < 0 {
		stats.OldestAge = 0
	}
	for _, entry := range entries {
		stats.Bytes += entry.size
	}
	return stats, nil
}

func (q *Queue) pruneFor(requiredBytes int64) error {
	if requiredBytes < 0 || requiredBytes > q.maxBytes || requiredBytes > report.MaxPayloadBytes {
		return fmt.Errorf("queued report exceeds queue capacity")
	}
	entries, err := q.entries()
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var total int64
	live := make([]queueEntry, 0, len(entries))
	for _, entry := range entries {
		if q.now().Sub(entry.modified) > q.maxAge {
			if err := os.Remove(entry.path); err != nil {
				return fmt.Errorf("prune expired report: %w", err)
			}
			continue
		}
		if entry.size > report.MaxPayloadBytes {
			if err := os.Remove(entry.path); err != nil {
				return fmt.Errorf("prune oversized report: %w", err)
			}
			continue
		}
		total += entry.size
		live = append(live, entry)
	}
	queuedCount := len(live)
	if requiredBytes > 0 {
		queuedCount++
	}
	for _, entry := range live {
		if total+requiredBytes <= q.maxBytes && queuedCount <= MaxQueueReports {
			break
		}
		if err := os.Remove(entry.path); err != nil {
			return fmt.Errorf("prune queued report: %w", err)
		}
		total -= entry.size
		queuedCount--
	}
	if total+requiredBytes > q.maxBytes || queuedCount > MaxQueueReports {
		return fmt.Errorf("report queue capacity is exhausted")
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
		if !validReportID(strings.TrimSuffix(entry.Name(), ".json")) {
			if err := os.Remove(filepath.Join(q.directory, entry.Name())); err != nil {
				return nil, fmt.Errorf("remove invalid queued report: %w", err)
			}
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("inspect queued report: %w", err)
		}
		if !info.Mode().IsRegular() {
			if info.Mode()&os.ModeSymlink != 0 {
				if err := os.Remove(filepath.Join(q.directory, entry.Name())); err != nil {
					return nil, fmt.Errorf("remove invalid queued report: %w", err)
				}
			}
			continue
		}
		result = append(result, queueEntry{path: filepath.Join(q.directory, entry.Name()), size: info.Size(), modified: info.ModTime()})
	}
	sort.Slice(result, func(left, right int) bool { return result[left].modified.Before(result[right].modified) })
	return result, nil
}

func validReportID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || value[14] != '4' {
		return false
	}
	for index, character := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if (index == 19 && character != '8' && character != '9' && character != 'a' && character != 'b') ||
			(index != 19 && !(character >= '0' && character <= '9') && !(character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
