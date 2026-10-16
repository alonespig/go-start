package log

import (
	"os"
	"path/filepath"

	"go-starter/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// init 在包被导入时自动用默认配置初始化全局 logger。
// 保证任何地方 import 本包即可直接打日志，即使忘了调 Init 也不会静默丢日志。
func init() {
	if logger, err := build(config.LogConfig{Level: "info"}); err == nil {
		zap.ReplaceGlobals(logger)
	}
}

// Init 用自定义配置（如 config.yaml 的 log 段）重新初始化全局 logger。
// 在加载完配置后调用，会覆盖 init() 的默认值。
func Init(cfg config.LogConfig) error {
	logger, err := build(cfg)
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}

func build(cfg config.LogConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, err
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var cores []zapcore.Core

	// 控制台输出
	cores = append(cores, zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	))

	// 文件输出（可选）
	if cfg.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0o755); err != nil {
			return nil, err
		}
		f, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, err
		}
		cores = append(cores, zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(f),
			level,
		))
	}

	// AddCallerSkip(1)：跳过本包的 Info/Error 等包装层，让 caller 指向真实调用处
	return zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1)), nil
}

// L 返回全局 logger。未 Init 时返回 no-op logger，不会 panic。
func L() *zap.Logger { return zap.L() }

// Info 结构化日志
func Info(msg string, fields ...zap.Field) { zap.L().Info(msg, fields...) }

// Error 结构化日志
func Error(msg string, fields ...zap.Field) { zap.L().Error(msg, fields...) }

// Warn 结构化日志
func Warn(msg string, fields ...zap.Field) { zap.L().Warn(msg, fields...) }

// Debug 结构化日志
func Debug(msg string, fields ...zap.Field) { zap.L().Debug(msg, fields...) }
