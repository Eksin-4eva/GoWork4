// Package logger 基于 zap 提供全局日志能力。
//
// 约定：业务代码一律使用本包的 Infof / Warnf / Errorf 等函数，
// 不要直接使用 fmt.Println 或 log 标准库。
package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 支持的日志级别。
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

var sugar = zap.NewNop().Sugar()

// Init 初始化全局 logger。
//
// 输出到 stdout（容器环境下由运行时负责收集），带调用位置，
// Error 及以上级别自动附带堆栈。
func Init(serviceName, level string) {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		parseLevel(level),
	)

	l := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zap.String("service", serviceName)),
	)

	sugar = l.Sugar()
	zap.ReplaceGlobals(l)
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Debugf 输出调试日志。
func Debugf(format string, args ...any) { sugar.Debugf(format, args...) }

// Infof 输出普通日志。
func Infof(format string, args ...any) { sugar.Infof(format, args...) }

// Warnf 输出警告日志。
func Warnf(format string, args ...any) { sugar.Warnf(format, args...) }

// Errorf 输出错误日志（自动附带堆栈）。
func Errorf(format string, args ...any) { sugar.Errorf(format, args...) }

// Fatalf 输出错误日志并退出进程。
func Fatalf(format string, args ...any) { sugar.Fatalf(format, args...) }
