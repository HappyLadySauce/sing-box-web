// Package logger provides logging functionality for sing-box-web.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 日志接口
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

// Config 日志配置
type Config struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
	Output string `yaml:"output" json:"output"`
	File   struct {
		Path       string `yaml:"path" json:"path"`
		MaxSize    int    `yaml:"max_size" json:"max_size"`
		MaxBackups int    `yaml:"max_backups" json:"max_backups"`
		MaxAge     int    `yaml:"max_age" json:"max_age"`
		Compress   bool   `yaml:"compress" json:"compress"`
	} `yaml:"file" json:"file"`
}

// Entry 日志条目包装器
type Entry struct {
	entry *logrus.Entry
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
}

// NewLogger 创建新的日志器
func NewLogger(config *Config) (Logger, error) {
	log := logrus.New()
	
	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %s", config.Level)
	}
	log.SetLevel(level)
	
	// 设置日志格式
	switch strings.ToLower(config.Format) {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	default:
		return nil, fmt.Errorf("unsupported log format: %s", config.Format)
	}
	
	// 设置输出目标
	writer, err := getWriter(config)
	if err != nil {
		return nil, fmt.Errorf("failed to setup log output: %w", err)
	}
	log.SetOutput(writer)
	
	// 设置报告调用者信息
	log.SetReportCaller(true)
	
	return &Entry{entry: logrus.NewEntry(log)}, nil
}

// getWriter 根据配置获取输出写入器
func getWriter(config *Config) (io.Writer, error) {
	switch strings.ToLower(config.Output) {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	case "file":
		if config.File.Path == "" {
			return nil, fmt.Errorf("file path is required when output is set to file")
		}
		
		// 确保日志目录存在
		if err := os.MkdirAll(filepath.Dir(config.File.Path), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		
		// 配置文件轮转
		return &lumberjack.Logger{
			Filename:   config.File.Path,
			MaxSize:    config.File.MaxSize,    // MB
			MaxBackups: config.File.MaxBackups,
			MaxAge:     config.File.MaxAge,     // days
			Compress:   config.File.Compress,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported output type: %s", config.Output)
	}
}

// Debug 输出调试日志
func (e *Entry) Debug(args ...interface{}) {
	e.entry.Debug(args...)
}

// Debugf 输出格式化调试日志
func (e *Entry) Debugf(format string, args ...interface{}) {
	e.entry.Debugf(format, args...)
}

// Info 输出信息日志
func (e *Entry) Info(args ...interface{}) {
	e.entry.Info(args...)
}

// Infof 输出格式化信息日志
func (e *Entry) Infof(format string, args ...interface{}) {
	e.entry.Infof(format, args...)
}

// Warn 输出警告日志
func (e *Entry) Warn(args ...interface{}) {
	e.entry.Warn(args...)
}

// Warnf 输出格式化警告日志
func (e *Entry) Warnf(format string, args ...interface{}) {
	e.entry.Warnf(format, args...)
}

// Error 输出错误日志
func (e *Entry) Error(args ...interface{}) {
	e.entry.Error(args...)
}

// Errorf 输出格式化错误日志
func (e *Entry) Errorf(format string, args ...interface{}) {
	e.entry.Errorf(format, args...)
}

// Fatal 输出致命错误日志并退出程序
func (e *Entry) Fatal(args ...interface{}) {
	e.entry.Fatal(args...)
}

// Fatalf 输出格式化致命错误日志并退出程序
func (e *Entry) Fatalf(format string, args ...interface{}) {
	e.entry.Fatalf(format, args...)
}

// WithField 添加字段
func (e *Entry) WithField(key string, value interface{}) Logger {
	return &Entry{entry: e.entry.WithField(key, value)}
}

// WithFields 添加多个字段
func (e *Entry) WithFields(fields map[string]interface{}) Logger {
	return &Entry{entry: e.entry.WithFields(logrus.Fields(fields))}
}

// WithError 添加错误字段
func (e *Entry) WithError(err error) Logger {
	return &Entry{entry: e.entry.WithError(err)}
}

// 全局默认日志器
var defaultLogger Logger

// InitDefaultLogger 初始化默认日志器
func InitDefaultLogger(config *Config) error {
	if config == nil {
		config = DefaultConfig()
	}
	
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	
	defaultLogger = logger
	return nil
}

// GetDefaultLogger 获取默认日志器
func GetDefaultLogger() Logger {
	if defaultLogger == nil {
		// 如果没有初始化，使用默认配置
		_ = InitDefaultLogger(DefaultConfig())
	}
	return defaultLogger
}

// 便捷函数，使用默认日志器

// Debug 输出调试日志
func Debug(args ...interface{}) {
	GetDefaultLogger().Debug(args...)
}

// Debugf 输出格式化调试日志
func Debugf(format string, args ...interface{}) {
	GetDefaultLogger().Debugf(format, args...)
}

// Info 输出信息日志
func Info(args ...interface{}) {
	GetDefaultLogger().Info(args...)
}

// Infof 输出格式化信息日志
func Infof(format string, args ...interface{}) {
	GetDefaultLogger().Infof(format, args...)
}

// Warn 输出警告日志
func Warn(args ...interface{}) {
	GetDefaultLogger().Warn(args...)
}

// Warnf 输出格式化警告日志
func Warnf(format string, args ...interface{}) {
	GetDefaultLogger().Warnf(format, args...)
}

// Error 输出错误日志
func Error(args ...interface{}) {
	GetDefaultLogger().Error(args...)
}

// Errorf 输出格式化错误日志
func Errorf(format string, args ...interface{}) {
	GetDefaultLogger().Errorf(format, args...)
}

// Fatal 输出致命错误日志并退出程序
func Fatal(args ...interface{}) {
	GetDefaultLogger().Fatal(args...)
}

// Fatalf 输出格式化致命错误日志并退出程序
func Fatalf(format string, args ...interface{}) {
	GetDefaultLogger().Fatalf(format, args...)
}

// WithField 添加字段
func WithField(key string, value interface{}) Logger {
	return GetDefaultLogger().WithField(key, value)
}

// WithFields 添加多个字段
func WithFields(fields map[string]interface{}) Logger {
	return GetDefaultLogger().WithFields(fields)
}

// WithError 添加错误字段
func WithError(err error) Logger {
	return GetDefaultLogger().WithError(err)
} 