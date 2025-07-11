package schedule

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"oadin/internal/convert"
	"oadin/internal/datastore"
	"oadin/internal/event"
	"oadin/internal/types"
)

type ServiceTask struct {
	Request  *types.ServiceRequest
	Target   *types.ServiceTarget
	Ch       chan *types.ServiceResult
	Error    error
	Schedule types.ScheduleDetails
}

func (st *ServiceTask) String() string {
	return fmt.Sprintf("ServiceTask{Id: %d, Request: %s, Target: %s}", st.Schedule.Id, st.Request, st.Target)
}

func NewStreamMode(header http.Header) *types.StreamMode {
	mode := types.StreamModeNonStream
	if contentType := header.Get("Content-Type"); contentType != "" {
		ct := strings.ToLower(contentType)
		if strings.Contains(ct, "text/event-stream") {
			mode = types.StreamModeEventStream
		} else if strings.Contains(ct, "application/x-ndjson") {
			mode = types.StreamModeNDJson
		}
	}
	return &types.StreamMode{Mode: mode, Header: header.Clone()}
}

func (st *ServiceTask) Run() error {
	if st.Target == nil || st.Target.ServiceProvider == nil {
		panic("[Service] ServiceTask is not dispatched before it goes to Run() " + st.String())
	}
	if st.Request.Model != "" && st.Target.Model != "" && st.Request.Model != st.Target.Model {
		slog.Warn("[Service] Model Mismatch", "mode_in_request", st.Request.Model,
			"model_to_use", st.Target.Model, "service_provider_id", st.Target.ServiceProvider.ProviderName,
			"taskid", st.Schedule.Id)
	}
	if st.Request.AskStreamMode && !st.Target.Stream {
		slog.Warn("[Service] Request asks for stream mode but it is not supported by the service provider",
			"service_provider_id", st.Target.ServiceProvider.ProviderName, "taskid", st.Schedule.Id)
	}
	// ------------------------------------------------------------------
	// 1. Get flavors and convert request if necessary
	// ------------------------------------------------------------------
	ds := datastore.GetDefaultDatastore()
	sp := &types.ServiceProvider{
		Flavor:        st.Target.ToFavor,
		ServiceSource: st.Target.Location,
		ServiceName:   st.Request.Service,
		Status:        1,
	}
	err := ds.Get(context.Background(), sp)
	if err != nil {
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
		requestCtx := convert.ConvertContext{"stream": st.Target.Stream}
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

	invokeURL := sp.URL
	serviceDefaultInfo := GetProviderServiceDefaultInfo(st.Target.ToFavor, st.Request.Service)
	if strings.ToUpper(sp.Method) == "GET" {
		// the body could be empty,
		// or it is GET with parameters, but the parameters should have been
		// marshaled in InvokeService() and maybe even converted above
		if len(content.Body) > 0 {
			queryParams := make(map[string][]string)
			err := json.Unmarshal(content.Body, &queryParams)
			if err != nil {
				slog.Error("[Service] Failed to unmarshal GET request", "taskid",
					st.Schedule.Id, "error", err, "body", string(content.Body))
				return err
			}
			u, err := url.Parse(sp.URL)
			if err != nil {
				slog.Error("Error parsing Service Provider's URL", "taskid",
					st.Schedule.Id, "sp.Url", sp.URL, "error", err)
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
	if sp.ExtraHeaders != "{}" {
		var extraHeader map[string]interface{}
		err := json.Unmarshal([]byte(sp.ExtraHeaders), &extraHeader)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return err
		}
		for k, v := range extraHeader {
			req.Header.Set(k, v.(string))
		}

	}
	// remote provider auth
	if sp.AuthType != types.AuthTypeNone {
		authParams := &AuthenticatorParams{
			Request:      req,
			ProviderInfo: sp,
			Content:      content,
		}
		authenticator := ChooseProviderAuthenticator(authParams)
		if authenticator == nil {
			return fmt.Errorf("[Service] Failed to choose authenticator")
		}
		err = authenticator.Authenticate()
		if err != nil {
			return err
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
		req.Method, "url", req.URL.String(), "header", fmt.Sprintf("%+v", req.Header), "body", nil)
	event.SysEvents.NotifyHTTPRequest("invoke_service_provider", req.Method, req.URL.String(), content.Header, nil)
	fmt.Println("[Service] Request Sending to Service Provider ...", "taskid", st.Schedule.Id, "method",
		req.Method, "url", req.URL.String(), "header", fmt.Sprintf("%+v", req.Header), "body", string(content.Body))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound && sp.Flavor == types.FlavorOllama {
			m := new(types.Model)
			m.ModelName = st.Target.Model
			err := ds.Get(context.Background(), m)
			if err == nil {
				m.Status = "downloading"
				_ = ds.Put(context.Background(), m)
			}
		}
		var sbody string
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			sbody = string(b)
		}
		slog.Warn("[Service] Service Provider returns Error", "taskid", st.Schedule.Id,
			"status_code", resp.StatusCode, "body", sbody)
		return &types.HTTPErrorResponse{
			StatusCode: resp.StatusCode,
			Header:     resp.Header.Clone(),
			Body:       b,
		}
	}
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()
		resp.Body = reader
		resp.Header.Set("Content-Encoding", "application/json")
	}
	if sp.Flavor == types.FlavorSmartVision && resp.Header.Get("Content-Type") == "application/json" {
		bodyData, _ := io.ReadAll(resp.Body)
		var respData map[string]interface{}
		err = json.Unmarshal(bodyData, &respData)
		if err != nil {
			return err
		}
		statusCode, ok := respData["status_code"].(float64)
		if ok {
			if statusCode != 200 {
				return &types.HTTPErrorResponse{
					StatusCode: 400,
					Header:     resp.Header.Clone(),
					Body:       bodyData,
				}
			}
		} else {
			return &types.HTTPErrorResponse{
				StatusCode: 400,
				Header:     resp.Header.Clone(),
				Body:       bodyData,
			}
		}
		resp.Body = io.NopCloser(bytes.NewReader(bodyData))
	}
	var body []byte
	// second request
	if serviceDefaultInfo.RequestSegments > 1 {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		type OutputData struct {
			TaskId     string `json:"task_id"`
			TaskStatus string `json:"task_status"`
		}
		type RespData struct {
			Output OutputData `json:"output"`
		}
		var submitRespData RespData
		err = json.Unmarshal(body, &submitRespData)
		if err != nil {
			return err
		}
		taskId := submitRespData.Output.TaskId
		for {
			GetResultURL := fmt.Sprintf("%s/%s", serviceDefaultInfo.RequestExtraUrl, taskId)
			GetTaskReq, err := http.NewRequest("GET", GetResultURL, nil)
			if err != nil {
				return err
			}
			getTaskAuthParams := AuthenticatorParams{
				Request:      GetTaskReq,
				ProviderInfo: sp,
			}
			getTaskAuthenticator := ChooseProviderAuthenticator(&getTaskAuthParams)
			err = getTaskAuthenticator.Authenticate()
			if err != nil {
				return err
			}
			resp, err = client.Do(GetTaskReq)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				var sbody string
				body, err = io.ReadAll(resp.Body)
				if err != nil {
					sbody = string(body)
				}
				slog.Warn("[Service] Service Provider returns Error", "taskid", st.Schedule.Id,
					"status_code", resp.StatusCode, "body", sbody)
				return &types.HTTPErrorResponse{
					StatusCode: resp.StatusCode,
					Header:     resp.Header.Clone(),
					Body:       body,
				}
			}
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			var getRespData RespData
			err = json.Unmarshal(body, &getRespData)
			if err != nil {
				return err
			}
			taskStatus := getRespData.Output.TaskStatus
			if taskStatus == "FAILED" || taskStatus == "SUCCEEDED" || taskStatus == "UNKNOWN" {
				newReader := bytes.NewReader(body)
				readCloser := io.NopCloser(newReader)
				resp.Body = readCloser
				break
			}
			time.Sleep(500 * time.Millisecond)
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
	respConvertCtx := convert.ConvertContext{"id": fmt.Sprintf("%d%d", rand.Uint64(), st.Schedule.Id)}

	if !respStreamMode.IsStream() {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("[Service] Failed to read response body", "taskid", st.Schedule.Id, "error", err.Error())
			return fmt.Errorf("[Service] Failed to read response body: %s", err.Error())
		}

		slog.Debug("[Service] Response Content (non-stream)", "taskid", st.Schedule.Id, "body", nil)
		event.SysEvents.NotifyHTTPResponse("service_provider_response", resp.StatusCode, resp.Header, nil)

		content = types.HTTPContent{Body: body, Header: resp.Header.Clone()}

		if conversionNeeded {
			content, err = ConvertBetweenFlavors(targetFlavor, requestFlavor, st.Request.Service, "response", content, respConvertCtx)
			if err != nil {
				slog.Error("[Service] Failed to convert response", "taskid", st.Schedule.Id, "from flavor", targetFlavor.Name(),
					"to flavor", requestFlavor.Name(), "error", err, "content", content)
				return fmt.Errorf("[Service] Failed to convert response: %s", err.Error())
			}
		}

		st.Ch <- &types.ServiceResult{
			Type: types.ServiceResultDone, TaskId: st.Schedule.Id,
			StatusCode: resp.StatusCode,
			HTTP:       content,
		}
	} else {
		isFirstTrunk := true
		reader := bufio.NewReader(resp.Body)
		prolog := requestFlavor.GetStreamResponseProlog(st.Request.Service)
		epilog := requestFlavor.GetStreamResponseEpilog(st.Request.Service)
		var sendBackConvertedStreamMode *types.StreamMode // only used if need conversion
		for {
			chunk, readChunkErr := respStreamMode.ReadChunk(reader)
			if readChunkErr != nil && readChunkErr != io.EOF { // real error
				slog.Error("[Service] Stream: Failed to read chunk", "taskid", st.Schedule.Id, "error", readChunkErr.Error())
				return readChunkErr
			}
			event.SysEvents.NotifyHTTPResponse("service_provider_response", resp.StatusCode, resp.Header, chunk)

			if readChunkErr == io.EOF {
				slog.Debug("[Service] Stream: Got EOF Response", "taskid", st.Schedule.Id, "chunk", string(chunk))
			} else {
				slog.Debug("[Service] Stream: Got Chunk Response", "taskid", st.Schedule.Id, "chunk", string(chunk))
			}
			// fmt.Println("[Service] Response Content", "taskid", st.Schedule.Id, "chunk", string(chunk))

			chunkStr := strings.TrimPrefix(string(chunk), "data:")
			chunk = []byte(chunkStr)
			content = types.HTTPContent{Body: chunk, Header: resp.Header.Clone()}
			var convertErr error
			if conversionNeeded { // need convert response
				content.Body = respStreamMode.UnwrapChunk(content.Body)
				// drop empty content
				if len(bytes.TrimSpace(chunk)) == 0 {
					convertErr = &types.DropAction{}
					slog.Warn("[Service] Stream: Received Empty Content from Service Provider - Drop it", "taskid", st.Schedule.Id, "content", content)
				} else {
					if isFirstTrunk {
						slog.Info("[Service] Stream: Convert Many Stream Response ...", "taskid", st.Schedule.Id, "from flavor", targetFlavor.Name(), "to flavor", requestFlavor.Name())
					}
					content, convertErr = ConvertBetweenFlavors(targetFlavor, requestFlavor, st.Request.Service, "stream_response", content, respConvertCtx)
					if convertErr != nil && !types.IsDropAction(convertErr) {
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
							st.Ch <- &types.ServiceResult{
								Type: types.ServiceResultChunk, TaskId: st.Schedule.Id,
								Error:      nil,
								StatusCode: 200,
								HTTP: types.HTTPContent{
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
						st.Ch <- &types.ServiceResult{
							Type: types.ServiceResultChunk, TaskId: st.Schedule.Id,
							Error:      nil,
							StatusCode: 200,
							HTTP: types.HTTPContent{
								Body:   sendBackConvertedStreamMode.WrapChunk([]byte(v)),
								Header: sendBackConvertedStreamMode.Header,
							},
						}
					} // end for epilog
				} // end conversion
				st.Ch <- &types.ServiceResult{
					Type: types.ServiceResultDone, TaskId: st.Schedule.Id,
					Error:      convertErr, // send back add / drop action etc.
					StatusCode: resp.StatusCode,
					HTTP:       content,
				}
				return nil
			} else {
				st.Ch <- &types.ServiceResult{
					Type: types.ServiceResultChunk, TaskId: st.Schedule.Id,
					Error:      convertErr, // send back add / drop action etc.
					StatusCode: resp.StatusCode,
					HTTP:       content,
				}
			}
		}
	}

	return nil
}
