package schedule

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"intel.com/aog/internal/client/grpc/grpc_client"
	"intel.com/aog/internal/convert"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/event"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
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

func HandleRequest(st *ServiceTask) error {

	if st.Request.Service == types.ServiceTextToImage {
		reqBody := st.Request.HTTP.Body
		var body map[string]interface{}
		err := json.Unmarshal(reqBody, &body)
		if err != nil {
			return err
		}
		imageType, typeOk := body["image_type"].(string)
		image, imageOk := body["image"].(string)
		if !typeOk && !imageOk {
			return nil
		} else if !typeOk && imageOk {
			return errors.New("image request param lost")
		} else if typeOk && !imageOk {
			return errors.New("image_type request param lost")
		}
		if !utils.Contains(types.SupportImageType, imageType) {
			return errors.New("unsupported image type")
		}
		if imageType == types.ImageTypePath && st.Target.Location == types.ServiceSourceRemote {
			imgData, err := os.ReadFile(image)
			if err != nil {
				return err
			}
			imgDataBase64Str := base64.StdEncoding.EncodeToString(imgData)
			body["image"] = imgDataBase64Str
		} else if imageType == types.ImageTypeUrl && st.Target.Location == types.ServiceSourceLocal {
			downLoadPath, err := utils.GetDownloadDir()
			if err != nil {
				return err
			}
			savePath, err := utils.DownloadFile(image, downLoadPath)
			if err != nil {
				return err
			}
			body["image"] = savePath
			// todo() Should the original images be deleted after using the local service?
			//os.Remove(savePath)
		}
		newReqBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		st.Request.HTTP.Body = newReqBody
		return nil
	}
	return nil
}

