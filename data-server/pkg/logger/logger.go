package logger

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Level 日志级别
type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
	FatalLevel Level = "fatal"
	PanicLevel Level = "panic"
)

// Format 日志格式
type Format string

const (
	JSONFormat Format = "json"
	TextFormat Format = "text"
)

// Config 日志配置
type Config struct {
	Level      Level  `yaml:"level" json:"level"`
	Format     Format `yaml:"format" json:"format"`
	Output     string `yaml:"output" json:"output"`
	Filename   string `yaml:"filename" json:"filename"`
	MaxSize    int    `yaml:"max_size" json:"max_size"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups"`
	MaxAge     int    `yaml:"max_age" json:"max_age"`
	Compress   bool   `yaml:"compress" json:"compress"`
	ReportCaller bool `yaml:"report_caller" json:"report_caller"`
}

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
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

// loggerImpl logrus实现
type loggerImpl struct {
	*logrus.Logger
}

var defaultLogger Logger

// NewLogger 创建新的日志实例
func NewLogger(config *Config) (Logger, error) {
	log := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(string(config.Level))
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// 设置日志格式
	switch config.Format {
	case JSONFormat:
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	default:
		log.SetFormatter(&CustomTextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
			ForceColors:     true,
		})
	}

	// 设置输出
	var output io.Writer
	switch config.Output {
	case "file":
		if config.Filename == "" {
			return nil, fmt.Errorf("filename is required when output is file")
		}
		// 确保目录存在
		dir := filepath.Dir(config.Filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		output = &lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
	case "both":
		if config.Filename == "" {
			return nil, fmt.Errorf("filename is required when output is both")
		}
		dir := filepath.Dir(config.Filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
		fileOutput := &lumberjack.Logger{
			Filename:   config.Filename,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
		output = io.MultiWriter(os.Stdout, fileOutput)
	default:
		output = os.Stdout
	}

	log.SetOutput(output)
	log.SetReportCaller(config.ReportCaller)

	return &loggerImpl{Logger: log}, nil
}

// InitDefaultLogger 初始化默认日志
func InitDefaultLogger(config *Config) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

// GetDefaultLogger 获取默认日志
func GetDefaultLogger() Logger {
	if defaultLogger == nil {
		// 使用默认配置
		config := &Config{
			Level:  InfoLevel,
			Format: TextFormat,
			Output: "stdout",
		}
		logger, _ := NewLogger(config)
		defaultLogger = logger
	}
	return defaultLogger
}

// 全局日志函数
func Debug(args ...interface{}) {
	GetDefaultLogger().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	GetDefaultLogger().Debugf(format, args...)
}

func Info(args ...interface{}) {
	GetDefaultLogger().Info(args...)
}

func Infof(format string, args ...interface{}) {
	GetDefaultLogger().Infof(format, args...)
}

func Warn(args ...interface{}) {
	GetDefaultLogger().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	GetDefaultLogger().Warnf(format, args...)
}

func Error(args ...interface{}) {
	GetDefaultLogger().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	GetDefaultLogger().Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	GetDefaultLogger().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	GetDefaultLogger().Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	GetDefaultLogger().Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	GetDefaultLogger().Panicf(format, args...)
}

func WithField(key string, value interface{}) Logger {
	return GetDefaultLogger().WithField(key, value)
}

func WithFields(fields map[string]interface{}) Logger {
	return GetDefaultLogger().WithFields(fields)
}

func WithError(err error) Logger {
	return GetDefaultLogger().WithError(err)
}

// loggerImpl 实现
func (l *loggerImpl) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

func (l *loggerImpl) Debugf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

func (l *loggerImpl) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

func (l *loggerImpl) Infof(format string, args ...interface{}) {
	l.Logger.Infof(format, args...)
}

func (l *loggerImpl) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

func (l *loggerImpl) Warnf(format string, args ...interface{}) {
	l.Logger.Warnf(format, args...)
}

func (l *loggerImpl) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

func (l *loggerImpl) Errorf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

func (l *loggerImpl) Fatal(args ...interface{}) {
	l.Logger.Fatal(args...)
}

func (l *loggerImpl) Fatalf(format string, args ...interface{}) {
	l.Logger.Fatalf(format, args...)
}

func (l *loggerImpl) Panic(args ...interface{}) {
	l.Logger.Panic(args...)
}

func (l *loggerImpl) Panicf(format string, args ...interface{}) {
	l.Logger.Panicf(format, args...)
}

func (l *loggerImpl) WithField(key string, value interface{}) Logger {
	return &loggerImpl{Logger: l.Logger.WithField(key, value).Logger}
}

func (l *loggerImpl) WithFields(fields map[string]interface{}) Logger {
	return &loggerImpl{Logger: l.Logger.WithFields(fields).Logger}
}

func (l *loggerImpl) WithError(err error) Logger {
	return &loggerImpl{Logger: l.Logger.WithError(err).Logger}
}

// CustomTextFormatter 自定义文本格式化器
type CustomTextFormatter struct {
	TimestampFormat string
	FullTimestamp   bool
	ForceColors     bool
}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}

	// 获取调用者信息
	caller := ""
	if entry.HasCaller() {
		funcName := filepath.Base(entry.Caller.Function)
		fileName := filepath.Base(entry.Caller.File)
		caller = fmt.Sprintf(" [%s:%d %s]", fileName, entry.Caller.Line, funcName)
	}

	// 格式化级别
	level := strings.ToUpper(entry.Level.String())
	level = level[:4] // 只显示前4个字符

	// 格式化时间
	timestamp := entry.Time.Format(timestampFormat)

	// 构建消息
	var msg string
	if f.FullTimestamp {
		msg = fmt.Sprintf("%s%s [%s] %s", timestamp, caller, level, entry.Message)
	} else {
		msg = fmt.Sprintf("%s [%s] %s", timestamp, level, entry.Message)
	}

	// 添加字段
	if len(entry.Data) > 0 {
		msg += " |"
		for k, v := range entry.Data {
			msg += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	msg += "\n"
	return []byte(msg), nil
}

// RequestIDMiddleware 请求ID中间件日志
func LogRequestIDMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 创建响应写入器包装器
			wrapped := &responseWriter{ResponseWriter: w, status: 200}

			defer func() {
				duration := time.Since(start)
				logger.WithFields(map[string]interface{}{
					"method":     r.Method,
					"path":       r.URL.Path,
					"status":     wrapped.status,
					"duration":   duration,
					"remote_addr": r.RemoteAddr,
					"user_agent": r.UserAgent(),
				}).Info("HTTP Request")
			}()

			next.ServeHTTP(wrapped, r)
		})
	}
}

// responseWriter 响应写入器包装器
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetCaller 获取调用者信息
func GetCaller(skip int) (string, string, int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", "", 0
	}

	funcName := runtime.FuncForPC(pc).Name()
	fileName := filepath.Base(file)

	// 清理函数名
	if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
		funcName = funcName[lastSlash+1:]
	}
	if lastDot := strings.LastIndex(funcName, "."); lastDot >= 0 {
		funcName = funcName[lastDot+1:]
	}

	return funcName, fileName, line
}