package logs

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Level      string
	StdFormat  string
	FilePath   string
	FileFormat string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

var Out *zap.Logger = zap.NewNop()

func InitLogger(cfg *Config) {
	level := zapcore.InfoLevel
	if cfg.Level != "" {
		if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
			level = zapcore.InfoLevel
		}
	}

	baseConfig := zap.NewProductionEncoderConfig()
	baseConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	baseConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	consoleConfig := baseConfig
	if cfg.StdFormat == "console" {
		consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	cores := []zapcore.Core{
		zapcore.NewCore(getEncoder(cfg.StdFormat, consoleConfig), zapcore.AddSync(os.Stdout), level),
	}

	if cfg.FilePath != "" {
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		})
		cores = append(cores, zapcore.NewCore(getEncoder(cfg.FileFormat, baseConfig), fileWriter, level))
	}

	Out = zap.New(zapcore.NewTee(cores...))
}

func Sync() {
	_ = Out.Sync()
}

func getEncoder(format string, config zapcore.EncoderConfig) zapcore.Encoder {
	if format == "json" {
		return zapcore.NewJSONEncoder(config)
	}
	return zapcore.NewConsoleEncoder(config)
}
