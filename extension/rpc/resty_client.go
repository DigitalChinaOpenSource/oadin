package rpc

import (
	"time"

	"github.com/go-resty/resty/v2"
	"oadin/config"
)

var GlobalClient *resty.Client

func init() {
	GlobalClient = resty.New().
		SetBaseURL(config.ConfigRootInstance.Vega.Url).
		SetTimeout(20*time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(10*time.Second).
		SetHeader("Content-Type", "application/json")
}