func (st *ServiceTask) Run() error {
	logger.LogicLogger.Debug("[Service] ServiceTask start run......")
	err := HandleRequest(st)
	if err != nil {
		return err
	}
	if st.Target == nil || st.Target.ServiceProvider == nil {
		panic("[Service] ServiceTask is not dispatched before it goes to Run() " + st.String())
	}
	if st.Request.Model != "" && st.Target.Model != "" && st.Request.Model != st.Target.Model {
		logger.LogicLogger.Warn("[Service] Model Mismatch", "mode_in_request", st.Request.Model,
			"model_to_use", st.Target.Model, "service_provider_id", st.Target.ServiceProvider.ProviderName,
			"taskid", st.Schedule.Id)
	}
	if st.Request.AskStreamMode && !st.Target.Stream {
		logger.LogicLogger.Warn("[Service] Request asks for stream mode but it is not supported by the service provider",
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
	err = ds.Get(context.Background(), sp)
	if err != nil {
		return fmt.Errorf("service Provider not found for %s of Service %s", st.Target.Location, st.Request.Service)
	}
	requestFlavor, err := GetAPIFlavor(st.Request.FromFlavor)
	if err != nil {
		logger.LogicLogger.Error("[Service] Unsupported API Flavor for Request", "task", st, "error", err)
		return fmt.Errorf("[Service] Unsupported API Flavor %s for Request: %s", st.Request.FromFlavor, err.Error())
	}
	targetFlavor, err := GetAPIFlavor(st.Target.ServiceProvider.Flavor)
	if err != nil {
		logger.LogicLogger.Error("[Service] Unsupported API Flavor for Service Provider", "task", st, "error", err)
		return fmt.Errorf("[Service] Unsupported API Flavor %s for Service Provider: %s", st.Target.ServiceProvider.Flavor, err.Error())
	}

	conversionNeeded := targetFlavor.Name() != requestFlavor.Name()
	// GRPC Body
	content := st.Request.HTTP

	// todo Here, the converter of grpc needs to be implemented later.
	logger.LogicLogger.Debug("[Service] ServiceTask conversion......")
	if conversionNeeded {
		logger.LogicLogger.Info("[Service] Converting Request", "taskid", st.Schedule.Id, "from flavor", requestFlavor.Name(), "to flavor", targetFlavor.Name())
		requestCtx := convert.ConvertContext{"stream": st.Target.Stream}
		if st.Target.Model != "" {
			requestCtx["model"] = st.Target.Model
		}

		var err error
		content, err = ConvertBetweenFlavors(requestFlavor, targetFlavor, st.Request.Service, "request", content, requestCtx)
		if err != nil {
			logger.LogicLogger.Error("[Service] Failed to convert request", "taskid", st.Schedule.Id, "from flavor", requestFlavor.Name(),
				"to flavor", targetFlavor.Name(), "error", err, "content", content)
			return fmt.Errorf("[Service] Failed to convert request: %s", err.Error())
		}
	}

	// ------------------------------------------------------------------
	// 2. Invoke the service provider and get response
	// ------------------------------------------------------------------

	resp := &http.Response{}
	if targetFlavor.Name() == types.FlavorOpenvino {
		resp, err = st.invokeGRPCServiceProvider(sp, content)
	} else {
		resp, err = st.invokeHTTPServiceProvider(sp, content)
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		logger.LogicLogger.Error("[Service] Failed to invoke service provider", "taskid", st.Schedule.Id, "error", err.Error())
		return fmt.Errorf("[Service] Failed to invoke service provider: %s", err.Error())
	}

	// ------------------------------------------------------------------
	// 3. Convert response if necessary and send back to handler
	// ------------------------------------------------------------------
	respStreamMode := NewStreamMode(resp.Header)

	logger.LogicLogger.Debug("[Service] Response is Stream?", "taskid", st.Schedule.Id, "stream", respStreamMode.Mode.String())

	// in case response to send out needs a id but not in response returned from service provider
	respConvertCtx := convert.ConvertContext{"id": fmt.Sprintf("%d%d", rand.Uint64(), st.Schedule.Id)}

	if !respStreamMode.IsStream() {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.LogicLogger.Error("[Service] Failed to read response body", "taskid", st.Schedule.Id, "error", err.Error())
			return fmt.Errorf("[Service] Failed to read response body: %s", err.Error())
		}

		logger.LogicLogger.Debug("[Service] Response Content (non-stream)", "taskid", st.Schedule.Id, "body", nil)
		event.SysEvents.NotifyHTTPResponse("service_provider_response", resp.StatusCode, resp.Header, nil)

		content = types.HTTPContent{Body: body, Header: resp.Header.Clone()}

		if conversionNeeded {
			content, err = ConvertBetweenFlavors(targetFlavor, requestFlavor, st.Request.Service, "response", content, respConvertCtx)
			if err != nil {
				logger.LogicLogger.Error("[Service] Failed to convert response", "taskid", st.Schedule.Id, "from flavor", targetFlavor.Name(),
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
				logger.LogicLogger.Error("[Service] Stream: Failed to read chunk", "taskid", st.Schedule.Id, "error", readChunkErr.Error())
				return readChunkErr
			}
			event.SysEvents.NotifyHTTPResponse("service_provider_response", resp.StatusCode, resp.Header, chunk)

			if readChunkErr == io.EOF {
				logger.LogicLogger.Debug("[Service] Stream: Got EOF Response", "taskid", st.Schedule.Id, "chunk", string(chunk))
			} else {
				logger.LogicLogger.Debug("[Service] Stream: Got Chunk Response", "taskid", st.Schedule.Id, "chunk", string(chunk))
			}

			content = types.HTTPContent{Body: chunk, Header: resp.Header.Clone()}
			var convertErr error
			if conversionNeeded { // need convert response
				content.Body = respStreamMode.UnwrapChunk(content.Body)
				// drop empty content
				if len(bytes.TrimSpace(chunk)) == 0 {
					convertErr = &types.DropAction{}
					logger.LogicLogger.Warn("[Service] Stream: Received Empty Content from Service Provider - Drop it", "taskid", st.Schedule.Id, "content", content)
				} else {
					if isFirstTrunk {
						logger.LogicLogger.Info("[Service] Stream: Convert Many Stream Response ...", "taskid", st.Schedule.Id, "from flavor", targetFlavor.Name(), "to flavor", requestFlavor.Name())
					}
					content, convertErr = ConvertBetweenFlavors(targetFlavor, requestFlavor, st.Request.Service, "stream_response", content, respConvertCtx)
					if convertErr != nil && !types.IsDropAction(convertErr) {
						logger.LogicLogger.Error("[Service] Failed to convert response", "taskid", st.Schedule.Id, "from flavor", targetFlavor.Name(),
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
							logger.LogicLogger.Info("[Service] Stream: Send Prolog", "taskid", st.Schedule.Id, "prolog", prolog)
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
						logger.LogicLogger.Info("[Service] Stream: Send Epilog", "taskid", st.Schedule.Id, "epilog", epilog)
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

func (st *ServiceTask) invokeGRPCServiceProvider(sp *types.ServiceProvider, content types.HTTPContent) (resp *http.Response, err error) {
	invokeURL := sp.URL
	resp = &http.Response{}

	if sp.ServiceName != types.ServiceTextToImage {
		return nil, fmt.Errorf("currently only support text to image service")
	}

	conn, err := grpc.Dial(invokeURL, grpc.WithInsecure())
	if err != nil {
		logger.LogicLogger.Error("Couldn't connect to endpoint %s: %v", invokeURL, err)
	}
	defer conn.Close()

	client := grpc_client.NewGRPCInferenceServiceClient(conn)

	switch sp.ServiceName {
	case types.ServiceTextToImage:
		var requestMap map[string]interface{}
		err := json.Unmarshal(content.Body, &requestMap)
		if err != nil {
			logger.LogicLogger.Error("[Service] Failed to unmarshal request body", "taskid", st.Schedule.Id, "error", err)
			return nil, err
		}
		prompt, ok := requestMap["prompt"].(string)
		if !ok {
			logger.LogicLogger.Error("[Service] Failed to get prompt from request body", "taskid", st.Schedule.Id)
			return nil, fmt.Errorf("failed to get prompt from request body")
		}
		batch, ok := requestMap["batch"].(float64)
		if !ok {
			batch = float64(1)
		}
		height := 1024
		width := 1024
		size, ok := requestMap["size"].(string)
		if ok {
			sizeStr := strings.Split(size, "x")
			num, err := strconv.Atoi(sizeStr[0])
			if err != nil {
				logger.LogicLogger.Error("[Service] Failed to parse size from request body", "taskid", st.Schedule.Id, "error", err)
				return nil, err
			}
			height = num

			num, err = strconv.Atoi(sizeStr[1])
			if err != nil {
				logger.LogicLogger.Error("[Service] Failed to parse size from request body", "taskid", st.Schedule.Id, "error", err)
				return nil, err
			}
			width = num
		}

		promptBytes := []byte(prompt)
		rawContents := make([][]byte, 0) // ovms 实际接收值
		rawContents = append(rawContents, promptBytes)
		rawContents = append(rawContents, []byte(fmt.Sprintf("%d", int(batch))))
		rawContents = append(rawContents, []byte(strconv.Itoa(height)))
		rawContents = append(rawContents, []byte(strconv.Itoa(width)))

		inferTensorInputs := make([]*grpc_client.ModelInferRequest_InferInputTensor, 0)
		inferTensorInputs = append(inferTensorInputs, &grpc_client.ModelInferRequest_InferInputTensor{
			Name:     "prompt",
			Datatype: "BYTES",
			Shape:    []int64{1},
		}, &grpc_client.ModelInferRequest_InferInputTensor{
			Name:     "batch",
			Datatype: "BYTES",
			Shape:    []int64{1},
		}, &grpc_client.ModelInferRequest_InferInputTensor{
			Name:     "height",
			Datatype: "BYTES",
		}, &grpc_client.ModelInferRequest_InferInputTensor{
			Name:     "width",
			Datatype: "BYTES",
		})

		inferOutputs := []*grpc_client.ModelInferRequest_InferRequestedOutputTensor{
			{
				Name: "image",
			},
		}

		grpcReq := &grpc_client.ModelInferRequest{
			ModelName:        st.Target.Model,
			Inputs:           inferTensorInputs,
			Outputs:          inferOutputs,
			RawInputContents: rawContents,
		}

		inferResponse, err := client.ModelInfer(context.Background(), grpcReq)
		if err != nil {
			logger.LogicLogger.Error("[Service] Error processing InferRequest", "taskid", st.Schedule.Id, "error", err)
			return nil, err
		}

		imageList, err := utils.ParseImageData(inferResponse.RawOutputContents[0])
		if err != nil {
			logger.LogicLogger.Error("[Service] Failed to parse image data", "taskid", st.Schedule.Id, "error", err)
			return nil, err
		}

		outputList := make([]string, 0)
		fmt.Println("InferResponse:", len(inferResponse.RawOutputContents))
		for i, imageData := range imageList {
			now := time.Now()
			randNum := rand.Intn(10000)
			DownloadPath, _ := utils.GetDownloadDir()
			imageName := fmt.Sprintf("%d%02d%02d%02d%02d%02d%04d%01d.png", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), randNum, i)
			imagePath := fmt.Sprintf("%s/%s", DownloadPath, imageName)
			err = os.WriteFile(imagePath, imageData, 0o644)
			if err != nil {
				logger.LogicLogger.Error("[Service] Failed to write image file", "taskid", st.Schedule.Id, "error", err)
				continue
			}

			outputList = append(outputList, imagePath)
		}
		respHeader := make(http.Header)
		respHeader.Set("Content-Type", "application/json")
		resp.Header = respHeader

		respBody := map[string]interface{}{
			"local_path": outputList,
		}
		respBodyBytes, err := json.Marshal(respBody)
		if err != nil {
			logger.LogicLogger.Error("[Service] Failed to marshal response body", "taskid", st.Schedule.Id, "error", err)
			return nil, err
		}

		resp.Body = io.NopCloser(bytes.NewReader(respBodyBytes))
	}

	logger.LogicLogger.Debug("[Service] Response Receiving", "taskid", st.Schedule.Id, "header",
		fmt.Sprintf("%+v", resp.Header), "task", st)

	return resp, nil
}

func (st *ServiceTask) ConvertFlavor(requestFlavor APIFlavor, targetFlavor APIFlavor, content types.HTTPContent) (grpc_client.ModelInferRequest, error) {
	logger.LogicLogger.Info("[Service] Converting Request", "taskid", st.Schedule.Id, "from flavor", requestFlavor.Name(), "to flavor", targetFlavor.Name())
	requestCtx := convert.ConvertContext{"stream": st.Target.Stream}
	if st.Target.Model != "" {
		requestCtx["model"] = st.Target.Model
	}

	var err error
	content, err = ConvertBetweenFlavors(requestFlavor, targetFlavor, st.Request.Service, "request", content, requestCtx)
	if err != nil {
		logger.LogicLogger.Error("[Service] Failed to convert request", "taskid", st.Schedule.Id, "from flavor", requestFlavor.Name(),
			"to flavor", targetFlavor.Name(), "error", err, "content", content)
		return grpc_client.ModelInferRequest{}, fmt.Errorf("[Service] Failed to convert request: %s", err.Error())
	}

	return grpc_client.ModelInferRequest{}, nil
}

func (st *ServiceTask) invokeHTTPServiceProvider(sp *types.ServiceProvider, content types.HTTPContent) (*http.Response, error) {
	// ------------------------------------------------------------------
	// 1. Invoke the service provider
	// ------------------------------------------------------------------
	invokeURL := sp.URL
	resp := &http.Response{}
	serviceDefaultInfo := GetProviderServiceDefaultInfo(st.Target.ToFavor, st.Request.Service)
	if strings.ToUpper(sp.Method) == "GET" {
		// the body could be empty,
		// or it is GET with parameters, but the parameters should have been
		// marshaled in InvokeService() and maybe even converted above
		if len(content.Body) > 0 {
			queryParams := make(map[string][]string)
			err := json.Unmarshal(content.Body, &queryParams)
			if err != nil {
				logger.LogicLogger.Error("[Service] Failed to unmarshal GET request", "taskid",
					st.Schedule.Id, "error", err, "body", string(content.Body))
				return nil, err
			}
			u, err := url.Parse(sp.URL)
			if err != nil {
				logger.LogicLogger.Error("Error parsing Service Provider's URL", "taskid",
					st.Schedule.Id, "sp.Url", sp.URL, "error", err)
				return nil, err
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
		return nil, err
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
			logger.LogicLogger.Error("Error parsing JSON:", err)
			return nil, err
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
			RequestBody:  string(content.Body),
		}
		authenticator := ChooseProviderAuthenticator(authParams)
		if authenticator == nil {
			logger.LogicLogger.Error("[Service] Failed to choose authenticator")
			return nil, fmt.Errorf("[Service] Failed to choose authenticator")
		}
		err = authenticator.Authenticate()
		if err != nil {
			logger.LogicLogger.Error("[Service] Failed to authenticate", "taskid", st.Schedule.Id, "error", err)
			return nil, err
		}
	}
	// TODO: further fine tuning of the transport
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: transport}
	logger.LogicLogger.Info("[Service] Request Sending to Service Provider ...", "taskid", st.Schedule.Id, "url", req.URL.String())
	logger.LogicLogger.Debug("[Service] Request Sending to Service Provider ...", "taskid", st.Schedule.Id, "method",
		req.Method, "url", req.URL.String(), "header", fmt.Sprintf("%+v", req.Header), "body", string(content.Body))
	event.SysEvents.NotifyHTTPRequest("invoke_service_provider", req.Method, req.URL.String(), content.Header, content.Body)
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var sbody string
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			sbody = string(b)
		}
		logger.LogicLogger.Warn("[Service] Service Provider returns Error", "taskid", st.Schedule.Id,
			"status_code", resp.StatusCode, "body", sbody)
		resp.Body.Close()
		return resp, errors.New("[Service] Service Provider API returns Error err: \n" + sbody)
	}
	var body []byte
	// second request
	if serviceDefaultInfo.RequestSegments > 1 {
		var reader io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			defer reader.Close()
		default:
			reader = resp.Body
		}
		body, err = io.ReadAll(reader)
		if err != nil {
			resp.Body.Close()
			return nil, err
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
			resp.Body.Close()
			return nil, err
		}
		taskId := submitRespData.Output.TaskId
		for {
			GetResultURL := fmt.Sprintf("%s/%s", serviceDefaultInfo.RequestExtraUrl, taskId)
			GetTaskReq, err := http.NewRequest("GET", GetResultURL, nil)
			if err != nil {
				resp.Body.Close()
				return nil, err
			}
			getTaskAuthParams := AuthenticatorParams{
				Request:      GetTaskReq,
				ProviderInfo: sp,
			}
			getTaskAuthenticator := ChooseProviderAuthenticator(&getTaskAuthParams)
			err = getTaskAuthenticator.Authenticate()
			if err != nil {
				resp.Body.Close()
				return nil, err
			}
			resp, err = client.Do(GetTaskReq)
			if err != nil {
				resp.Body.Close()
				return nil, err
			}
			if resp.StatusCode != http.StatusOK {
				var sbody string
				body, err = io.ReadAll(resp.Body)
				if err != nil {
					sbody = string(body)
				}
				logger.LogicLogger.Warn("[Service] Service Provider returns Error", "taskid", st.Schedule.Id,
					"status_code", resp.StatusCode, "body", sbody)
				resp.Body.Close()
				return nil, &types.HTTPErrorResponse{
					StatusCode: resp.StatusCode,
					Header:     resp.Header.Clone(),
					Body:       body,
				}
			}
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				resp.Body.Close()
				return nil, err
			}
			var getRespData RespData
			err = json.Unmarshal(body, &getRespData)
			if err != nil {
				resp.Body.Close()
				return nil, err
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

	logger.LogicLogger.Debug("[Service] Response Receiving", "taskid", st.Schedule.Id, "header",
		fmt.Sprintf("%+v", resp.Header), "task", st)

	return resp, nil
}
