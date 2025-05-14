package logger

import (
	"log/slog"
	"strings"

	"github.com/natefinch/lumberjack"
)

const (
	LoggerMaxSize    = 100
	LoggerMaxBackups = 7
	LoggerMaxAge     = 0
	LoggerCompress   = true
)

var loggerNameArray = []string{"logic", "api", "engine"}

var (
	LogicLogger  *slog.Logger
	ApiLogger    *slog.Logger
	EngineLogger *slog.Logger
)

type LogConfig struct {
	LogLevel string `json:"log_level"`
	LogPath  string `json:"log_path"`
}

type LogManager struct {
	loggers map[string]*slog.Logger
}

func GetLoggerLevel(loggerLevel string) slog.Level {
	var logLevel slog.Level
	switch strings.ToLower(loggerLevel) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError

	default:
		logLevel = slog.LevelWarn
	}
	return logLevel
}

func NewLogManager(c LogConfig) *LogManager {
	// Configuring lumberjack for log file management
	lm := &LogManager{
		loggers: make(map[string]*slog.Logger),
	}
	for _, name := range loggerNameArray {
		lm.AddLogger(c, name)
	}
	return lm
}

func (lm *LogManager) AddLogger(c LogConfig, name string) {
	logLevel := GetLoggerLevel(c.LogLevel)
	lumberjackLogger := &lumberjack.Logger{
		Filename:   c.LogPath + "/" + name + ".log",
		MaxSize:    LoggerMaxSize,    // Maximum size of a single log file (MB)
		MaxBackups: LoggerMaxBackups, // Maximum number of old log files to keep
		MaxAge:     LoggerMaxAge,     // Maximum number of days reserved
		Compress:   LoggerCompress,
	}

	// Create a log handler in JSON format
	jsonHandler := slog.NewJSONHandler(lumberjackLogger, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(jsonHandler)
	lm.loggers[name] = logger
}

func (lm *LogManager) GetLogger(name string) *slog.Logger {
	return lm.loggers[name]
}

func InitLogger(c LogConfig) {
	lm := NewLogManager(c)
	LogicLogger = lm.GetLogger("logic")
	ApiLogger = lm.GetLogger("api")
	EngineLogger = lm.GetLogger("engine")
}
