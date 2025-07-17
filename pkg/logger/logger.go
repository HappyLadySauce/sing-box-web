package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	configv1 "sing-box-web/pkg/config/v1"
)

var (
	// Global logger instance
	globalLogger *zap.Logger
	// Global sugared logger instance
	globalSugar *zap.SugaredLogger
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
	config configv1.LogConfig
}

// SugaredLogger wraps zap.SugaredLogger with additional functionality
type SugaredLogger struct {
	*zap.SugaredLogger
	config configv1.LogConfig
}

// InitLogger initializes the global logger with the given configuration
func InitLogger(config configv1.LogConfig) error {
	logger, err := NewLogger(config)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	
	globalLogger = logger.Logger
	globalSugar = logger.Logger.Sugar()
	
	return nil
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config configv1.LogConfig) (*Logger, error) {
	// Build encoder config
	encoderConfig := buildEncoderConfig()
	
	// Build encoder
	var encoder zapcore.Encoder
	switch strings.ToLower(config.Format) {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "text", "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}
	
	// Build writer syncer
	writeSyncer, err := buildWriteSyncer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build write syncer: %w", err)
	}
	
	// Parse log level
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}
	
	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)
	
	// Create logger with options
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}
	
	logger := zap.New(core, opts...)
	
	return &Logger{
		Logger: logger,
		config: config,
	}, nil
}

// buildEncoderConfig builds the encoder configuration
func buildEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// buildWriteSyncer builds the write syncer based on configuration
func buildWriteSyncer(config configv1.LogConfig) (zapcore.WriteSyncer, error) {
	switch strings.ToLower(config.Output) {
	case "stdout":
		return zapcore.AddSync(os.Stdout), nil
	case "stderr":
		return zapcore.AddSync(os.Stderr), nil
	default:
		// File output with rotation
		if !filepath.IsAbs(config.Output) {
			return nil, fmt.Errorf("log file path must be absolute: %s", config.Output)
		}
		
		// Create directory if it doesn't exist
		dir := filepath.Dir(config.Output)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		
		// Configure log rotation
		rotator := &lumberjack.Logger{
			Filename:   config.Output,
			MaxSize:    config.MaxSize,    // MB
			MaxAge:     config.MaxAge,     // days
			MaxBackups: config.MaxBackups, // files
			Compress:   config.Compress,
			LocalTime:  true,
		}
		
		return zapcore.AddSync(rotator), nil
	}
}

// parseLogLevel parses string log level to zapcore.Level
func parseLogLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// Sugar returns a sugared logger
func (l *Logger) Sugar() *SugaredLogger {
	return &SugaredLogger{
		SugaredLogger: l.Logger.Sugar(),
		config:        l.config,
	}
}

// With adds fields to the logger
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		config: l.config,
	}
}

// Named creates a named logger
func (l *Logger) Named(name string) *Logger {
	return &Logger{
		Logger: l.Logger.Named(name),
		config: l.config,
	}
}

// WithContext adds context fields to the logger
func (l *Logger) WithContext(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return l.With(zapFields...)
}

// With adds fields to the sugared logger
func (s *SugaredLogger) With(args ...interface{}) *SugaredLogger {
	return &SugaredLogger{
		SugaredLogger: s.SugaredLogger.With(args...),
		config:        s.config,
	}
}

// Named creates a named sugared logger
func (s *SugaredLogger) Named(name string) *SugaredLogger {
	return &SugaredLogger{
		SugaredLogger: s.SugaredLogger.Named(name),
		config:        s.config,
	}
}

// Global logger functions

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		// Initialize with default config if not initialized
		config := configv1.LogConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		}
		if err := InitLogger(config); err != nil {
			panic(fmt.Sprintf("failed to initialize default logger: %v", err))
		}
	}
	return globalLogger
}

// GetSugar returns the global sugared logger instance
func GetSugar() *zap.SugaredLogger {
	if globalSugar == nil {
		// Initialize with default config if not initialized
		config := configv1.LogConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		}
		if err := InitLogger(config); err != nil {
			panic(fmt.Sprintf("failed to initialize default logger: %v", err))
		}
	}
	return globalSugar
}

// Sync flushes any buffered log entries
func Sync() {
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
	if globalSugar != nil {
		_ = globalSugar.Sync()
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Debugf logs a debug message with formatting
func Debugf(template string, args ...interface{}) {
	GetSugar().Debugf(template, args...)
}

// Infof logs an info message with formatting
func Infof(template string, args ...interface{}) {
	GetSugar().Infof(template, args...)
}

// Warnf logs a warning message with formatting
func Warnf(template string, args ...interface{}) {
	GetSugar().Warnf(template, args...)
}

// Errorf logs an error message with formatting
func Errorf(template string, args ...interface{}) {
	GetSugar().Errorf(template, args...)
}

// Fatalf logs a fatal message with formatting and exits
func Fatalf(template string, args ...interface{}) {
	GetSugar().Fatalf(template, args...)
}

// WithField returns a logger with a single field
func WithField(key string, value interface{}) *zap.Logger {
	return GetLogger().With(zap.Any(key, value))
}

// WithFields returns a logger with multiple fields
func WithFields(fields map[string]interface{}) *zap.Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return GetLogger().With(zapFields...)
}

// WithError returns a logger with an error field
func WithError(err error) *zap.Logger {
	return GetLogger().With(zap.Error(err))
}

// WithRequestID returns a logger with a request ID field
func WithRequestID(requestID string) *zap.Logger {
	return GetLogger().With(zap.String("request_id", requestID))
}

// WithUserID returns a logger with a user ID field
func WithUserID(userID string) *zap.Logger {
	return GetLogger().With(zap.String("user_id", userID))
}

// WithNodeID returns a logger with a node ID field
func WithNodeID(nodeID string) *zap.Logger {
	return GetLogger().With(zap.String("node_id", nodeID))
}

// WithDuration returns a logger with a duration field
func WithDuration(duration time.Duration) *zap.Logger {
	return GetLogger().With(zap.Duration("duration", duration))
}

// LogWithLevel logs a message at the specified level
func LogWithLevel(level zapcore.Level, msg string, fields ...zap.Field) {
	if ce := GetLogger().Check(level, msg); ce != nil {
		ce.Write(fields...)
	}
}

// SetLevel dynamically changes the log level
func SetLevel(level string) error {
	_, err := parseLogLevel(level)
	if err != nil {
		return err
	}
	
	// This requires rebuilding the logger to change the level
	// For simplicity, we'll return an error indicating restart is needed
	return fmt.Errorf("dynamic level change requires logger restart")
}

// GetCurrentLevel returns the current log level as a string
func GetCurrentLevel() string {
	// This would require storing the current level
	// For now, return a placeholder
	return "info"
}