package vega

import (
	"byze/internal/utils/client"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type VegaClient struct {
	client.Client
}

func NewVegaClient() *VegaClient {
	// default host
	host := "127.0.0.1:16677"

	// default scheme
	scheme := "http"

	return &VegaClient{
		Client: *client.NewClient(&url.URL{
			Scheme: scheme,
			Host:   host,
		}, http.DefaultClient),
	}
}

func QueryCloudModelJson(hybridPolicy string) error {
	c := NewVegaClient()
	routerPath := fmt.Sprintf("/vega/%s/service", "0.1")

	req := QueryCloudModelJsonRequest{
		HybridPolicy: hybridPolicy,
	}
	resp := QueryCloudModelJsonRespond{}
	err := c.Client.Do(context.Background(), http.MethodPost, routerPath, req, &resp)
	if err != nil {
		return err
	}
	if resp.HTTPCode > 200 {
		fmt.Println(resp.Message)
		return fmt.Errorf(resp.Message)
	}
	return nil
}
