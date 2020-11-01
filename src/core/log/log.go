package log

import (
	"errors"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogType int8

const (
	DEBUG LogType = 1
	INFO  LogType = 2
	WARN  LogType = 3
	ERROR LogType = 4
)

var once sync.Once

var logger *zap.Logger
var logType LogType

func Init(logfile string, t LogType) {
	once.Do(func() {
		var err error

		var level zap.AtomicLevel
		switch t {
		case ERROR:
			level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
		case WARN:
			level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
		case INFO:
			level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		case DEBUG:
			level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		default:
			panic(errors.New("log type error"))
		}

		logType = t

		logger, err = zap.Config{
			Encoding:    "json",
			Level:       level,
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

func Log(msg string, err error, t LogType) {
	if logger != nil {
		if t < logType {
			return
		}
		switch t {
		case ERROR:
			logger.Error(msg, zap.Error(err))
		case WARN:
			logger.Warn(msg, zap.Error(err))
		case INFO:
			logger.Info(msg, zap.Error(err))
		case DEBUG:
			logger.Debug(msg, zap.Error(err))
		}
	}
}

func LogFields(msg string, t LogType, fields ...zapcore.Field) {
	if logger != nil {
		if t < logType {
			return
		}
		switch t {
		case ERROR:
			logger.Error(msg, fields...)
		case WARN:
			logger.Warn(msg, fields...)
		case INFO:
			logger.Info(msg, fields...)
		case DEBUG:
			logger.Debug(msg, fields...)
		}
	}
}

func Sync() {
	if logger != nil {
		logger.Sync()
	}
}
