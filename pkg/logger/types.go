package logger

import "time"

// LogType represents the type of log entry
type LogType string

const (
	LogTypeApplication LogType = "application"
	LogTypeRequest     LogType = "request"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp     string                 `json:"@timestamp"`
	Version       string                 `json:"@version"`
	Application   string                 `json:"application"`
	Message       string                 `json:"message"`
	LoggerName    string                 `json:"logger_name"`
	ThreadName    string                 `json:"thread_name"`
	Level         string                 `json:"level"`
	LevelValue    int                    `json:"level_value"`
	Type          LogType                `json:"type"`
	CorrelationID string                 `json:"correlation_id"`
	Method        string                 `json:"method"`
	URI           string                 `json:"uri"`
	StatusCode    *int                   `json:"status_code,omitempty"`
	DurationMs    *int64                 `json:"duration_ms,omitempty"`
	RemoteAddress string                 `json:"remote_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	RequestBody   interface{}            `json:"request_body"`
	ResponseBody  interface{}            `json:"response_body"`
	Error         *ErrorInfo             `json:"error,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}

// ErrorInfo represents error details in log
type ErrorInfo struct {
	Class      string   `json:"class"`
	Message    string   `json:"message"`
	StackTrace []string `json:"stack_trace,omitempty"`
	RootCause  string   `json:"root_cause,omitempty"`
}

// RequestInfo represents request details
type RequestInfo struct {
	Headers       map[string]string `json:"headers,omitempty"`
	QueryParams   map[string]string `json:"query_params,omitempty"`
	PathParams    map[string]string `json:"path_params,omitempty"`
	Body          interface{}       `json:"body,omitempty"`
	ContentType   string            `json:"content_type,omitempty"`
	ContentLength int64             `json:"content_length,omitempty"`
}

// ResponseInfo represents response details
type ResponseInfo struct {
	Headers       map[string]string `json:"headers,omitempty"`
	Body          interface{}       `json:"body,omitempty"`
	ContentType   string            `json:"content_type,omitempty"`
	ContentLength int64             `json:"content_length,omitempty"`
}

// Config represents logger configuration
type Config struct {
	ApplicationName    string
	Enabled            bool
	RequestLogging     RequestLoggingConfig
	ApplicationLogging ApplicationLoggingConfig
	MaskedHeaders      []string
	MaskedFields       []string
	MaskValue          string
}

// RequestLoggingConfig represents request logging settings
type RequestLoggingConfig struct {
	Enabled         bool
	LogHeaders      bool
	LogBody         bool
	LogResponseBody bool
	MaxBodySize     int
	ExcludePatterns []string
	IncludePatterns []string
}

// ApplicationLoggingConfig represents application logging settings
type ApplicationLoggingConfig struct {
	Enabled            bool
	IncludeStackTrace  bool
	MaxStackTraceDepth int
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		ApplicationName: "application",
		Enabled:         true,
		RequestLogging: RequestLoggingConfig{
			Enabled:         true,
			LogHeaders:      true,
			LogBody:         true,
			LogResponseBody: true,
			MaxBodySize:     10240,
			ExcludePatterns: []string{"/health", "/metrics", "/favicon.ico"},
			IncludePatterns: []string{},
		},
		ApplicationLogging: ApplicationLoggingConfig{
			Enabled:            true,
			IncludeStackTrace:  true,
			MaxStackTraceDepth: 50,
		},
		MaskedHeaders: []string{"Authorization", "X-Api-Key", "Cookie", "Set-Cookie"},
		MaskedFields:  []string{"password", "secret", "token", "creditCard", "ssn"},
		MaskValue:     "***MASKED***",
	}
}

// LevelValue maps log level to numeric value
var LevelValue = map[string]int{
	"TRACE": 5000,
	"DEBUG": 10000,
	"INFO":  20000,
	"WARN":  30000,
	"ERROR": 40000,
}

// FormatTimestamp formats time to ISO 8601 with timezone
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000-07:00")
}
