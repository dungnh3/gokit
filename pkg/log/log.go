package log

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level for log
type Level int

const (
	_ = iota
	// Debug level
	Debug Level = iota + 1
	// Info level
	Info
	// Warn level
	Warn
	// Error level
	Error
	//DPanic level
	DPanic
	//Panic level
	Panic
	//Fatal level
	Fatal
)

//Logger interface, support native zap logger or sugared logger
type Logger interface {
	L(string, ...zapcore.Field)
	F(string, ...interface{})
}

// Config holding log config
type Config struct {
	Level    string `json:"level"`
	Mode     string `json:"mode"`
	Encoding string `json:"encoding"`
}

// DefaultConfig of logger, should set default mode to production to avoid mistake
func DefaultConfig() Config {
	return Config{
		Mode: "production",
	}
}

type levelMap map[string]zapcore.Level

func (lm levelMap) get(level string) (zapcore.Level, bool) {
	if lvl, ok := lm[strings.ToLower(level)]; ok {
		return lvl, ok
	}
	return zapcore.InfoLevel, false
}

var zapLevelMap = levelMap{
	"debug":  zap.DebugLevel,
	"info":   zap.InfoLevel,
	"warn":   zap.WarnLevel,
	"error":  zap.ErrorLevel,
	"dpanic": zap.DPanicLevel,
	"panic":  zap.PanicLevel,
	"fatal":  zap.FatalLevel,
}

// Build construct zap Logger by config
func (c Config) Build(opts ...zap.Option) (*zap.Logger, error) {

	zapConfig := zap.NewProductionConfig()
	if c.Mode == "development" {
		zapConfig = zap.NewDevelopmentConfig()
	}

	if lvl, ok := zapLevelMap.get(c.Level); ok {
		zapConfig.Level = zap.NewAtomicLevelAt(lvl)
	}

	for _, encoding := range []string{"json", "console"} {
		if encoding == c.Encoding {
			zapConfig.Encoding = c.Encoding
		}
	}

	return zapConfig.Build(opts...)
}
