package telemetry

import (
	"fmt"
	"testing"
)

func TestCollectorCollectsOnlyDocumentedHostMetrics(t *testing.T) {
	files := map[string]string{procStat: "cpu  100 0 100 800 0 0 0 0 0 0\n", procLoadavg: "1.42 1.10 0.94 1/100 1\n", procMeminfo: "MemTotal: 1000 kB\nMemAvailable: 320 kB\nSwapTotal: 200 kB\nSwapFree: 150 kB\n", procUptime: "86400.9 1.0\n"}
	collector := NewCollector()
	collector.readFile = func(path string) ([]byte, error) {
		value, ok := files[path]
		if !ok {
			return nil, fmt.Errorf("unexpected path %q", path)
		}
		return []byte(value), nil
	}
	host, err := collector.Collect()
	if err != nil {
		t.Fatal(err)
	}
	if host.LogicalCPUCount < 1 || host.LoadAverage1m != 1.42 || host.RAMUsagePercent != 68 || host.SwapUsagePercent != 25 || host.UptimeSeconds != 86400 {
		t.Fatalf("unexpected host telemetry: %+v", host)
	}
}
func TestParseCPURejectsInvalidInput(t *testing.T) {
	if _, err := parseCPU([]byte("invalid\n")); err == nil {
		t.Fatal("expected invalid CPU input to fail")
	}
}
