package logger

import (
	"log/slog"
	"strings"

	"github.com/natefinch/lumberjack"
)

// GlobalLogger Global logger
var GlobalLogger *slog.Logger

type NewLogConfig struct {
	LogLevel string `json:"log_level"`
	LogPath  string `json:"log_path"`
}

func NewSysLogger(c NewLogConfig) {
	// Configuring lumberjack for log file management
	lumberjackLogger := &lumberjack.Logger{
		Filename:   c.LogPath,
		MaxSize:    100, // Maximum size of a single log file (MB)
		MaxBackups: 7,   // Maximum number of old log files to keep
		MaxAge:     0,   // Maximum number of days reserved
		Compress:   true,
	}

	// Create a log handler in JSON format
	var logLevel slog.Level
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError

	}

	jsonHandler := slog.NewJSONHandler(lumberjackLogger, &slog.HandlerOptions{
		Level: logLevel,
	})

	// Create a global logger
	GlobalLogger = slog.New(jsonHandler)
	slog.SetDefault(GlobalLogger)
}

// GetModuleLogger Partial logger example
func GetModuleLogger(module string) *slog.Logger {
	// Create a local logger for a specific module, adding the module name as context information
	return GlobalLogger.With("module", module)
}
