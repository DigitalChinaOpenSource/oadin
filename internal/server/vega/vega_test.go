package vega

import (
	"testing"
)

func TestGetCloudModelJson(t *testing.T) {
	err := QueryCloudModelJson("test")
	if err != nil {
		t.Errorf("GetCloudModelJson(%s) = %s; want nil", "test", err)
	}
}
