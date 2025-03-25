package types

import (
	"log/slog"
	"net/http"
	"reflect"
)

type EventListener func(string, any)

type EventManager struct {
	SupportedEventTypes []string
	Listensers          []EventListener
}

type HttpRequestEventData struct {
	Method string
	Url    string
	Header http.Header
	Body   []byte
}

type HttpResponseEventData struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (m *EventManager) GetSupportedEventTypes() []string {
	return m.SupportedEventTypes
}

func (m *EventManager) AddListener(listener EventListener) {
	m.Listensers = append(m.Listensers, listener)
}

func (m *EventManager) Notify(eventType string, data any) {
	for _, supportedType := range m.SupportedEventTypes {
		if eventType == supportedType {
			for _, listener := range m.Listensers {
				listener(eventType, data)
			}
			return
		}
	}
	slog.Error("[EventManager] Unsupported event type", "eventType", eventType)
}

func (m *EventManager) RemoveListener(listener EventListener) {
	for i, fn := range m.Listensers {
		if reflect.ValueOf(fn).Pointer() == reflect.ValueOf(listener).Pointer() {
			// Remove the function pointer by slicing
			m.Listensers = append(m.Listensers[:i], m.Listensers[i+1:]...)
			return
		}
	}
}

func (m *EventManager) NotifyHTTPRequest(etype string, method string, url string, header http.Header, body []byte) {
	m.Notify(etype, HttpRequestEventData{method, url, header, body})
}

func (m *EventManager) NotifyHTTPResponse(etype string, statusCode int, header http.Header, body []byte) {
	m.Notify(etype, HttpResponseEventData{statusCode, header, body})
}
