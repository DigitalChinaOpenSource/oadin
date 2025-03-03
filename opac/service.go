// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package opac

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var textContentTypes = []string{"text/", "application/json", "application/xml", "application/javascript", "application/x-ndjson"}

func IsHTTPText(header http.Header) bool {
	if contentType := header.Get("Content-Type"); contentType != "" {
		ct := strings.ToLower(contentType)
		for _, t := range textContentTypes {
			if strings.Contains(ct, t) {
				return true
			}
		}
	}
	return false
}

func BodyToString(header http.Header, body []byte) string {
	if IsHTTPText(header) {
		return string(body)
	}
	return fmt.Sprintf("<Binary Data: %d bytes>", len(body))
}

type HTTPContent struct {
	Body   []byte
	Header http.Header
}

func (hc HTTPContent) String() string {
	return fmt.Sprintf("HTTPContent{Header: %+v, Body: %s}", hc.Header, BodyToString(hc.Header, hc.Body))
}

type HTTPErrorResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (hc *HTTPErrorResponse) Error() string {
	return fmt.Sprintf("HTTPErrorResponse{StatusCode: %d, Header: %+v, Body: %s}", hc.StatusCode, hc.Header, BodyToString(hc.Header, hc.Body))
}

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
			w.Write(httpError.Body)
			SysEvents.NotifyHTTPResponse("send_back_response", httpError.StatusCode, w.Header(), httpError.Body)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		var errBytes []byte
		if sr.Error != nil {
			errBytes = []byte(sr.Error.Error())
			w.Write(errBytes)
		}
		SysEvents.NotifyHTTPResponse("send_back_response", http.StatusInternalServerError, w.Header(), errBytes)
	} else {
		clear(w.Header())
		w.WriteHeader(sr.StatusCode)
		for k, v := range sr.HTTP.Header {
			w.Header().Set(k, v[0])
		}
		SysEvents.NotifyHTTPResponse("send_back_response", sr.StatusCode, w.Header(), sr.HTTP.Body)
		w.Write(sr.HTTP.Body)
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
		stype, sr.StatusCode, sr.TaskId, sr.Error, sr.HTTP.Header, BodyToString(sr.HTTP.Header, sr.HTTP.Body))
}

// The body of the OriginalRequest has been read out so need to placed here
type ServiceRequest struct {
	AskStreamMode         bool          `json:"stream"`
	Model                 string        `json:"model"`
	HybridPolicy          string        `json:"hybrid_policy"`
	RemoteServiceProvider string        `json:"remote_service_provider"`
	FromFlavor            string        `json:"-"`
	Service               string        `json:"-"`
	Priority              int           `json:"-"`
	HTTP                  HTTPContent   `json:"-"`
	OriginalRequest       *http.Request `json:"-"`
}

func (sr *ServiceRequest) String() string {
	return fmt.Sprintf("ServiceRequest{FromFlavor: %s, Service: %s, Model: %s, Stream: %t, Hybrid: %s}",
		sr.FromFlavor, sr.Service, sr.Model, sr.AskStreamMode, sr.HybridPolicy)
}

type ServiceTarget struct {
	Location        string
	Stream          bool
	Model           string
	ToFavor         string
	XPU             string
	ServiceProvider *ServiceProviderInfo
}

func (sr *ServiceTarget) String() string {
	return fmt.Sprintf("ServiceDispatch{Location: %s, Provider: %s}", sr.Location, sr.ServiceProvider)
}

type ServiceTask struct {
	Request  *ServiceRequest
	Target   *ServiceTarget
	Ch       chan *ServiceResult
	Error    error
	Schedule ScheduleDetails
}

func (st *ServiceTask) String() string {
	return fmt.Sprintf("ServiceTask{Id: %d, Request: %s, Target: %s}", st.Schedule.Id, st.Request, st.Target)
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

func NewStreamMode(header http.Header) *StreamMode {
	mode := StreamModeNonStream
	if contentType := header.Get("Content-Type"); contentType != "" {
		ct := strings.ToLower(contentType)
		if strings.Contains(ct, "text/event-stream") {
			mode = StreamModeEventStream
		} else if strings.Contains(ct, "application/x-ndjson") {
			mode = StreamModeNDJson
		}
	}
	return &StreamMode{Mode: mode, Header: header.Clone()}
}

func (sm *StreamMode) IsStream() bool {
	return sm.Mode != StreamModeNonStream
}

// the chunks of the stream is delimited by "\n\n" for event-stream, and "\n"
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

// Get real data
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
	}
	if n > nNeed {
		return chunk[:len(chunk)-n+nNeed]
	}
	for i := 0; i < nNeed-n; i++ {
		chunk = append(chunk, '\n')
	}
	if sm.Mode == StreamModeEventStream && !bytes.HasPrefix(chunk, []byte("data: ")) {
		chunk = append([]byte("data: "), chunk...)
	}
	return chunk
}

