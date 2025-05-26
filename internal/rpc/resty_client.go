package rpc

import (
	"github.com/go-resty/resty/v2"
	"time"
)

var GlobalClient *resty.Client

func init() {
	GlobalClient = resty.New().
		SetBaseURL("http://localhost:3000").
		SetTimeout(20*time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(10*time.Second).
		SetHeader("Content-Type", "application/json")
}
