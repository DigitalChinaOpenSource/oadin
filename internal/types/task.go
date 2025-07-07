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

package types

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"intel.com/aog/internal/logger"
)

type ServiceResultType int

const (
	ServiceResultDone ServiceResultType = iota
	ServiceResultFailed
	ServiceResultChunk
)

type ServiceResult struct {
	Type       ServiceResultType
	TaskId     uint64
	Error      error
	StatusCode int
	HTTP       HTTPContent
}

func IsDropAction(err error) bool {
	if err == nil {
		return false
	}
	var dropAction *DropAction
	ok := errors.As(err, &dropAction)
	return ok
}

func (sr *ServiceResult) WriteBack(w http.ResponseWriter) {
	if IsDropAction(sr.Error) {
		return
	}
	if sr.Type == ServiceResultFailed {
		if httpError, ok := sr.Error.(*HTTPErrorResponse); ok {
			clear(w.Header())
			w.WriteHeader(httpError.StatusCode)
			for k, v := range httpError.Header {
				w.Header().Set(k, v[0])
			}
			_, _ = w.Write(httpError.Body)
			logger.LogicLogger.Info("send_back_response", httpError.StatusCode, w.Header(), httpError.Body)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		var errBytes []byte
		if sr.Error != nil {
			errBytes = []byte(sr.Error.Error())
			_, _ = w.Write(errBytes)
		}
	} else {
		clear(w.Header())
		w.WriteHeader(sr.StatusCode)
		for k, v := range sr.HTTP.Header {
			w.Header().Set(k, v[0])
		}
		_, _ = w.Write(sr.HTTP.Body)
	}
}

func (sr *ServiceResult) String() string {
	var stype string
	switch sr.Type {
	case ServiceResultDone:
		stype = "ServiceResultDone"
	case ServiceResultFailed:
		stype = "ServiceResultFailed"
	case ServiceResultChunk:
		stype = "ServiceResultChunk"
	}
	return fmt.Sprintf("ServiceResult{Type: %s, TaskId: %d, StatusCode: %d, Error: %v, HTTP {Header: %+v, Body: %s}}",
		stype, sr.StatusCode, sr.TaskId, sr.Error, sr.HTTP.Header, string(sr.HTTP.Body))
}

// ServiceRequest The body of the OriginalRequest has been read out so need to placed here
type ServiceRequest struct {
	AskStreamMode         bool          `json:"stream"`
	Model                 string        `json:"model"`
	HybridPolicy          string        `json:"hybrid_policy"`
	RemoteServiceProvider string        `json:"remote_service_provider"`
	FromFlavor            string        `json:"-"`
	Service               string        `json:"-"`
	Priority              int           `json:"-"`
	RequestSegments       int           `json:"request_segments"`
	RequestExtraUrl       string        `json:"extra_url"`
	HTTP                  HTTPContent   `json:"-"`
	OriginalRequest       *http.Request `json:"-"`
	WebSocketConnID       string        `json:"-"` // WebSocket连接ID，用于关联同一连接上的多个消息
}

func (sr *ServiceRequest) String() string {
	wsInfo := ""
	if sr.WebSocketConnID != "" {
		wsInfo = fmt.Sprintf(", WebSocketConnID: %s", sr.WebSocketConnID)
	}
	return fmt.Sprintf("ServiceRequest{FromFlavor: %s, Service: %s, Model: %s, Stream: %t, Hybrid: %s%s}",
		sr.FromFlavor, sr.Service, sr.Model, sr.AskStreamMode, sr.HybridPolicy, wsInfo)
}

type ServiceTarget struct {
	Location        string
	Stream          bool
	Model           string
	ToFavor         string
	XPU             string
	Protocol        string // 服务协议类型，如GRPC_STREAM
	ServiceProvider *ServiceProvider
}

func (sr *ServiceTarget) String() string {
	return fmt.Sprintf("ServiceDispatch{Location: %s, Provider: %s}", sr.Location, sr.ServiceProvider.ProviderName)
}

type StreamModeType int

const (
	StreamModeNonStream StreamModeType = iota
	StreamModeEventStream
	StreamModeNDJson
)

func (st StreamModeType) String() string {
	switch st {
	case StreamModeNonStream:
		return "NonStream"
	case StreamModeEventStream:
		return "EventStream"
	case StreamModeNDJson:
		return "NDJson"
	}
	return "Unknown"
}

type StreamMode struct {
	Mode   StreamModeType
	Header http.Header
}

func (sm *StreamMode) IsStream() bool {
	return sm.Mode != StreamModeNonStream
}

// ReadChunk the chunks of the stream is delimited by "\n\n" for event-stream, and "\n"
// or "\r\n" for x=ndjson so only need to read to '\n'. However, reader can only
// stop at one character, so we need to handle it by ourselves for "\n\n" case
func (sm *StreamMode) ReadChunk(reader *bufio.Reader) ([]byte, error) {
	if sm.Mode == StreamModeNonStream {
		return io.ReadAll(reader)
	}

	if sm.Mode == StreamModeNDJson {
		return reader.ReadBytes('\n')
	}

	var buffer bytes.Buffer
	var line []byte
	var err error
	for {
		line, err = reader.ReadBytes('\n')
		if err != nil && err != io.EOF { // real error
			break
		}
		buffer.Write(line)
		if err == io.EOF { // the end
			break
		}
		// Check the character before '\n'
		if buffer.Len() >= 2 && buffer.Bytes()[buffer.Len()-2] == '\n' && buffer.Bytes()[buffer.Len()-1] == '\n' {
			break
		}
	}
	return buffer.Bytes(), err
}

// UnwrapChunk Get real data
// event-stream is started with "data: " which need to be removed
// TODO: event-stream may contain multiple fields, and we only need to pick
// the data: fields. Need to handle it in the future. Currently we assume there
// is only a data: field.
func (sm *StreamMode) UnwrapChunk(chunk []byte) []byte {
	// remove "data: " at the beginning of the chunk
	if sm.Mode == StreamModeEventStream {
		if len(chunk) >= 6 && bytes.HasPrefix(chunk, []byte("data: ")) {
			return chunk[6:]
		}
	}
	return chunk
}

func (sm *StreamMode) WrapChunk(chunk []byte) []byte {
	if sm.Mode == StreamModeNonStream {
		return chunk
	}
	n := 0
	for i := len(chunk) - 1; i >= 0 && chunk[i] == '\n'; i-- {
		n++
	}

	nNeed := 0
	if sm.Mode == StreamModeNDJson {
		nNeed = 1
	} else if sm.Mode == StreamModeEventStream {
		nNeed = 2
	}

	if n == nNeed {
		return chunk
	} else if n > nNeed {
		return chunk[:len(chunk)-n+nNeed]
	} else {
		for range nNeed - n {
			chunk = append(chunk, '\n')
		}
	}

	if sm.Mode == StreamModeEventStream && !bytes.HasPrefix(chunk, []byte("data: ")) {
		chunk = append([]byte("data: "), chunk...)
	}
	return chunk
}