func (st *ServiceTask) Run() error {
	if st.Target == nil || st.Target.ServiceProvider == nil {
		panic("[Service] ServiceTask is not dispatched before it goes to Run() " + st.String())
	}
	if st.Request.Model != "" && st.Target.Model != "" && st.Request.Model != st.Target.Model {
		slog.Warn("[Service] Model Mismatch", "mode_in_request", st.Request.Model,
			"model_to_use", st.Target.Model, "service_provider_id", st.Target.ServiceProvider.Id,
			"taskid", st.Schedule.Id)
	}
	if st.Request.AskStreamMode && !st.Target.Stream {
		slog.Warn("[Service] Request asks for stream mode but it is not supported by the service provider",
			"service_provider_id", st.Target.ServiceProvider.Id, "taskid", st.Schedule.Id)
	}
	// ------------------------------------------------------------------
	// 1. Get flavors and convert request if necessary
	// ------------------------------------------------------------------
	sp := GetPlatformInfo().GetServiceProviderInfo(st.Request.Service, st.Target.Location)
	if sp == nil {
		return fmt.Errorf("service Provider not found for %s of Service %s", st.Target.Location, st.Request.Service)
	}
	requestFlavor, err := GetAPIFlavor(st.Request.FromFlavor)
	if err != nil {
		slog.Error("[Service] Unsupported API Flavor for Request", "task", st, "error", err)
		return fmt.Errorf("[Service] Unsupported API Flavor %s for Request: %s", st.Request.FromFlavor, err.Error())
	}
	targetFlavor, err := GetAPIFlavor(st.Target.ServiceProvider.Flavor)
	if err != nil {
		slog.Error("[Service] Unsupported API Flavor for Service Provider", "task", st, "error", err)
		return fmt.Errorf("[Service] Unsupported API Flavor %s for Service Provider: %s", st.Target.ServiceProvider.Flavor, err.Error())
	}

	conversionNeeded := targetFlavor.Name() != requestFlavor.Name()
	content := st.Request.HTTP

	if conversionNeeded {
		slog.Info("[Service] Converting Request", "taskid", st.Schedule.Id, "from flavor", requestFlavor.Name(), "to flavor", targetFlavor.Name())
		requestCtx := ConvertContext{"stream": st.Target.Stream}
		if st.Target.Model != "" {
			requestCtx["model"] = st.Target.Model
		}

		var err error
		content, err = ConvertBetweenFlavors(requestFlavor, targetFlavor, st.Request.Service, "request", content, requestCtx)

		if err != nil {
			slog.Error("[Service] Failed to convert request", "taskid", st.Schedule.Id, "from flavor", requestFlavor.Name(),
				"to flavor", targetFlavor.Name(), "error", err, "content", content)
			return fmt.Errorf("[Service] Failed to convert request: %s", err.Error())
		}
	}

	// ------------------------------------------------------------------
	// 2. Invoke the service provider and get response
	// ------------------------------------------------------------------

	invokeURL := sp.Url

	if strings.ToUpper(sp.Method) == "GET" {
		// the body could be empty
		// or it is GET with parameters, but the paramters should have beeen
		// marshaled in InvokeService() and maybe even converted above
		if len(content.Body) > 0 {
			queryParams := make(map[string][]string)
			err := json.Unmarshal(content.Body, &queryParams)
			if err != nil {
				slog.Error("[Service] Failed to unmarshal GET request", "taskid",
					st.Schedule.Id, "error", err, "body", string(content.Body))
				return err
			}
			u, err := url.Parse(sp.Url)
			if err != nil {
				slog.Error("Error parsing Service Provider's URL", "taskid",
					st.Schedule.Id, "sp.Url", sp.Url, "error", err)
				return err
			}

			q := u.Query()
			for key, values := range queryParams {
				for _, value := range values {
					q.Add(key, value)
				}
			}

			u.RawQuery = q.Encode()
			invokeURL = u.String()

			content.Body = nil
		}
	}

	req, err := http.NewRequest(sp.Method, invokeURL, bytes.NewReader(content.Body))
	if err != nil {
		return err
	}

	for k, v := range content.Header {
		if k != "Content-Length" {
			req.Header.Set(k, v[0])
		}
	}
	// TODO: further fine tuning of the transport
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: transport}
	slog.Info("[Service] Request Sending to Service Provider ...", "taskid", st.Schedule.Id, "url", req.URL.String())
	slog.Debug("[Service] Request Sending to Service Provider ...", "taskid", st.Schedule.Id, "method",
		req.Method, "url", req.URL.String(), "header", fmt.Sprintf("%+v", req.Header), "body", string(content.Body))
	SysEvents.NotifyHTTPRequest("invoke_service_provider", req.Method, req.URL.String(), content.Header, content.Body)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var sbody string
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			sbody = string(b)
		}
		slog.Warn("[Service] Service Provider returns Error", "taskid", st.Schedule.Id,
			"status_code", resp.StatusCode, "body", sbody)
		return &HTTPErrorResponse{
			StatusCode: resp.StatusCode,
			Header:     resp.Header.Clone(),
			Body:       b,
		}
	}

	slog.Debug("[Service] Response Receiving", "taskid", st.Schedule.Id, "header",
		fmt.Sprintf("%+v", resp.Header), "task", st)
	// ------------------------------------------------------------------
	// 3. Convert response if necessary and send back to handler
	// ------------------------------------------------------------------
	respStreamMode := NewStreamMode(resp.Header)

	slog.Debug("[Service] Response is Stream?", "taskid", st.Schedule.Id, "stream", respStreamMode.Mode.String())

	// in case response to send out needs a id but not in response returned from service provider
	respConvertCtx := ConvertContext{"id": fmt.Sprintf("%d%d", rand.Uint64(), st.Schedule.Id)}

	if !respStreamMode.IsStream() {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("[Service] Failed to read response body", "taskid", st.Schedule.Id, "error", err.Error())
			return fmt.Errorf("[Service] Failed to read response body: %s", err.Error())
		}

		slog.Debug("[Service] Response Content (non-stream)", "taskid", st.Schedule.Id, "body", BodyToString(resp.Header, body))
		SysEvents.NotifyHTTPResponse("service_provider_response", resp.StatusCode, resp.Header, body)

		content = HTTPContent{Body: body, Header: resp.Header.Clone()}

		if conversionNeeded {
			content, err = ConvertBetweenFlavors(targetFlavor, requestFlavor, st.Request.Service, "response", content, respConvertCtx)
			if err != nil {
				slog.Error("[Service] Failed to convert response", "taskid", st.Schedule.Id, "from flavor", targetFlavor.Name(),
					"to flavor", requestFlavor.Name(), "error", err, "content", content)
				return fmt.Errorf("[Service] Failed to convert response: %s", err.Error())
			}
		}

		st.Ch <- &ServiceResult{Type: ServiceResultDone, TaskId: st.Schedule.Id,
			StatusCode: resp.StatusCode,
			HTTP:       content}
	} else {
		isFirstTrunk := true
		reader := bufio.NewReader(resp.Body)
		prolog := requestFlavor.GetStreamResponseProlog(st.Request.Service)
		epilog := requestFlavor.GetStreamResponseEpilog(st.Request.Service)
		var sendBackConvertedStreamMode *StreamMode // only used if need conversion
		for {
			chunk, readChunkErr := respStreamMode.ReadChunk(reader)
			if readChunkErr != nil && readChunkErr != io.EOF { // real error
				slog.Error("[Service] Stream: Failed to read chunk", "taskid", st.Schedule.Id, "error", readChunkErr.Error())
				return readChunkErr
			}
			SysEvents.NotifyHTTPResponse("service_provider_response", resp.StatusCode, resp.Header, chunk)

			if readChunkErr == io.EOF {
				slog.Debug("[Service] Stream: Got EOF Response", "taskid", st.Schedule.Id, "chunk", string(chunk))
			} else {
				slog.Debug("[Service] Stream: Got Chunk Response", "taskid", st.Schedule.Id, "chunk", string(chunk))
			}

			content = HTTPContent{Body: chunk, Header: resp.Header.Clone()}
			var convertErr error
			if conversionNeeded { // need convert response
				content.Body = respStreamMode.UnwrapChunk(content.Body)
				// drop empty content
				if len(bytes.TrimSpace(chunk)) == 0 {
					convertErr = &DropAction{}
					slog.Warn("[Service] Stream: Received Empty Content from Service Provider - Drop it", "taskid", st.Schedule.Id, "content", content)
				} else {
					if isFirstTrunk {
						slog.Info("[Service] Stream: Convert Many Stream Response ...", "taskid", st.Schedule.Id, "from flavor", targetFlavor.Name(), "to flavor", requestFlavor.Name())
					}
					content, convertErr = ConvertBetweenFlavors(targetFlavor, requestFlavor, st.Request.Service, "stream_response", content, respConvertCtx)
					if convertErr != nil && !IsDropAction(convertErr) {
						slog.Error("[Service] Failed to convert response", "taskid", st.Schedule.Id, "from flavor", targetFlavor.Name(),
							"to flavor", requestFlavor.Name(), "error", err, "content", content)
						return fmt.Errorf("[Service] Failed to convert response: %s", convertErr.Error())
					}
				}
				if convertErr == nil { // not drop etc.
					// target stream mode maybe changed from service provider's
					if sendBackConvertedStreamMode == nil {
						sendBackConvertedStreamMode = NewStreamMode(content.Header) // got a most valid header to send back
					}
					content.Body = sendBackConvertedStreamMode.WrapChunk(content.Body)
					if isFirstTrunk { // send Wrapped prolog
						if len(prolog) > 0 {
							slog.Info("[Service] Stream: Send Prolog", "taskid", st.Schedule.Id, "prolog", prolog)
						}
						for i := len(prolog) - 1; i >= 0; i-- {
							st.Ch <- &ServiceResult{Type: ServiceResultChunk, TaskId: st.Schedule.Id,
								Error:      nil,
								StatusCode: 200,
								HTTP: HTTPContent{
									Body:   sendBackConvertedStreamMode.WrapChunk([]byte(prolog[i])),
									Header: sendBackConvertedStreamMode.Header,
								},
							}
						} // end for prolog
					} // end first trunk
				} // end conversion succeed
			} // end conversion
			isFirstTrunk = false

			if readChunkErr == io.EOF {
				if conversionNeeded {
					if len(epilog) > 0 {
						slog.Info("[Service] Stream: Send Epilog", "taskid", st.Schedule.Id, "epilog", epilog)
					}
					for _, v := range epilog {
						st.Ch <- &ServiceResult{Type: ServiceResultChunk, TaskId: st.Schedule.Id,
							Error:      nil,
							StatusCode: 200,
							HTTP: HTTPContent{
								Body:   sendBackConvertedStreamMode.WrapChunk([]byte(v)),
								Header: sendBackConvertedStreamMode.Header,
							},
						}
					} // end for epilog
				} // end conversion
				st.Ch <- &ServiceResult{Type: ServiceResultDone, TaskId: st.Schedule.Id,
					Error:      convertErr, // send back add / drop action etc.
					StatusCode: resp.StatusCode,
					HTTP:       content}
				return nil
			} else {
				st.Ch <- &ServiceResult{Type: ServiceResultChunk, TaskId: st.Schedule.Id,
					Error:      convertErr, // send back add / drop action etc.
					StatusCode: resp.StatusCode,
					HTTP:       content}
			}
		}
	}

	return nil
}

