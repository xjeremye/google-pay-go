package logger

import (
	"os"

	"github.com/golang-pay-core/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// InitLogger 初始化日志
func InitLogger() error {
	var core zapcore.Core
	var encoder zapcore.Encoder

	// 设置编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if config.Cfg.Log.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置日志级别
	var level zapcore.Level
	switch config.Cfg.Log.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 设置输出
	if config.Cfg.Log.Output == "file" {
		writer := &lumberjack.Logger{
			Filename:   config.Cfg.Log.FilePath,
			MaxSize:    config.Cfg.Log.MaxSize,
			MaxBackups: config.Cfg.Log.MaxBackups,
			MaxAge:     config.Cfg.Log.MaxAge,
			Compress:   config.Cfg.Log.Compress,
		}
		core = zapcore.NewCore(encoder, zapcore.AddSync(writer), level)
	} else {
		core = zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	}

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

// Sync 同步日志
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}
