// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package opac

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
)

// A simple event system for UI and logging purpose
// To make it easier, listeners are notified for all events, it should do
// the filtering of the type of events by itself
// The event type is basically a string, to avoid misspell of event type,
// the allowed event types should be supplied when create the manager

type EventListener func(string, any)

type EventManager struct {
	supportedEventTypes []string
	listensers          []EventListener
}

func NewEventManager(supportedEventTypes []string) *EventManager {
	return &EventManager{
		supportedEventTypes: supportedEventTypes,
		listensers:          []EventListener{},
	}
}

func (m *EventManager) SupportedEventTypes() []string {
	return m.supportedEventTypes
}

func (m *EventManager) AddListener(listener EventListener) {
	m.listensers = append(m.listensers, listener)
}

func (m *EventManager) Notify(eventType string, data any) {
	for _, supportedType := range m.supportedEventTypes {
		if eventType == supportedType {
			for _, listener := range m.listensers {
				listener(eventType, data)
			}
			return
		}
	}
	slog.Error("[EventManager] Unsupported event type", "eventType", eventType)
}

func (m *EventManager) RemoveListener(listener EventListener) {
	for i, fn := range m.listensers {
		if reflect.ValueOf(fn).Pointer() == reflect.ValueOf(listener).Pointer() {
			// Remove the function pointer by slicing
			m.listensers = append(m.listensers[:i], m.listensers[i+1:]...)
			return
		}
	}
}

// We intentionally don't log time and even remote Date from header
// so that the log file is repeatable i.e. for the smae invocation,
// you can compare this log with previous
type FileLogger struct {
	logFilePath string
	isEnabled   bool
}

func NewFileLogger(logFilePath string) *FileLogger {
	return &FileLogger{
		logFilePath: logFilePath,
		isEnabled:   true,
	}
}

func (l *FileLogger) SetEnable(isEnabled bool) {
	l.isEnabled = isEnabled
}

func (l *FileLogger) Log(msg string) {
	if !l.isEnabled {
		return
	}
	file, err := os.OpenFile(l.logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Failed to open log file", "error", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(msg)
	if err != nil {
		slog.Error("Failed to write to log file", "error", err)
		return
	}
	writer.Flush()
}

func (l *FileLogger) LogSection(char string) {
	if !l.isEnabled {
		return
	}
	l.Log("\n" + strings.Repeat(char, 80) + "\n\n")
}

func (l *FileLogger) LogHTTPRequest(title string, method string, url string, header http.Header, body []byte) {
	if !l.isEnabled {
		return
	}
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
		buffer.WriteString(fmt.Sprintf("%-20s: %s\n", k, strings.Join(header[k], ", ")))
	}

	buffer.WriteString("\n" + BodyToString(header, body) + "\n\n")
	l.Log(buffer.String())
}

func (l *FileLogger) LogHTTPResponse(title string, statusCode int, header http.Header, body []byte) {
	if !l.isEnabled {
		return
	}
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
	buffer.WriteString("\n" + BodyToString(header, body) + "\n\n")
	l.Log(buffer.String())
}

type httpRequestEventData struct {
	Method string
	Url    string
	Header http.Header
	Body   []byte
}

type httpResponseEventData struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (m *EventManager) NotifyHTTPRequest(etype string, method string, url string, header http.Header, body []byte) {
	m.Notify(etype, httpRequestEventData{method, url, header, body})
}

func (m *EventManager) NotifyHTTPResponse(etype string, statusCode int, header http.Header, body []byte) {
	m.Notify(etype, httpResponseEventData{statusCode, header, body})
}

var SysEvents *EventManager

func InitSysEvents(log string) {
	SysEvents = NewEventManager([]string{"start_app", "start_session",
		"end_session", "receive_service_request", "request_converted_to_opac", "invoke_service_provider",
		"service_provider_response", "response_converted_to_opac", "send_back_response"})
	if log == "" {
		return
	}
	fl := NewFileLogger(log)
	SysEvents.AddListener(func(eventType string, data any) {
		switch eventType {
		case "start_app":
			fl.LogSection("#")
		case "start_session":
			fl.LogSection("=")
		case "receive_service_request", "request_converted_to_opac", "invoke_service_provider":
			d := data.(httpRequestEventData)
			fl.LogHTTPRequest(eventType, d.Method, d.Url, d.Header, d.Body)
		case "service_provider_response", "response_converted_to_opac", "send_back_response":
			d := data.(httpResponseEventData)
			fl.LogHTTPResponse(eventType, d.StatusCode, d.Header, d.Body)
		}
	})
}
