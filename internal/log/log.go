// Package log 基于标准库 log/slog 构建结构化 logger。
package log

import (
	"log/slog"
	"os"
	"strings"

	"github.com/Shinonome-Ryoko14/Distributed-Seckill/internal/config"
)

// New 按配置构建结构化 logger：format 为 text 时用文本，否则用 JSON。
func New(cfg config.LogConfig) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(cfg.Level)}

	var handler slog.Handler
	if strings.EqualFold(cfg.Format, "text") {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
