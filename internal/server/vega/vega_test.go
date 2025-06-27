package vega

import (
	"context"
	"encoding/json"
	"oadin/internal/api/dto"
	"oadin/internal/provider/template"
	"reflect"
	"testing"
)

func TestQueryCloudModelJsonLocal(t *testing.T) {
	res, err := GetModels(context.Background(), "local")
	if err != nil {
		t.Errorf("GetModels() error = %v", err)
		return
	}

	localOllamaServiceMap := make(map[string][]dto.LocalSupportModelData)
	fileContent, err := template.FlavorTemplateFs.ReadFile("local_model.json")
	if err != nil {
		t.Errorf("Read file failed: %v\n", err)
	}
	// parse struct
	err = json.Unmarshal(fileContent, &localOllamaServiceMap)
	if err != nil {
		t.Errorf("Parse JSON failed: %v\n", err)
	}
	if len(localOllamaServiceMap) == 0 {
		t.Errorf("Parse JSON failed: %v\n", err)
	}
	t.Logf("Parse JSON success: %v\n", localOllamaServiceMap)

	for k, v := range localOllamaServiceMap {
		for i := 0; i < len(v); i++ {
			if v[i].OllamaId != res[k][i].OllamaId {
				t.Errorf("Model OllamaId not match: %s, %s\n", v[i].OllamaId, res[k][i].OllamaId)
			} else if v[i].Name != res[k][i].Name {
				t.Errorf("Model Name not match: %s, %s\n", v[i].Name, res[k][i].Name)
			} else if v[i].Avatar != res[k][i].Avatar {
				t.Errorf("Model Avatar not match: %s, %s\n", v[i].Avatar, res[k][i].Avatar)
			} else if v[i].Description != res[k][i].Description {
				t.Errorf("Model Description not match: %s, %s\n", v[i].Description, res[k][i].Description)
			} else if v[i].Flavor != res[k][i].Flavor {
				t.Errorf("Model Flavor not match: %s, %s\n", v[i].Flavor, res[k][i].Flavor)
			} else if v[i].Size != res[k][i].Size {
				t.Errorf("Model Size not match: %s, %s\n", v[i].Size, res[k][i].Size)
			} else if v[i].ParamsSize != res[k][i].ParamsSize {
				t.Errorf("Model ParamsSize not match: %f, %f\n", v[i].ParamsSize, res[k][i].ParamsSize)
			} else {
				t.Logf("Model %s match\n", v[i].Name)
			}
		}
	}

	if reflect.DeepEqual(localOllamaServiceMap, res) {
		t.Logf("TestQueryCloudModelJson() success")
	} else {
		t.Errorf("TestQueryCloudModelJson() failed")
	}
}

func TestQueryCloudModelJsonRemote(t *testing.T) {
	res, err := GetModels(context.Background(), "remote")
	if err != nil {
		t.Errorf("GetModels() error = %v", err)
		return
	}

	localOllamaServiceMap := make(map[string][]dto.LocalSupportModelData)
	fileContent, err := template.FlavorTemplateFs.ReadFile("remote_model.json")
	if err != nil {
		t.Errorf("Read file failed: %v\n", err)
	}
	// parse struct
	err = json.Unmarshal(fileContent, &localOllamaServiceMap)
	if err != nil {
		t.Errorf("Parse JSON failed: %v\n", err)
	}
	if len(localOllamaServiceMap) == 0 {
		t.Errorf("Parse JSON failed: %v\n", err)
	}
	t.Logf("Parse JSON success: %v\n", localOllamaServiceMap)

	for k, v := range localOllamaServiceMap {
		for i := 0; i < len(v); i++ {
			if v[i].OllamaId != res[k][i].OllamaId {
				t.Errorf("Model OllamaId not match: %s, %s\n", v[i].OllamaId, res[k][i].OllamaId)
			} else if v[i].Name != res[k][i].Name {
				t.Errorf("Model Name not match: %s, %s\n", v[i].Name, res[k][i].Name)
			} else if v[i].Avatar != res[k][i].Avatar {
				t.Errorf("Model Avatar not match: %s, %s\n", v[i].Avatar, res[k][i].Avatar)
			} else if v[i].Description != res[k][i].Description {
				t.Errorf("Model Description not match: %s, %s\n", v[i].Description, res[k][i].Description)
			} else if v[i].Flavor != res[k][i].Flavor {
				t.Errorf("Model Flavor not match: %s, %s\n", v[i].Flavor, res[k][i].Flavor)
			} else if v[i].Size != res[k][i].Size {
				t.Errorf("Model Size not match: %s, %s\n", v[i].Size, res[k][i].Size)
			} else if v[i].ParamsSize != res[k][i].ParamsSize {
				t.Errorf("Model ParamsSize not match: %f, %f\n", v[i].ParamsSize, res[k][i].ParamsSize)
			} else {
				t.Logf("Model %s match\n", v[i].Name)
			}
		}
	}

	if reflect.DeepEqual(localOllamaServiceMap, res) {
		t.Logf("TestQueryCloudModelJson() success")
	} else {
		t.Errorf("TestQueryCloudModelJson() failed")
	}
}

func TestQueryCloudSupplierJsonAll(t *testing.T) {
	res, err := QueryCloudSupplierJson(context.Background(), 1, 1000)
	if err != nil {
		t.Errorf("GetModels() error = %v", err)
		return
	}
	s, err := GetSuppliers(context.Background(), res)
	if err != nil {
		t.Errorf("GetModels() error = %v", err)
		return
	}
	t.Logf("QueryCloudSupplierJson() success: %v", s)
}
