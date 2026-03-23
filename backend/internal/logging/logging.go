package logging

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level represents log level
type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
)

// Config holds logging configuration
type Config struct {
	Level  Level  `json:"level"`
	Format string `json:"format"` // json, console
}

// Logger wraps zap.Logger with convenience methods
type Logger struct {
	*zap.Logger
	config Config
}

var (
	global     *Logger
	globalOnce sync.Once
)

// New creates a new logger
func New(config Config) *Logger {
	// Parse level
	var level zapcore.Level
	switch config.Level {
	case DebugLevel:
		level = zapcore.DebugLevel
	case InfoLevel:
		level = zapcore.InfoLevel
	case WarnLevel:
		level = zapcore.WarnLevel
	case ErrorLevel:
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Create encoder
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if config.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create core
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		Logger: logger,
		config: config,
	}
}

// Global returns global logger instance
func Global() *Logger {
	globalOnce.Do(func() {
		global = New(Config{
			Level:  InfoLevel,
			Format: "json",
		})
	})
	return global
}

// SetGlobal sets global logger
func SetGlobal(logger *Logger) {
	global = logger
}

// WithContext adds context fields to logger
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	// Extract trace ID, request ID, etc. from context
	// For now, just return base logger
	return l.Logger
}

// Debug logs debug message
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Info logs info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn logs warning message
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error logs error message
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// With creates a new logger with fields
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		config: l.config,
	}
}

// Middleware creates logging middleware for HTTP handlers
func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request
		l.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
		)

		// Wrap response writer
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		// Log response
		l.Info("HTTP response",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", wrapped.statusCode),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Fields provides common field builders
type Fields struct{}

// F returns Fields instance
func F() Fields {
	return Fields{}
}

// String creates string field
func (Fields) String(key, value string) zap.Field {
	return zap.String(key, value)
}

// Int creates int field
func (Fields) Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Error creates error field
func (Fields) Error(err error) zap.Field {
	return zap.Error(err)
}

// Duration creates duration field
func (Fields) Duration(key string, d time.Duration) zap.Field {
	return zap.Duration(key, d)
}

// Bool creates bool field
func (Fields) Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

// Any creates field of any type
func (Fields) Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}