func InvokeService(fromFlavor string, service string, request *http.Request) (uint64, chan *ServiceResult, error) {
	slog.Info("[Service] Invoking Service", "fromFlavor", fromFlavor, "service", service)

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return 0, nil, err
	}

	SysEvents.NotifyHTTPRequest("receive_service_request", request.Method,
		request.URL.String(), request.Header, body)

	if request.Method == http.MethodGet {
		queryParams := request.URL.Query()
		queryParamsJSON, err := json.Marshal(queryParams)
		if err != nil {
			slog.Error("[Service] Failed to unmarshal GET request", "error", err, "body", string(body))
			return 0, nil, err
		}
		slog.Debug("[Service] GET Request Query Params", "params", string(queryParamsJSON))

		body = queryParamsJSON
	} // TODO: handle the case that the body is not json and not text
	if request.Method == http.MethodPost &&
		!strings.Contains(request.Header.Get("Content-Type"), "application/json") &&
		!strings.Contains(request.Header.Get("Content-Type"), "text/plain") {
		panic("TO SUPPORT non JSON or non text request")
	}

	serviceRequest := ServiceRequest{
		FromFlavor:      fromFlavor,
		Service:         service,
		Priority:        0,
		HTTP:            HTTPContent{Body: body, Header: request.Header},
		OriginalRequest: request,
	}

	err = json.Unmarshal(body, &serviceRequest)

	if err != nil {
		slog.Error("[Service] Failed to unmarshal POST request", "error", err, "body", string(body))
		return 0, nil, err
	}

	taskid, ch := GetScheduler().Enqueue(&serviceRequest)

	return taskid, ch, err
}
