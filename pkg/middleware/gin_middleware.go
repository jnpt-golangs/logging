package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jnpt-golangs/logging/pkg/logger"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggingMiddleware returns a Gin middleware for structured logging
// If config is nil, it uses the default config
func LoggingMiddleware(log *logger.AppLogger, config ...*logger.Config) gin.HandlerFunc {
	cfg := logger.GetDefaultConfig()
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// Skip excluded patterns
		if shouldSkip(c.Request.URL.Path, cfg) {
			c.Next()
			return
		}

		// Setup correlation ID
		correlationID := c.GetHeader(logger.CorrelationIDHeader)
		if correlationID == "" {
			correlationID = logger.GenerateCorrelationID()
		}
		ctx := logger.WithCorrelationID(c.Request.Context(), correlationID)
		c.Request = c.Request.WithContext(ctx)

		// Set correlation ID in response header
		c.Header(logger.CorrelationIDHeader, correlationID)

		// Capture request body
		var requestBody interface{}
		if cfg.RequestLogging.LogBody && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if len(bodyBytes) > 0 {
				var jsonBody interface{}
				if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
					requestBody = jsonBody
				} else {
					requestBody = truncateBody(string(bodyBytes), cfg.RequestLogging.MaxBodySize)
				}
			}
		}

		// Build request info
		reqInfo := &logger.RequestInfo{
			Headers:       extractHeaders(c.Request.Header, cfg),
			QueryParams:   extractQueryParams(c.Request.URL.Query()),
			PathParams:    extractPathParams(c),
			Body:          requestBody,
			ContentType:   c.ContentType(),
			ContentLength: c.Request.ContentLength,
		}

		// Log incoming request
		log.LogIncomingRequest(ctx, c.Request.Method, c.Request.URL.Path, reqInfo,
			c.ClientIP(), c.GetHeader("User-Agent"))

		// Wrap response writer to capture response body
		rw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = rw

		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime).Milliseconds()

		// Build response info
		var responseBody interface{}
		if cfg.RequestLogging.LogResponseBody && rw.body.Len() > 0 {
			var jsonBody interface{}
			if err := json.Unmarshal(rw.body.Bytes(), &jsonBody); err == nil {
				responseBody = jsonBody
			} else {
				responseBody = truncateBody(rw.body.String(), cfg.RequestLogging.MaxBodySize)
			}
		}

		resInfo := &logger.ResponseInfo{
			Headers:       extractResponseHeaders(c.Writer.Header(), cfg),
			Body:          responseBody,
			ContentType:   c.Writer.Header().Get("Content-Type"),
			ContentLength: int64(rw.body.Len()),
		}

		// Log incoming response
		log.LogIncomingResponse(ctx, c.Request.Method, c.Request.URL.Path,
			c.Writer.Status(), duration, resInfo)
	}
}

func shouldSkip(path string, config *logger.Config) bool {
	for _, pattern := range config.RequestLogging.ExcludePatterns {
		if matchPattern(pattern, path) {
			return true
		}
	}

	if len(config.RequestLogging.IncludePatterns) > 0 {
		for _, pattern := range config.RequestLogging.IncludePatterns {
			if matchPattern(pattern, path) {
				return false
			}
		}
		return true
	}

	return false
}

func matchPattern(pattern, path string) bool {
	// Simple pattern matching (supports * as wildcard)
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return strings.HasPrefix(path, prefix)
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}
	return pattern == path
}

func extractHeaders(headers http.Header, config *logger.Config) map[string]string {
	if !config.RequestLogging.LogHeaders {
		return nil
	}
	result := make(map[string]string)
	for key, values := range headers {
		result[key] = strings.Join(values, ", ")
	}
	return result
}

func extractResponseHeaders(headers http.Header, config *logger.Config) map[string]string {
	if !config.RequestLogging.LogHeaders {
		return nil
	}
	result := make(map[string]string)
	for key, values := range headers {
		result[key] = strings.Join(values, ", ")
	}
	return result
}

func extractQueryParams(query map[string][]string) map[string]string {
	if len(query) == 0 {
		return nil
	}
	result := make(map[string]string)
	for key, values := range query {
		result[key] = strings.Join(values, ", ")
	}
	return result
}

func extractPathParams(c *gin.Context) map[string]string {
	params := c.Params
	if len(params) == 0 {
		return nil
	}
	result := make(map[string]string)
	for _, p := range params {
		result[p.Key] = p.Value
	}
	return result
}

func truncateBody(body string, maxSize int) string {
	if len(body) > maxSize {
		return body[:maxSize] + "... [TRUNCATED]"
	}
	return body
}
