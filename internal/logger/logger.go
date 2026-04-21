package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var (
	// globalLogger 全局日志实例
	globalLogger *logrus.Logger
	quietMode    bool = false // 安静模式标志
)

// Init 初始化日志系统
func Init(logLevel string, logFile string) error {
	globalLogger = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	globalLogger.SetLevel(level)

	// 设置日志格式
	globalLogger.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		DisableColors:    false,
		FullTimestamp:    false, // 关闭时间戳，简化输出
		TimestampFormat:  "15:04:05",
		DisableSorting:   false,
	})

	// 设置输出
	if logFile != "" {
		// 创建日志目录
		logDir := filepath.Dir(logFile)
		if logDir != "." {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return err
			}
		}

		// 打开日志文件
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		// 同时输出到文件和控制台
		multiWriter := io.MultiWriter(os.Stdout, file)
		globalLogger.SetOutput(multiWriter)
	} else {
		globalLogger.SetOutput(os.Stdout)
	}

	return nil
}

// GetLogger 获取全局日志实例
func GetLogger() *logrus.Logger {
	if globalLogger == nil {
		// 默认初始化
		_ = Init("info", "")
	}
	return globalLogger
}

// SetLevel 动态设置日志级别
func SetLevel(level logrus.Level) {
	logger := GetLogger()
	logger.SetLevel(level)
}

// SetFormatter 设置日志格式
func SetFormatter(formatter logrus.Formatter) {
	logger := GetLogger()
	logger.SetFormatter(formatter)
}

// SetQuiet 设置安静模式
func SetQuiet(quiet bool) {
	quietMode = quiet
}

// Result 输出扫描结果（简洁格式）
func Result(status int, path string, size int64) {
	// 根据状态码选择颜色
	var color string
	switch {
	case status >= 200 && status < 300:
		color = "\033[32m" // 绿色 - 成功
	case status >= 300 && status < 400:
		color = "\033[33m" // 黄色 - 重定向
	case status >= 400 && status < 500:
		color = "\033[31m" // 红色 - 客户端错误
	case status >= 500:
		color = "\033[31m" // 红色 - 服务器错误
	default:
		color = "\033[35m" // 紫色 - 未知
	}
	reset := "\033[0m"

	if quietMode {
		// 安静模式：直接输出，无日志前缀
		fmt.Printf("[+] %s%d%s - %s - %d bytes\n", color, status, reset, path, size)
	} else {
		// 普通模式：通过日志系统
		GetLogger().Infof("[+] %s%d%s - %s - %d bytes", color, status, reset, path, size)
	}
}

// Debug 记录调试日志
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf 记录格式化调试日志
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info 记录信息日志
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof 记录格式化信息日志
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn 记录警告日志
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf 记录格式化警告日志
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error 记录错误日志
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf 记录格式化错误日志
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal 记录致命日志并退出
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf 记录格式化致命日志并退出
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// Panic 记录恐慌日志并抛出 panic
func Panic(args ...interface{}) {
	GetLogger().Panic(args...)
}

// Panicf 记录格式化恐慌日志并抛出 panic
func Panicf(format string, args ...interface{}) {
	GetLogger().Panicf(format, args...)
}

// WithFields 创建带字段的日志条目
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithField 创建带单个字段的日志条目
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// WithError 创建带错误的日志条目
func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}
