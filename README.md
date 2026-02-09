# Go Structured Logs

A Go library for structured JSON logging with Gin framework support. Similar to the Spring Boot version, it provides two main log types: **request** and **application**.

## Features

- 📝 **Structured JSON Logging** - All logs are output in a consistent JSON format
- 🔗 **Correlation ID** - Automatic correlation ID propagation across services
- 📥 **Request Logging** - Automatic logging of incoming HTTP requests and responses (Gin middleware)
- 📤 **Outgoing Request Logging** - Log calls to external services via HTTP client
- 🔒 **Security** - Automatic masking of sensitive headers and fields
- ⚙️ **Configurable** - Extensive configuration options
- 🚀 **Easy to use** - Simple API similar to standard logging

## Installation

```bash
go get github.com/jnp/go-structured-logs
```

## Quick Start

### Middleware Usage (2 แบบ)

#### แบบง่าย (ใช้ default config)

```go
log := logger.New("my-service")
r.Use(middleware.LoggingMiddleware(log))
```

#### แบบกำหนด config เอง (ถ้าต้องการ)

```go
log := logger.New("my-service")
config := &logger.Config{
    ApplicationName: "my-service",
    RequestLogging: logger.RequestLoggingConfig{
        LogBody:         true,
        LogResponseBody: true,
        MaxBodySize:     10000,
    },
}
r.Use(middleware.LoggingMiddleware(log, config))
```

---

### HTTP Client Usage (2 แบบ)

#### แบบง่าย (ใช้ default config)

```go
log := logger.New("MyService")
client := httpclient.NewClient(log, nil)
```

#### แบบกำหนด config เอง (ถ้าต้องการ)

```go
log := logger.New("MyService")
config := &logger.Config{
    ApplicationName: "my-service",
    RequestLogging: logger.RequestLoggingConfig{
        LogBody:         true,
        LogResponseBody: true,
    },
}
client := httpclient.NewClient(log, config)
```

---

### 1. Simple Usage (No Configuration Required!)

```go
package main

import (
    "context"
    "github.com/jnp/go-structured-logs/pkg/logger"
)

func main() {
    log := logger.New("MyService")
    ctx := context.Background()

    log.Info(ctx, "Processing started")
    log.Debug(ctx, "Debug information")
    log.Warn(ctx, "Warning message")
    log.ErrorWithErr(ctx, "Error occurred", err)

    log.Info(ctx, "User action", map[string]interface{}{
        "userId": "12345",
        "action": "LOGIN",
    })
}
```

### 2. With Gin Framework (Full Example)

#### Simple Usage (Recommended)

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/jnp/go-structured-logs/pkg/logger"
    "github.com/jnp/go-structured-logs/pkg/middleware"
)

func main() {
    log := logger.New("main")

    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(middleware.LoggingMiddleware(log))

    r.GET("/api/users", getUsers)
    r.Run(":8080")
}

func getUsers(c *gin.Context) {
    log := logger.New("UserHandler")
    ctx := c.Request.Context()

    log.Info(ctx, "Fetching users")
    
    c.JSON(200, gin.H{"users": []string{}})
}
```

#### With Custom Configuration (Optional)

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/jnp/go-structured-logs/pkg/logger"
    "github.com/jnp/go-structured-logs/pkg/middleware"
)

func main() {
    config := &logger.Config{
        ApplicationName: "my-service",
        RequestLogging: logger.RequestLoggingConfig{
            LogBody:         true,
            LogResponseBody: true,
            MaxBodySize:     10000,
            ExcludePatterns: []string{"/health", "/metrics"},
        },
    }
    
    logger.ConfigureDefaults(config)

    log := logger.New("main")

    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(middleware.LoggingMiddleware(log, config))

    r.GET("/api/users", getUsers)
    r.Run(":8080")
}
```

### 3. HTTP Client with Logging (Full Example)

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/jnp/go-structured-logs/pkg/httpclient"
    "github.com/jnp/go-structured-logs/pkg/logger"
)

func main() {
    log := logger.New("MyService")

    client := httpclient.NewClient(log, nil)

    ctx := context.Background()
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
    resp, err := client.Do(req)
    // ...
}
```

## Configuration

All configuration is **optional**. The library uses sensible defaults.

### Default Values

| Property | Default Value |
|----------|---------------|
| ApplicationName | "application" |
| Enabled | true |
| RequestLogging.LogBody | true |
| RequestLogging.LogResponseBody | true |
| RequestLogging.MaxBodySize | 10240 |

## License

MIT