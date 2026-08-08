package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DefaultInterval = 60 * time.Second

type Config struct {
	ReportURL     string         `json:"report_url"`
	Interval      Duration       `json:"interval"`
	ServiceChecks []ServiceCheck `json:"service_checks"`
	Retry         Retry          `json:"retry"`
	Queue         Queue          `json:"queue"`
	ProxyURL      string         `json:"proxy_url,omitempty"`
	CAFile        string         `json:"ca_file,omitempty"`
}
type Duration time.Duration

func (d *Duration) UnmarshalJSON(value []byte) error {
	var text string
	if err := json.Unmarshal(value, &text); err != nil {
		return fmt.Errorf("duration must be a string: %w", err)
	}
	parsed, err := time.ParseDuration(text)
	if err != nil {
		return fmt.Errorf("parse duration: %w", err)
	}
	*d = Duration(parsed)
	return nil
}
func (d Duration) Value() time.Duration { return time.Duration(d) }

type ServiceCheck struct {
	ID      string   `json:"id"`
	Type    string   `json:"type"`
	Target  string   `json:"target"`
	Timeout Duration `json:"timeout"`
}
type Retry struct {
	MaxAttempts int      `json:"max_attempts"`
	BaseDelay   Duration `json:"base_delay"`
	MaxDelay    Duration `json:"max_delay"`
}
type Queue struct {
	Directory string   `json:"directory"`
	MaxBytes  int64    `json:"max_bytes"`
	MaxAge    Duration `json:"max_age"`
}

func Load(path string) (Config, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Config{}, fmt.Errorf("read configuration metadata: %w", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return Config{}, fmt.Errorf("configuration permissions must be owner-only")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read configuration: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse configuration: %w", err)
	}
	cfg.applyDefaults(filepath.Dir(path))
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
func (c *Config) applyDefaults(configDirectory string) {
	if c.Interval == 0 {
		c.Interval = Duration(DefaultInterval)
	}
	if c.Retry.MaxAttempts == 0 {
		c.Retry.MaxAttempts = 3
	}
	if c.Retry.BaseDelay == 0 {
		c.Retry.BaseDelay = Duration(time.Second)
	}
	if c.Retry.MaxDelay == 0 {
		c.Retry.MaxDelay = Duration(30 * time.Second)
	}
	if c.Queue.Directory == "" {
		c.Queue.Directory = filepath.Join(configDirectory, "queue")
	}
	if c.Queue.MaxBytes == 0 {
		c.Queue.MaxBytes = 16 * 1024 * 1024
	}
	if c.Queue.MaxAge == 0 {
		c.Queue.MaxAge = Duration(24 * time.Hour)
	}
	for i := range c.ServiceChecks {
		if c.ServiceChecks[i].Timeout == 0 {
			c.ServiceChecks[i].Timeout = Duration(5 * time.Second)
		}
	}
}
func (c Config) Validate() error {
	reportURL, err := url.ParseRequestURI(c.ReportURL)
	if err != nil || reportURL.Scheme != "https" || reportURL.Host == "" {
		return fmt.Errorf("report_url must be an HTTPS URL")
	}
	if c.Interval.Value() < 10*time.Second {
		return fmt.Errorf("interval must be at least 10s")
	}
	if c.Retry.MaxAttempts < 1 || c.Retry.MaxAttempts > 10 || c.Retry.BaseDelay.Value() <= 0 || c.Retry.MaxDelay.Value() < c.Retry.BaseDelay.Value() {
		return fmt.Errorf("retry settings are invalid")
	}
	if c.Queue.Directory == "" || c.Queue.MaxBytes < 1024 || c.Queue.MaxAge.Value() < time.Minute {
		return fmt.Errorf("queue settings are invalid")
	}
	if c.ProxyURL != "" {
		proxy, err := url.ParseRequestURI(c.ProxyURL)
		if err != nil || (proxy.Scheme != "http" && proxy.Scheme != "https") || proxy.Host == "" {
			return fmt.Errorf("proxy_url must be an HTTP(S) URL")
		}
	}
	if len(c.ServiceChecks) > 20 {
		return fmt.Errorf("at most 20 service checks are supported")
	}
	ids := make(map[string]struct{}, len(c.ServiceChecks))
	for _, check := range c.ServiceChecks {
		if check.ID == "" || len(check.ID) > 100 {
			return fmt.Errorf("service check id is invalid")
		}
		if _, exists := ids[check.ID]; exists {
			return fmt.Errorf("service check ids must be unique")
		}
		ids[check.ID] = struct{}{}
		if check.Timeout.Value() <= 0 || check.Timeout.Value() > 30*time.Second {
			return fmt.Errorf("service check timeout is invalid")
		}
		if err := validateLocalTarget(check.Type, check.Target); err != nil {
			return fmt.Errorf("service check %q: %w", check.ID, err)
		}
	}
	return nil
}
func validateLocalTarget(kind, target string) error {
	switch kind {
	case "http":
		parsed, err := url.ParseRequestURI(target)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || !isLoopbackHost(parsed.Hostname()) {
			return fmt.Errorf("HTTP target must use a loopback HTTP(S) URL")
		}
	case "tcp":
		host, port, err := net.SplitHostPort(target)
		if err != nil || port == "" || !isLoopbackHost(host) {
			return fmt.Errorf("TCP target must use a loopback host and port")
		}
	default:
		return fmt.Errorf("type must be http or tcp")
	}
	return nil
}
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}
