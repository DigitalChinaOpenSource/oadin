package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"intel.com/aog/internal/utils/bcode"
	"intel.com/aog/internal/utils/progress"
)

type Client struct {
	base *url.URL
	http *http.Client
}

var ModelClientMap = make(map[string][]context.CancelFunc)

func checkError(resp *http.Response, body []byte) error {
	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	apiError := bcode.Bcode{
		HTTPCode: int32(resp.StatusCode),
	}

	err := json.Unmarshal(body, &apiError)
	if err != nil {
		// Use the full body as the message if we fail to decode a response.
		apiError.Message = string(body)
	}

	return &apiError
}

func NewClient(base *url.URL, http *http.Client) *Client {
	return &Client{
		base: base,
		http: http,
	}
}

func (c *Client) Do(ctx context.Context, method, path string, reqData, respData any) error {
	var reqBody io.Reader
	var data []byte
	var err error

	switch reqData := reqData.(type) {
	case io.Reader:
		// reqData is already an io.Reader
		reqBody = reqData
	case nil:
		// noop
	default:
		data, err = json.Marshal(reqData)
		if err != nil {
			return err
		}

		reqBody = bytes.NewReader(data)
	}

	requestURL := c.base.JoinPath(path)
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), reqBody)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	// request.Header.Set("User-Agent", fmt.Sprintf("ollama/%s (%s %s) Go/%s", version.Version, runtime.GOARCH, runtime.GOOS, runtime.Version()))

	respObj, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer respObj.Body.Close()

	respBody, err := io.ReadAll(respObj.Body)
	if err != nil {
		return err
	}

	if err := checkError(respObj, respBody); err != nil {
		return err
	}

	if len(respBody) > 0 && respData != nil {
		if err := json.Unmarshal(respBody, respData); err != nil {
			return err
		}
	}
	return nil
}

const maxBufferSize = 512 * progress.KiloByte

func (c *Client) Stream(ctx context.Context, method, path string, data any, fn func([]byte) error) error {
	var buf *bytes.Buffer
	if data != nil {
		bts, err := json.Marshal(data)
		if err != nil {
			return err
		}

		buf = bytes.NewBuffer(bts)
	}

	requestURL := c.base.JoinPath(path)
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), buf)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/x-ndjson")

	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	scanner := bufio.NewScanner(response.Body)
	// increase the buffer size to avoid running out of space
	scanBuf := make([]byte, 0, maxBufferSize)
	scanner.Buffer(scanBuf, maxBufferSize)
	for scanner.Scan() {
		var errorResponse struct {
			Error string `json:"error,omitempty"`
		}

		bts := scanner.Bytes()
		if err := json.Unmarshal(bts, &errorResponse); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		if errorResponse.Error != "" {
			return errors.New(errorResponse.Error)
		}

		if response.StatusCode >= http.StatusBadRequest {
			return errors.New(errorResponse.Error)
		}

		if err := fn(bts); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) StreamResponse(ctx context.Context, method, path string, reqData any) (chan []byte, chan error) {
	dataCh := make(chan []byte)
	errCh := make(chan error, 1) // Buffered channels avoid goroutine blocking

	go func() {
		defer close(dataCh)
		defer close(errCh)

		// Build the request body
		var reqBody io.Reader
		switch v := reqData.(type) {
		case io.Reader:
			reqBody = v
		case nil:
		default:
			data, err := json.Marshal(v)
			if err != nil {
				errCh <- fmt.Errorf("marshal request data failed: %w", err)
				return
			}
			reqBody = bytes.NewReader(data)
		}

		requestURL := c.base.JoinPath(path)
		request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), reqBody)
		if err != nil {
			errCh <- fmt.Errorf("create request failed: %w", err)
			return
		}

		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json") // Ollama通常返回JSON流

		resp, err := c.http.Do(request)
		if err != nil {
			errCh <- fmt.Errorf("execute request failed: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				line := scanner.Bytes()
				if len(line) == 0 {
					continue
				}

				chunk := make([]byte, len(line))
				copy(chunk, line)
				dataCh <- chunk
			}
		}

		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("reading response failed: %w", err)
		}
	}()

	return dataCh, errCh
}
