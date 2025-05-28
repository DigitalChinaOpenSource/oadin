package vega

import (
	"context"
	"testing"
)

func TestQueryCloudModelJson(t *testing.T) {
	data, err := QueryCloudModelJson(context.Background(), "remote", 1, 5)
	if err != nil {
		t.Errorf("QueryCloudModelJson() error = %v", err)
		return
	}
	if len(data) == 0 {
		t.Errorf("QueryCloudModelJson() got = %v, want not empty", data)
	}

	res, err := GetRemoteModels(data)
	if err != nil {
		t.Errorf("GetRemoteModels() error = %v", err)
		return
	}
	if len(res) == 0 {
		t.Errorf("GetRemoteModels() got = %v, want not empty", res)
	}
	t.Logf("GetRemoteModels() got = %v", res)
}
