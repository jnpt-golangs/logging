package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/jnpt-golangs/logging/pkg/logger"
)

// LoggingTransport wraps http.RoundTripper with logging
type LoggingTransport struct {
	Transport http.RoundTripper
	Logger    *logger.AppLogger
	Config    *logger.Config
}

// NewLoggingTransport creates a new LoggingTransport
func NewLoggingTransport(log *logger.AppLogger, config *logger.Config) *LoggingTransport {
	return &LoggingTransport{
		Transport: http.DefaultTransport,
		Logger:    log,
		Config:    config,
	}
}

// RoundTrip implements http.RoundTripper
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// Ensure correlation ID
	ctx, correlationID := logger.EnsureCorrelationID(ctx)
	req = req.WithContext(ctx)
	req.Header.Set(logger.CorrelationIDHeader, correlationID)

	// Capture request body
	var requestBody interface{}
	if t.Config.RequestLogging.LogBody && req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > 0 {
				var jsonBody interface{}
				if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
					requestBody = jsonBody
				} else {
					requestBody = truncateBody(string(bodyBytes), t.Config.RequestLogging.MaxBodySize)
				}
			}
		}
	}

	// Build request info
	reqInfo := &logger.RequestInfo{
		Headers:     extractHeaders(req.Header, t.Config),
		Body:        requestBody,
		ContentType: req.Header.Get("Content-Type"),
	}
	if req.ContentLength > 0 {
		reqInfo.ContentLength = req.ContentLength
	}

	// Execute request
	startTime := time.Now()
	resp, err := t.transport().RoundTrip(req)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		t.Logger.ErrorWithErr(ctx, "Outgoing request failed", err, map[string]interface{}{
			"method":      req.Method,
			"url":         req.URL.String(),
			"duration_ms": duration,
		})
		return nil, err
	}

	// Capture response body
	var responseBody interface{}
	if t.Config.RequestLogging.LogResponseBody && resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > 0 {
				var jsonBody interface{}
				if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
					responseBody = jsonBody
				} else {
					responseBody = truncateBody(string(bodyBytes), t.Config.RequestLogging.MaxBodySize)
				}
			}
		}
	}

	// Build response info
	resInfo := &logger.ResponseInfo{
		Headers:       extractResponseHeaders(resp.Header, t.Config),
		Body:          responseBody,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
	}

	// Log request + response as single entry
	t.Logger.LogOutgoingHTTP(ctx, req.Method, req.URL.String(), resp.StatusCode, duration, reqInfo, resInfo)

	return resp, nil
}

func (t *LoggingTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

func extractHeaders(headers http.Header, config *logger.Config) map[string]string {
	if !config.RequestLogging.LogHeaders {
		return nil
	}
	result := make(map[string]string)
	for key, values := range headers {
		result[key] = values[0]
		if len(values) > 1 {
			for i := 1; i < len(values); i++ {
				result[key] += ", " + values[i]
			}
		}
	}
	return result
}

func extractResponseHeaders(headers http.Header, config *logger.Config) map[string]string {
	return extractHeaders(headers, config)
}

func truncateBody(body string, maxSize int) string {
	if len(body) > maxSize {
		return body[:maxSize] + "... [TRUNCATED]"
	}
	return body
}

// NewClient creates an http.Client with logging transport
func NewClient(log *logger.AppLogger, config *logger.Config) *http.Client {
	return &http.Client{
		Transport: NewLoggingTransport(log, config),
	}
}
