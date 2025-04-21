// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"byze/config"
	"byze/internal/logger"
	"byze/internal/types"
	"byze/internal/utils"
)

// A simple event system for UI and logging purpose
// To make it easier, listeners are notified for all events, it should do
// the filtering of the type of events by itself
// The event type is basically a string, to avoid misspell of event type,
// the allowed event types should be supplied when create the manager

func NewEventManager(supportedEventTypes []string) *types.EventManager {
	return &types.EventManager{
		SupportedEventTypes: supportedEventTypes,
		Listensers:          []types.EventListener{},
	}
}

func NewFileLogger(logFilePath string) *slog.Logger {
	logger.NewSysLogger(logger.NewLogConfig{
		LogLevel: config.GlobalByzeEnvironment.LogLevel,
		LogPath:  logFilePath,
	})
	return logger.GlobalLogger
}

func LogHTTPRequest(l *slog.Logger, title string, method string, url string, header http.Header, body []byte) {
	//if !l.isEnabled {
	//	return
	//}
	var buffer bytes.Buffer
	method = strings.ToUpper(method)
	buffer.WriteString(fmt.Sprintf("\n------------------ >>> %s >>> ------------------\n\n", title))
	buffer.WriteString(fmt.Sprintf("%-6s    %s\n\n", method, url))

	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		l.Info(fmt.Sprintf("%-20s: %s\n", k, strings.Join(header[k], ", ")))
	}

	buffer.WriteString("\n" + utils.BodyToString(header, body) + "\n\n")
	l.Info(buffer.String())
}

func LogHTTPResponse(l *slog.Logger, title string, statusCode int, header http.Header, body []byte) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("\n------------------ <<< %s <<< ------------------\n\n", title))
	buffer.WriteString(fmt.Sprintf("Status Code: %d\n\n", statusCode))

	keys := make([]string, 0, len(header))
	for k := range header {
		if k != "Date" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		buffer.WriteString(fmt.Sprintf("%-20s: %s\n", k, strings.Join(header[k], ", ")))
	}
	buffer.WriteString("\n" + utils.BodyToString(header, body) + "\n\n")
	l.Info(buffer.String())
}

var SysEvents *types.EventManager

func InitSysEvents() {
	SysEvents = NewEventManager([]string{
		"start_app", "start_session",
		"end_session", "receive_service_request", "request_converted_to_byze", "invoke_service_provider",
		"service_provider_response", "response_converted_to_byze", "send_back_response",
	})
	if config.GlobalByzeEnvironment.LogHTTP == "" {
		return
	}
	fl := NewFileLogger(config.GlobalByzeEnvironment.LogHTTP)
	testLog := logger.GetModuleLogger("test")
	testLog.Info("test start http log")
	SysEvents.AddListener(func(eventType string, data any) {
		switch eventType {
		case "start_app":
			fl.Info("start app")
		case "start_session":
			fl.Info("start session")
		case "receive_service_request", "request_converted_to_byze", "invoke_service_provider":
			d := data.(types.HttpRequestEventData)
			LogHTTPRequest(fl, eventType, d.Method, d.Url, d.Header, d.Body)
		case "service_provider_response", "response_converted_to_byze", "send_back_response":
			d := data.(types.HttpResponseEventData)
			LogHTTPResponse(fl, eventType, d.StatusCode, d.Header, d.Body)
		}
	})
}
