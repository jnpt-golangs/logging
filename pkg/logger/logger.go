package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// AppLogger is the main structured logger
type AppLogger struct {
	name        string
	config      *Config
	masking     *MaskingUtil
	output      io.Writer
	mu          sync.Mutex
}

var (
	defaultConfig *Config
	configMu      sync.RWMutex
)

func init() {
	defaultConfig = DefaultConfig()
}

// ConfigureDefaults sets default configuration for all loggers
func ConfigureDefaults(config *Config) {
	configMu.Lock()
	defer configMu.Unlock()
	if config != nil {
		defaultConfig = config
	}
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return defaultConfig
}

// New creates a new AppLogger with the given name
func New(name string) *AppLogger {
	return NewWithConfig(name, GetDefaultConfig())
}

// NewWithConfig creates a new AppLogger with custom config
func NewWithConfig(name string, config *Config) *AppLogger {
	if config == nil {
		config = DefaultConfig()
	}
	return &AppLogger{
		name:    name,
		config:  config,
		masking: NewMaskingUtil(config),
		output:  os.Stdout,
	}
}

// SetOutput sets the output writer
func (l *AppLogger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// Debug logs a debug message
func (l *AppLogger) Debug(ctx context.Context, message string, extra ...map[string]interface{}) {
	if !l.isApplicationLoggingEnabled() {
		return
	}
	l.logApplication(ctx, "DEBUG", message, nil, mergeExtra(extra))
}

// Info logs an info message
func (l *AppLogger) Info(ctx context.Context, message string, extra ...map[string]interface{}) {
	if !l.isApplicationLoggingEnabled() {
		return
	}
	l.logApplication(ctx, "INFO", message, nil, mergeExtra(extra))
}

// Warn logs a warning message
func (l *AppLogger) Warn(ctx context.Context, message string, extra ...map[string]interface{}) {
	if !l.isApplicationLoggingEnabled() {
		return
	}
	l.logApplication(ctx, "WARN", message, nil, mergeExtra(extra))
}

// WarnError logs a warning message with error
func (l *AppLogger) WarnError(ctx context.Context, message string, err error, extra ...map[string]interface{}) {
	if !l.isApplicationLoggingEnabled() {
		return
	}
	l.logApplication(ctx, "WARN", message, err, mergeExtra(extra))
}

// Error logs an error message
func (l *AppLogger) Error(ctx context.Context, message string, extra ...map[string]interface{}) {
	if !l.isApplicationLoggingEnabled() {
		return
	}
	l.logApplication(ctx, "ERROR", message, nil, mergeExtra(extra))
}

// ErrorWithErr logs an error message with error
func (l *AppLogger) ErrorWithErr(ctx context.Context, message string, err error, extra ...map[string]interface{}) {
	if !l.isApplicationLoggingEnabled() {
		return
	}
	l.logApplication(ctx, "ERROR", message, err, mergeExtra(extra))
}

// LogIncomingRequest logs an incoming HTTP request
func (l *AppLogger) LogIncomingRequest(ctx context.Context, method, uri string, request *RequestInfo, remoteAddr, userAgent string) {
	if !l.isRequestLoggingEnabled() {
		return
	}
	l.logRequest(ctx, "INFO", "Incoming request", method, uri, nil, nil, remoteAddr, userAgent, request, nil, nil)
}

// LogIncomingResponse logs an incoming HTTP response
func (l *AppLogger) LogIncomingResponse(ctx context.Context, method, uri string, statusCode int, durationMs int64, response *ResponseInfo) {
	if !l.isRequestLoggingEnabled() {
		return
	}
	level := l.getStatusLevel(statusCode)
	l.logRequest(ctx, level, "Incoming response", method, uri, &statusCode, &durationMs, "", "", nil, response, nil)
}

// LogOutgoingRequest logs an outgoing HTTP request
func (l *AppLogger) LogOutgoingRequest(ctx context.Context, method, uri string, request *RequestInfo) {
	if !l.isRequestLoggingEnabled() {
		return
	}
	extra := map[string]interface{}{"direction": "outgoing"}
	l.logRequest(ctx, "INFO", "Outgoing request", method, uri, nil, nil, "", "", request, nil, extra)
}

// LogOutgoingResponse logs an outgoing HTTP response
func (l *AppLogger) LogOutgoingResponse(ctx context.Context, method, uri string, statusCode int, durationMs int64, response *ResponseInfo) {
	if !l.isRequestLoggingEnabled() {
		return
	}
	level := l.getStatusLevel(statusCode)
	extra := map[string]interface{}{"direction": "outgoing"}
	l.logRequest(ctx, level, "Outgoing response", method, uri, &statusCode, &durationMs, "", "", nil, response, extra)
}

func (l *AppLogger) logApplication(ctx context.Context, level, message string, err error, extra map[string]interface{}) {
	entry := LogEntry{
		Timestamp:     FormatTimestamp(time.Now()),
		Version:       "1",
		Application:   l.config.ApplicationName,
		Message:       message,
		LoggerName:    l.name,
		ThreadName:    getGoroutineID(),
		Level:         level,
		LevelValue:    LevelValue[level],
		Type:          LogTypeApplication,
		CorrelationID: GetCorrelationID(ctx),
		Method:        "",
		URI:           "",
		RequestBody:   map[string]interface{}{},
		ResponseBody:  map[string]interface{}{},
		Extra:         extra,
	}

	if err != nil && l.config.ApplicationLogging.IncludeStackTrace {
		entry.Error = l.buildErrorInfo(err)
	}

	l.writeLog(entry)
}

func (l *AppLogger) logRequest(ctx context.Context, level, message, method, uri string,
	statusCode *int, durationMs *int64, remoteAddr, userAgent string,
	request *RequestInfo, response *ResponseInfo, extra map[string]interface{}) {

	var reqBody interface{} = map[string]interface{}{}
	var resBody interface{} = map[string]interface{}{}

	if request != nil {
		reqBody = l.buildRequestBody(request)
	}
	if response != nil {
		resBody = l.buildResponseBody(response)
	}

	entry := LogEntry{
		Timestamp:     FormatTimestamp(time.Now()),
		Version:       "1",
		Application:   l.config.ApplicationName,
		Message:       message,
		LoggerName:    l.name,
		ThreadName:    getGoroutineID(),
		Level:         level,
		LevelValue:    LevelValue[level],
		Type:          LogTypeRequest,
		CorrelationID: GetCorrelationID(ctx),
		Method:        method,
		URI:           uri,
		StatusCode:    statusCode,
		DurationMs:    durationMs,
		RemoteAddress: remoteAddr,
		UserAgent:     userAgent,
		RequestBody:   reqBody,
		ResponseBody:  resBody,
		Extra:         extra,
	}

	l.writeLog(entry)
}

func (l *AppLogger) buildRequestBody(req *RequestInfo) map[string]interface{} {
	body := make(map[string]interface{})

	if l.config.RequestLogging.LogHeaders && req.Headers != nil {
		body["headers"] = l.masking.MaskHeaders(req.Headers)
	}
	if req.QueryParams != nil {
		body["query_params"] = req.QueryParams
	}
	if req.PathParams != nil {
		body["path_params"] = req.PathParams
	}
	if l.config.RequestLogging.LogBody && req.Body != nil {
		body["body"] = l.masking.MaskBody(req.Body)
	}
	if req.ContentType != "" {
		body["content_type"] = req.ContentType
	}
	if req.ContentLength > 0 {
		body["content_length"] = req.ContentLength
	}

	return body
}

func (l *AppLogger) buildResponseBody(res *ResponseInfo) map[string]interface{} {
	body := make(map[string]interface{})

	if l.config.RequestLogging.LogHeaders && res.Headers != nil {
		body["headers"] = l.masking.MaskHeaders(res.Headers)
	}
	if l.config.RequestLogging.LogResponseBody && res.Body != nil {
		body["body"] = l.masking.MaskBody(res.Body)
	}
	if res.ContentType != "" {
		body["content_type"] = res.ContentType
	}
	if res.ContentLength > 0 {
		body["content_length"] = res.ContentLength
	}

	return body
}

func (l *AppLogger) buildErrorInfo(err error) *ErrorInfo {
	if err == nil {
		return nil
	}

	info := &ErrorInfo{
		Class:   fmt.Sprintf("%T", err),
		Message: err.Error(),
	}

	// Capture stack trace
	if l.config.ApplicationLogging.IncludeStackTrace {
		info.StackTrace = l.captureStackTrace()
	}

	return info
}

func (l *AppLogger) captureStackTrace() []string {
	var pcs [64]uintptr
	n := runtime.Callers(4, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var trace []string
	maxDepth := l.config.ApplicationLogging.MaxStackTraceDepth
	for i := 0; i < maxDepth; i++ {
		frame, more := frames.Next()
		if !more {
			break
		}
		trace = append(trace, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
	}
	return trace
}

func (l *AppLogger) writeLog(entry LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(l.output, "Failed to marshal log entry: %v\n", err)
		return
	}
	fmt.Fprintln(l.output, string(data))
}

func (l *AppLogger) getStatusLevel(statusCode int) string {
	if statusCode >= 500 {
		return "ERROR"
	}
	if statusCode >= 400 {
		return "WARN"
	}
	return "INFO"
}

func (l *AppLogger) isApplicationLoggingEnabled() bool {
	return l.config.Enabled && l.config.ApplicationLogging.Enabled
}

func (l *AppLogger) isRequestLoggingEnabled() bool {
	return l.config.Enabled && l.config.RequestLogging.Enabled
}

func mergeExtra(extras []map[string]interface{}) map[string]interface{} {
	if len(extras) == 0 {
		return nil
	}
	result := make(map[string]interface{})
	for _, extra := range extras {
		for k, v := range extra {
			result[k] = v
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func getGoroutineID() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	return fmt.Sprintf("goroutine-%s", idField)
}
