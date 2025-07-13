package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	apperrors "api/errors"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// HTTPClient provides a client with retry capabilities for outbound requests
type HTTPClient struct {
	client      *http.Client
	retryConfig RetryConfig
}

// NewHTTPClient creates a new HTTP client with timeout and retry configuration
func NewHTTPClient(timeout time.Duration, retryConfig RetryConfig) *HTTPClient {
	// Create base transport with OpenTelemetry instrumentation
	var transport http.RoundTripper = &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// Wrap with OpenTelemetry instrumentation if enabled
	if otelEnabled {
		transport = otelhttp.NewTransport(transport)
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		retryConfig: retryConfig,
	}
}

// DefaultHTTPClient provides a preconfigured HTTP client with sensible defaults
var DefaultHTTPClient = NewHTTPClient(30*time.Second, DefaultRetryConfig)

// Do executes an HTTP request with automatic retry on failures
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	var lastErr error

	// Add tracing context if available
	ctx := req.Context()
	if otelEnabled {
		var span trace.Span
		ctx, span = StartSpanWithAttributes(ctx, "http.client.request", []attribute.KeyValue{
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL.String()),
			attribute.String("http.scheme", req.URL.Scheme),
			attribute.String("http.host", req.Host),
		})
		defer span.End()

		// Update request with traced context
		req = req.WithContext(ctx)
	}

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Clone request for retry (necessary because body might be consumed)
		reqClone := c.cloneRequest(req)

		// Add attempt information to span
		if otelEnabled {
			span := trace.SpanFromContext(ctx)
			span.SetAttributes(
				attribute.Int("http.attempt", attempt+1),
				attribute.Int("http.max_retries", c.retryConfig.MaxRetries+1),
			)
		}

		// Execute request
		resp, err := c.client.Do(reqClone)

		// Record the response in telemetry
		if otelEnabled {
			span := trace.SpanFromContext(ctx)
			if err != nil {
				RecordError(ctx, err, "HTTP request failed")
			} else {
				span.SetAttributes(
					attribute.Int("http.status_code", resp.StatusCode),
					attribute.Int64("http.response_content_length", resp.ContentLength),
				)
			}
		}

		// If no error and status doesn't warrant retry, return response
		if err == nil && !ShouldRetry(resp.StatusCode) {
			// Add baggage values to spans for successful requests
			if otelEnabled {
				EnrichSpanWithBaggage(ctx)
			}
			return resp, nil
		}

		// Store error for potential return
		if err != nil {
			lastErr = err
		} else {
			lastErr = apperrors.WrapErrHTTPStatus(resp.StatusCode, resp.Status)
			resp.Body.Close() // Close body before retry
		}

		// Don't sleep after last attempt
		if attempt < c.retryConfig.MaxRetries {
			delay := c.retryConfig.ExponentialBackoff(attempt)

			// Log retry attempt with trace context
			if otelEnabled {
				AddSpanEvent(ctx, "http.retry",
					attribute.Int("attempt", attempt+1),
					attribute.String("reason", lastErr.Error()),
					attribute.String("delay", delay.String()),
				)
			}

			log.Printf("HTTP request failed (attempt %d/%d), retrying in %v: %v",
				attempt+1, c.retryConfig.MaxRetries+1, delay, lastErr)

			select {
			case <-time.After(delay):
				// Continue to next attempt
			case <-req.Context().Done():
				if otelEnabled {
					RecordError(ctx, req.Context().Err(), "HTTP request cancelled during retry")
				}
				return nil, req.Context().Err()
			}
		}
	}

	// Record final failure
	if otelEnabled {
		finalErr := fmt.Errorf("request failed after %d attempts: %w", c.retryConfig.MaxRetries+1, lastErr)
		RecordError(ctx, finalErr, "HTTP request exhausted all retries")
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryConfig.MaxRetries+1, lastErr)
}

// Get performs a GET request with retry
func (c *HTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post performs a POST request with retry
func (c *HTTPClient) Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

// cloneRequest creates a copy of the HTTP request for retry purposes
func (c *HTTPClient) cloneRequest(req *http.Request) *http.Request {
	// For requests with body, we need to handle it carefully
	// This is a simplified version - in production you might want to buffer the body
	clone := req.Clone(req.Context())
	return clone
}

// WithTimeout creates a new client with different timeout
func (c *HTTPClient) WithTimeout(timeout time.Duration) *HTTPClient {
	newClient := *c
	newClient.client = &http.Client{
		Timeout:   timeout,
		Transport: c.client.Transport,
	}
	return &newClient
}

// WithRetryConfig creates a new client with different retry configuration
func (c *HTTPClient) WithRetryConfig(config RetryConfig) *HTTPClient {
	newClient := *c
	newClient.retryConfig = config
	return &newClient
}

// GetInstrumented creates an instrumented HTTP client for making external API calls
func GetInstrumented() *HTTPClient {
	return NewHTTPClient(30*time.Second, DefaultRetryConfig)
}
