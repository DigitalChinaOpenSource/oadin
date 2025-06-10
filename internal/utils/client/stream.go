package client

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) StreamRequest(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	if !u.IsAbs() {
		u = c.base.ResolveReference(u)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")

	// 添加流式输出所需的请求头
	req.Header.Set("Cache-Control", "no-cache")

	// 发送请求
	return c.http.Do(req)
}
