package delivery

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/marcel-breuer/webguard-server-agent/internal/config"
	"github.com/marcel-breuer/webguard-server-agent/internal/report"
)

type Client struct {
	httpClient *http.Client
	reportURL  string
	retry      config.Retry
	sleep      func(time.Duration)
}

func NewClient(cfg config.Config) (*Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.ProxyURL != "" {
		proxy, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("parse proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	if cfg.CAFile != "" {
		certificate, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read CA file: %w", err)
		}
		pool, _ := x509.SystemCertPool()
		if pool == nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(certificate) {
			return nil, fmt.Errorf("parse CA file")
		}
		transport.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	}
	return &Client{httpClient: &http.Client{Transport: transport, Timeout: 15 * time.Second}, reportURL: cfg.ReportURL, retry: cfg.Retry, sleep: time.Sleep}, nil
}
func (c *Client) Deliver(ctx context.Context, payload report.Payload) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("serialize report: %w", err)
	}
	var deliveryErr error
	for attempt := 1; attempt <= c.retry.MaxAttempts; attempt++ {
		deliveryErr = c.send(ctx, encoded)
		if deliveryErr == nil || !IsRetryable(deliveryErr) {
			return deliveryErr
		}
		if attempt < c.retry.MaxAttempts {
			c.sleep(backoff(c.retry, attempt))
		}
	}
	return deliveryErr
}
func (c *Client) send(ctx context.Context, encoded []byte) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.reportURL, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("create report request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return RetryableError{Err: fmt.Errorf("send report: transport failure")}
	}
	defer response.Body.Close()
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}
	if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500 {
		return RetryableError{Err: fmt.Errorf("report endpoint returned %d", response.StatusCode)}
	}
	return fmt.Errorf("report endpoint rejected report with %d", response.StatusCode)
}

type RetryableError struct{ Err error }

func (e RetryableError) Error() string { return e.Err.Error() }
func (e RetryableError) Unwrap() error { return e.Err }
func IsRetryable(err error) bool       { var retryable RetryableError; return errors.As(err, &retryable) }
func backoff(retry config.Retry, attempt int) time.Duration {
	delay := retry.BaseDelay.Value() << (attempt - 1)
	if delay > retry.MaxDelay.Value() {
		delay = retry.MaxDelay.Value()
	}
	return delay + time.Duration(rand.Int64N(int64(delay)/2+1))
}
