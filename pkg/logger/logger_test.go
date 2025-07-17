package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	configv1 "sing-box-web/pkg/config/v1"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		config  configv1.LogConfig
		wantErr bool
	}{
		{
			name: "stdout json logger",
			config: configv1.LogConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
			wantErr: false,
		},
		{
			name: "stderr console logger",
			config: configv1.LogConfig{
				Level:  "debug",
				Format: "console",
				Output: "stderr",
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: configv1.LogConfig{
				Level:  "invalid",
				Format: "json",
				Output: "stdout",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("NewLogger() returned nil logger when no error expected")
			}
		})
	}
}

func TestFileLogger(t *testing.T) {
	// Create temporary log file
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	config := configv1.LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     logFile,
		MaxSize:    1,
		MaxAge:     1,
		MaxBackups: 1,
		Compress:   false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create file logger: %v", err)
	}

	// Test logging
	logger.Info("test message", zap.String("key", "value"))
	logger.Sync()

	// Check if file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestGlobalLogger(t *testing.T) {
	// Reset global logger
	globalLogger = nil
	globalSugar = nil

	config := configv1.LogConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}

	err := InitLogger(config)
	if err != nil {
		t.Fatalf("Failed to initialize global logger: %v", err)
	}

	// Test global logger functions
	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger() returned nil")
	}

	sugar := GetSugar()
	if sugar == nil {
		t.Error("GetSugar() returned nil")
	}

	// Test logging functions
	Debug("debug message", zap.String("key", "value"))
	Info("info message", zap.String("key", "value"))
	Warn("warn message", zap.String("key", "value"))
	Error("error message", zap.String("key", "value"))

	Debugf("debug message: %s", "formatted")
	Infof("info message: %s", "formatted")
	Warnf("warn message: %s", "formatted")
	Errorf("error message: %s", "formatted")
}

func TestLoggerWith(t *testing.T) {
	config := configv1.LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test With methods
	withLogger := logger.With(zap.String("service", "test"))
	if withLogger == nil {
		t.Error("With() returned nil logger")
	}

	namedLogger := logger.Named("test-service")
	if namedLogger == nil {
		t.Error("Named() returned nil logger")
	}

	contextLogger := logger.WithContext(map[string]interface{}{
		"request_id": "123",
		"user_id":    "456",
	})
	if contextLogger == nil {
		t.Error("WithContext() returned nil logger")
	}

	// Test logging with context
	contextLogger.Info("test message with context")
}

func TestSugaredLogger(t *testing.T) {
	config := configv1.LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	sugar := logger.Sugar()
	if sugar == nil {
		t.Error("Sugar() returned nil")
	}

	// Test With methods
	withSugar := sugar.With("service", "test")
	if withSugar == nil {
		t.Error("With() returned nil sugared logger")
	}

	namedSugar := sugar.Named("test-service")
	if namedSugar == nil {
		t.Error("Named() returned nil sugared logger")
	}

	// Test logging
	sugar.Info("test message")
	sugar.Infof("test message: %s", "formatted")
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		level    string
		expected zapcore.Level
		wantErr  bool
	}{
		{"debug", zapcore.DebugLevel, false},
		{"info", zapcore.InfoLevel, false},
		{"warn", zapcore.WarnLevel, false},
		{"warning", zapcore.WarnLevel, false},
		{"error", zapcore.ErrorLevel, false},
		{"fatal", zapcore.FatalLevel, false},
		{"invalid", zapcore.InfoLevel, true},
		{"DEBUG", zapcore.DebugLevel, false}, // Test case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			level, err := parseLogLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLogLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && level != tt.expected {
				t.Errorf("parseLogLevel() = %v, want %v", level, tt.expected)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	// Initialize logger
	config := configv1.LogConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}
	err := InitLogger(config)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Test helper functions
	fieldLogger := WithField("key", "value")
	if fieldLogger == nil {
		t.Error("WithField() returned nil")
	}

	fieldsLogger := WithFields(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})
	if fieldsLogger == nil {
		t.Error("WithFields() returned nil")
	}

	errorLogger := WithError(err)
	if errorLogger == nil {
		t.Error("WithError() returned nil")
	}

	requestLogger := WithRequestID("req-123")
	if requestLogger == nil {
		t.Error("WithRequestID() returned nil")
	}

	userLogger := WithUserID("user-456")
	if userLogger == nil {
		t.Error("WithUserID() returned nil")
	}

	nodeLogger := WithNodeID("node-789")
	if nodeLogger == nil {
		t.Error("WithNodeID() returned nil")
	}

	durationLogger := WithDuration(time.Second)
	if durationLogger == nil {
		t.Error("WithDuration() returned nil")
	}

	// Test logging with helpers
	fieldLogger.Info("test message with field")
	fieldsLogger.Info("test message with fields")
	requestLogger.Info("test message with request ID")
}

func TestLogWithLevel(t *testing.T) {
	// Initialize logger
	config := configv1.LogConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	}
	err := InitLogger(config)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Test LogWithLevel
	LogWithLevel(zapcore.InfoLevel, "test message", zap.String("key", "value"))
	LogWithLevel(zapcore.DebugLevel, "debug message", zap.String("key", "value"))
	LogWithLevel(zapcore.WarnLevel, "warn message", zap.String("key", "value"))
	LogWithLevel(zapcore.ErrorLevel, "error message", zap.String("key", "value"))
}

func TestSync(t *testing.T) {
	// Initialize logger
	config := configv1.LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	err := InitLogger(config)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Test sync (should not panic)
	Sync()
}

func BenchmarkLogger(b *testing.B) {
	config := configv1.LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	logger, err := NewLogger(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("benchmark message",
				zap.String("key1", "value1"),
				zap.String("key2", "value2"),
				zap.Int("number", 42),
			)
		}
	})
}

func BenchmarkSugaredLogger(b *testing.B) {
	config := configv1.LogConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}

	logger, err := NewLogger(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	sugar := logger.Sugar()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sugar.Infow("benchmark message",
				"key1", "value1",
				"key2", "value2",
				"number", 42,
			)
		}
	})
}