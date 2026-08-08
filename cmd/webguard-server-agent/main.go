package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marcel-breuer/webguard-server-agent/internal/checks"
	"github.com/marcel-breuer/webguard-server-agent/internal/config"
	"github.com/marcel-breuer/webguard-server-agent/internal/delivery"
	"github.com/marcel-breuer/webguard-server-agent/internal/report"
	"github.com/marcel-breuer/webguard-server-agent/internal/telemetry"
)

var version = "dev"

const defaultConfigPath = "/etc/webguard-server-agent/config.json"

func main() {
	configPath := flag.String("config", defaultConfigPath, "path to the JSON configuration file")
	once := flag.Bool("once", false, "collect and report once")
	status := flag.Bool("status", false, "validate configuration and show queued report count")
	health := flag.Bool("health", false, "validate configuration and report local readiness")
	showVersion := flag.Bool("version", false, "show the agent version")
	flag.Parse()
	if *showVersion || (flag.NArg() == 1 && flag.Arg(0) == "version") {
		fmt.Println(version)
		return
	}
	if err := run(*configPath, *once, *status || *health); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
func run(configPath string, once, status bool) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	queue := delivery.NewQueue(cfg.Queue)
	if status {
		reports, err := queue.Reports()
		if err != nil {
			return err
		}
		fmt.Printf("configured; %d report(s) queued\n", len(reports))
		return nil
	}
	client, err := delivery.NewClient(cfg)
	if err != nil {
		return err
	}
	collector := telemetry.NewCollector()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	reportOnce := func() {
		if err := flushQueue(ctx, client, queue); err != nil {
			log.Printf("queued reports remain pending: %v", err)
		}
		if err := collectAndDeliver(ctx, collector, client, queue, cfg.ServiceChecks); err != nil {
			log.Printf("report remains pending: %v", err)
		}
	}
	reportOnce()
	if once {
		return nil
	}
	ticker := time.NewTicker(cfg.Interval.Value())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			reportOnce()
		}
	}
}
func flushQueue(ctx context.Context, client *delivery.Client, queue *delivery.Queue) error {
	payloads, err := queue.Reports()
	if err != nil {
		return err
	}
	for _, payload := range payloads {
		err := client.Deliver(ctx, payload)
		if err == nil || !delivery.IsRetryable(err) {
			if deleteErr := queue.Delete(payload.ReportID); deleteErr != nil {
				return deleteErr
			}
			if err != nil {
				log.Printf("discarded rejected queued report: %v", err)
			}
			continue
		}
		return err
	}
	return nil
}
func collectAndDeliver(ctx context.Context, collector *telemetry.Collector, client *delivery.Client, queue *delivery.Queue, configuredChecks []config.ServiceCheck) error {
	host, err := collector.Collect()
	if err != nil {
		return err
	}
	reportID, err := newUUID()
	if err != nil {
		return err
	}
	payload := report.Payload{SchemaVersion: report.SchemaVersion, ReportID: reportID, SampledAt: time.Now().UTC(), Host: host, ServiceChecks: checks.Run(ctx, configuredChecks), Agent: report.Agent{Version: version}}
	if err := client.Deliver(ctx, payload); err != nil {
		if delivery.IsRetryable(err) {
			if queueErr := queue.Enqueue(payload); queueErr != nil {
				return queueErr
			}
		}
		return err
	}
	return nil
}
func newUUID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate report id: %w", err)
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}
