package report

import "time"

const SchemaVersion = 1

type Payload struct {
	SchemaVersion int            `json:"schema_version"`
	ReportID      string         `json:"report_id"`
	SampledAt     time.Time      `json:"sampled_at"`
	Host          Host           `json:"host"`
	ServiceChecks []ServiceCheck `json:"service_checks,omitempty"`
	Agent         Agent          `json:"agent"`
}

type Host struct {
	CPUUsagePercent  float64 `json:"cpu_usage_percent"`
	LogicalCPUCount  int     `json:"logical_cpu_count"`
	LoadAverage1m    float64 `json:"load_average_1m"`
	LoadAverage5m    float64 `json:"load_average_5m"`
	LoadAverage15m   float64 `json:"load_average_15m"`
	RAMUsagePercent  float64 `json:"ram_usage_percent"`
	SwapUsagePercent float64 `json:"swap_usage_percent"`
	UptimeSeconds    int64   `json:"uptime_seconds"`
}

type ServiceCheck struct {
	ID             string  `json:"id"`
	Success        bool    `json:"success"`
	ResponseTimeMS float64 `json:"response_time_ms"`
	StatusCode     *int    `json:"status_code,omitempty"`
}

type Agent struct {
	Version string `json:"version"`
}
