package accesslog

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var once sync.Once
var logger *zap.Logger

func Init(logfile string) {
	once.Do(func() {
		var err error
		logger, err = zap.Config{
			Encoding:    "json",
			Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
			OutputPaths: []string{"stdout", logfile},
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey:   "message",
				LevelKey:     "level",
				EncodeLevel:  zapcore.CapitalLevelEncoder, // INFO
				TimeKey:      "time",
				EncodeTime:   zapcore.ISO8601TimeEncoder,
				CallerKey:    "caller",
				EncodeCaller: zapcore.ShortCallerEncoder,
			},
		}.Build()
		if err != nil {
			panic(err)
		}
	})
}

func Log(msg, key, value string) {
	if logger != nil {
		logger.Info(msg, zap.String(key, value))
	}
}

func LogFields(msg string, fields ...zapcore.Field) {
	if logger != nil {
		logger.Info(msg, fields...)
	}
}

func Sync() {
	if logger != nil {
		logger.Sync()
	}
}
