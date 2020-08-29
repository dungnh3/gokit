package level

import (
	"context"
	"github.com/dungnh3/gokit/log"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultCallerSkip = 1
)

// ZapWrapper to short syntax
type ZapWrapper struct {
	*zap.Logger
	Level log.Level
}

// L logging message to stderr/stdout
func (zl *ZapWrapper) L(msg string, fields ...zapcore.Field) {
	switch zl.Level {
	case log.Debug:
		zl.Debug(msg, fields...)
	case log.Info:
		zl.Info(msg, fields...)
	case log.Warn:
		zl.Warn(msg, fields...)
	case log.Error:
		zl.Error(msg, fields...)
	case log.DPanic:
		zl.DPanic(msg, fields...)
	case log.Panic:
		zl.Panic(msg, fields...)
	case log.Fatal:
		zl.Fatal(msg, fields...)
	}
}

// F logging format message with template
func (zl *ZapWrapper) F(template string, args ...interface{}) {
	switch zl.Level {
	case log.Debug:
		zl.Sugar().Debugf(template, args...)
	case log.Info:
		zl.Sugar().Infof(template, args...)
	case log.Warn:
		zl.Sugar().Warnf(template, args...)
	case log.Error:
		zl.Sugar().Errorf(template, args...)
	case log.DPanic:
		zl.Sugar().DPanicf(template, args...)
	case log.Panic:
		zl.Sugar().Panicf(template, args...)
	case log.Fatal:
		zl.Sugar().Fatalf(template, args...)
	}
}

// DPanic return DPanic logger for development
func DPanic(ctx context.Context) log.Logger {
	return yieldLogger(ctx, log.DPanic, defaultCallerSkip)
}

//Debug return Debug logger
func Debug(ctx context.Context) log.Logger {
	return yieldLogger(ctx, log.Debug, defaultCallerSkip)
}

//Error return Error logger
func Error(ctx context.Context) log.Logger {
	return yieldLogger(ctx, log.Error, defaultCallerSkip)
}

//Info return Info logger
func Info(ctx context.Context) log.Logger {
	return yieldLogger(ctx, log.Info, defaultCallerSkip)
}

// Panic return panic logger
func Panic(ctx context.Context) log.Logger {
	return yieldLogger(ctx, log.Panic, defaultCallerSkip)
}

// Warn return warning logger
func Warn(ctx context.Context) log.Logger {
	return yieldLogger(ctx, log.Warn, defaultCallerSkip)
}

func yieldLogger(ctx context.Context, level log.Level, callerSkip int) log.Logger {
	zapLogger := ctxzap.Extract(ctx)
	return &ZapWrapper{
		Logger: zapLogger.WithOptions(zap.AddCallerSkip(callerSkip)),
		Level:  level,
	}
}
