package telemetry

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/marcel-breuer/webguard-server-agent/internal/report"
)

const (
	procStat    = "/proc/stat"
	procLoadavg = "/proc/loadavg"
	procMeminfo = "/proc/meminfo"
	procUptime  = "/proc/uptime"
)

type Collector struct {
	readFile func(string) ([]byte, error)
	previous *cpuSample
}
type cpuSample struct {
	total uint64
	idle  uint64
}

func NewCollector() *Collector { return &Collector{readFile: os.ReadFile} }

func (c *Collector) Collect() (report.Host, error) {
	stat, err := c.readFile(procStat)
	if err != nil {
		return report.Host{}, fmt.Errorf("read cpu statistics: %w", err)
	}
	loadavg, err := c.readFile(procLoadavg)
	if err != nil {
		return report.Host{}, fmt.Errorf("read load average: %w", err)
	}
	meminfo, err := c.readFile(procMeminfo)
	if err != nil {
		return report.Host{}, fmt.Errorf("read memory statistics: %w", err)
	}
	uptime, err := c.readFile(procUptime)
	if err != nil {
		return report.Host{}, fmt.Errorf("read uptime: %w", err)
	}
	cpu, err := parseCPU(stat)
	if err != nil {
		return report.Host{}, err
	}
	loads, err := parseLoadAverage(loadavg)
	if err != nil {
		return report.Host{}, err
	}
	ram, swap, err := parseMemory(meminfo)
	if err != nil {
		return report.Host{}, err
	}
	uptimeSeconds, err := parseUptime(uptime)
	if err != nil {
		return report.Host{}, err
	}
	usage := 0.0
	if c.previous != nil && cpu.total > c.previous.total {
		totalDelta := cpu.total - c.previous.total
		idleDelta := cpu.idle - c.previous.idle
		usage = 100 * float64(totalDelta-idleDelta) / float64(totalDelta)
	}
	c.previous = &cpu
	return report.Host{CPUUsagePercent: usage, LogicalCPUCount: runtime.NumCPU(), LoadAverage1m: loads[0], LoadAverage5m: loads[1], LoadAverage15m: loads[2], RAMUsagePercent: ram, SwapUsagePercent: swap, UptimeSeconds: uptimeSeconds}, nil
}
func parseCPU(content []byte) (cpuSample, error) {
	fields := strings.Fields(strings.SplitN(string(content), "\n", 2)[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuSample{}, fmt.Errorf("parse cpu statistics")
	}
	var sample cpuSample
	for index, field := range fields[1:] {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return cpuSample{}, fmt.Errorf("parse cpu statistics: %w", err)
		}
		sample.total += value
		if index == 3 || index == 4 {
			sample.idle += value
		}
	}
	return sample, nil
}
func parseLoadAverage(content []byte) ([3]float64, error) {
	fields := strings.Fields(string(content))
	if len(fields) < 3 {
		return [3]float64{}, fmt.Errorf("parse load average")
	}
	var loads [3]float64
	for i := range loads {
		v, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return [3]float64{}, fmt.Errorf("parse load average: %w", err)
		}
		loads[i] = v
	}
	return loads, nil
}
func parseMemory(content []byte) (float64, float64, error) {
	values := map[string]uint64{}
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse memory statistics: %w", err)
		}
		values[strings.TrimSuffix(fields[0], ":")] = value
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("scan memory statistics: %w", err)
	}
	if values["MemTotal"] == 0 {
		return 0, 0, fmt.Errorf("parse memory statistics: missing MemTotal")
	}
	ram := 100 * float64(values["MemTotal"]-values["MemAvailable"]) / float64(values["MemTotal"])
	swap := 0.0
	if values["SwapTotal"] > 0 {
		swap = 100 * float64(values["SwapTotal"]-values["SwapFree"]) / float64(values["SwapTotal"])
	}
	return ram, swap, nil
}
func parseUptime(content []byte) (int64, error) {
	fields := strings.Fields(string(content))
	if len(fields) == 0 {
		return 0, fmt.Errorf("parse uptime")
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse uptime: %w", err)
	}
	return int64(seconds), nil
}
