//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
)

const (
	LoggerMaxSize    = 100
	LoggerMaxBackups = 30
	LoggerMaxAge     = 30
	LoggerCompress   = false
)

var loggerNameArray = []string{"logic", "api", "engine"}

var (
	LogicLogger  *slog.Logger
	ApiLogger    *slog.Logger
	EngineLogger *slog.Logger
)

type customLogger struct {
	lj      *lumberjack.Logger
	lastDay string
}

func (c *customLogger) Write(p []byte) (n int, err error) {
	now := time.Now().Format("2006-01-02")
	if now != c.lastDay {
		c.lj.Rotate()
		c.lastDay = now
	}

	return c.lj.Write(p)
}

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
		Filename:   filepath.Join(c.LogPath, name+".log"),
		MaxSize:    LoggerMaxSize,    // Maximum size of a single log file (MB)
		MaxBackups: LoggerMaxBackups, // Maximum number of old log files to keep
		MaxAge:     LoggerMaxAge,     // Maximum number of days reserved
		Compress:   LoggerCompress,
	}
	// Get file date
	fileInfo, err := os.Stat(c.LogPath)
	if err != nil && !os.IsExist(err) {
		_ = os.MkdirAll(c.LogPath, 0o750)
	}
	if _, err := os.Stat(lumberjackLogger.Filename); os.IsNotExist(err) {
		_, err = os.Create(lumberjackLogger.Filename)
		if err != nil {
			return
		}
		fileInfo, _ = os.Stat(lumberjackLogger.Filename)
	}

	creationTime := fileInfo.ModTime()

	day := creationTime.Format("2006-01-02")
	cl := &customLogger{
		lj:      lumberjackLogger,
		lastDay: day,
	}

	// Create a multi-writer to write to both file and stdout
	mw := io.MultiWriter(cl, os.Stdout)
	// Create separate handlers for console and file
	// Console handler: beautiful custom format with colors
	consoleHandler := &BeautifulConsoleHandler{
		level: logLevel,
		name:  name,
	}

	// File handler: JSON format
	fileHandler := slog.NewJSONHandler(mw, &slog.HandlerOptions{
		Level: logLevel,
	})

	// Create a custom handler that writes to both console (text) and file (JSON)
	multiHandler := &MultiFormatHandler{
		consoleHandler: consoleHandler,
		fileHandler:    fileHandler,
	}

	logger := slog.New(multiHandler)
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

// MultiFormatHandler implements slog.Handler to write different formats to console and file
type MultiFormatHandler struct {
	consoleHandler slog.Handler
	fileHandler    slog.Handler
}

func (h *MultiFormatHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.consoleHandler.Enabled(ctx, level) || h.fileHandler.Enabled(ctx, level)
}

func (h *MultiFormatHandler) Handle(ctx context.Context, record slog.Record) error {
	// Write to console with text format
	if h.consoleHandler.Enabled(ctx, record.Level) {
		if err := h.consoleHandler.Handle(ctx, record); err != nil {
			return err
		}
	}

	// Write to file with JSON format
	if h.fileHandler.Enabled(ctx, record.Level) {
		if err := h.fileHandler.Handle(ctx, record); err != nil {
			return err
		}
	}

	return nil
}

func (h *MultiFormatHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MultiFormatHandler{
		consoleHandler: h.consoleHandler.WithAttrs(attrs),
		fileHandler:    h.fileHandler.WithAttrs(attrs),
	}
}

func (h *MultiFormatHandler) WithGroup(name string) slog.Handler {
	return &MultiFormatHandler{
		consoleHandler: h.consoleHandler.WithGroup(name),
		fileHandler:    h.fileHandler.WithGroup(name),
	}
}

// ANSI color codes for beautiful console output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"

	// Background colors
	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
	BgBlue   = "\033[44m"

	// Styles
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Underline = "\033[4m"
)

// BeautifulConsoleHandler implements a beautiful console output format
type BeautifulConsoleHandler struct {
	level slog.Level
	name  string
	attrs []slog.Attr
	group string
}

func (h *BeautifulConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *BeautifulConsoleHandler) Handle(ctx context.Context, record slog.Record) error {
	// Format time with color
	timeStr := ColorGray + record.Time.Format("15:04:05") + ColorReset

	// Format level with appropriate color and icon
	var levelStr string
	switch record.Level {
	case slog.LevelDebug:
		levelStr = ColorCyan + "🔍 DBG" + ColorReset
	case slog.LevelInfo:
		levelStr = ColorGreen + "ℹ️  INF" + ColorReset
	case slog.LevelWarn:
		levelStr = ColorYellow + "⚠️  WRN" + ColorReset
	case slog.LevelError:
		levelStr = ColorRed + "❌ ERR" + ColorReset
	default:
		levelStr = ColorWhite + "   LOG" + ColorReset
	}

	// Format component name with color
	componentStr := ColorPurple + "[" + strings.ToUpper(h.name) + "]" + ColorReset

	// Format message
	messageStr := ColorWhite + record.Message + ColorReset

	// Collect attributes
	var attrPairs []string
	record.Attrs(func(a slog.Attr) bool {
		key := a.Key
		value := a.Value.String()

		// Special formatting for common attributes
		switch key {
		case "error":
			attrPairs = append(attrPairs, ColorRed+key+ColorReset+"="+ColorRed+value+ColorReset)
		case "url", "path", "file":
			attrPairs = append(attrPairs, ColorCyan+key+ColorReset+"="+ColorCyan+value+ColorReset)
		case "duration", "time":
			attrPairs = append(attrPairs, ColorYellow+key+ColorReset+"="+ColorYellow+value+ColorReset)
		case "count", "size", "port", "pid":
			attrPairs = append(attrPairs, ColorBlue+key+ColorReset+"="+ColorBlue+value+ColorReset)
		default:
			attrPairs = append(attrPairs, ColorGray+key+ColorReset+"="+value)
		}
		return true
	})

	// Add handler attributes
	for _, attr := range h.attrs {
		key := attr.Key
		value := attr.Value.String()
		attrPairs = append(attrPairs, ColorGray+key+ColorReset+"="+value)
	}

	// Build the final output
	var output strings.Builder
	output.WriteString(timeStr)
	output.WriteString(" ")
	output.WriteString(levelStr)
	output.WriteString(" ")
	output.WriteString(componentStr)
	output.WriteString(" ")
	output.WriteString(messageStr)

	if len(attrPairs) > 0 {
		output.WriteString(" " + ColorGray + "│" + ColorReset + " ")
		output.WriteString(strings.Join(attrPairs, " "))
	}

	output.WriteString("\n")

	// Write to stdout
	_, err := os.Stdout.Write([]byte(output.String()))
	return err
}

func (h *BeautifulConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &BeautifulConsoleHandler{
		level: h.level,
		name:  h.name,
		attrs: newAttrs,
		group: h.group,
	}
}

func (h *BeautifulConsoleHandler) WithGroup(name string) slog.Handler {
	group := name
	if h.group != "" {
		group = h.group + "." + name
	}

	return &BeautifulConsoleHandler{
		level: h.level,
		name:  h.name,
		attrs: h.attrs,
		group: group,
	}
}
