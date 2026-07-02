package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Result struct {
	HTTPStatus int
	Body       string
}

// Client is a shared HTTP client with connection pooling optimised for high-throughput delivery.
type Client struct {
	http *http.Client
}

func New(timeout time.Duration) *Client {
	return &Client{
		http: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  true,
			},
		},
	}
}

// Post sends a JSON payload to the target URL and returns the HTTP status and body.
func (c *Client) Post(ctx context.Context, url, payload string) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url,
		bytes.NewBufferString(payload))
	if err != nil {
		return Result{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	return Result{HTTPStatus: resp.StatusCode, Body: string(body)}, nil
}
