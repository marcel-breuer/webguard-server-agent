package checks

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/marcel-breuer/webguard-server-agent/internal/config"
	"github.com/marcel-breuer/webguard-server-agent/internal/report"
)

func Run(ctx context.Context, configured []config.ServiceCheck) []report.ServiceCheck {
	results := make([]report.ServiceCheck, 0, len(configured))
	for _, check := range configured {
		results = append(results, run(ctx, check))
	}
	return results
}
func run(parent context.Context, check config.ServiceCheck) report.ServiceCheck {
	ctx, cancel := context.WithTimeout(parent, check.Timeout.Value())
	defer cancel()
	started := time.Now()
	result := report.ServiceCheck{ID: check.ID}
	switch check.Type {
	case "http":
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, check.Target, nil)
		if err == nil {
			response, err := (&http.Client{}).Do(request)
			if err == nil {
				defer response.Body.Close()
				status := response.StatusCode
				result.StatusCode = &status
				result.Success = status >= http.StatusOK && status < http.StatusMultipleChoices
			}
		}
	case "tcp":
		connection, err := (&net.Dialer{}).DialContext(ctx, "tcp", check.Target)
		if err == nil {
			result.Success = true
			_ = connection.Close()
		}
	}
	result.ResponseTimeMS = float64(time.Since(started).Microseconds()) / 1000
	return result
}
