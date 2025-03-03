// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package opac

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Id is something unique in global
// Name is something local, can be duplicated in global
type ServiceProviderInfo struct {
	Id            string                    `json:"-"`
	Desc          string                    `json:"desc"`
	Method        string                    `json:"method"`
	Url           string                    `json:"url"`
	Flavor        string                    `json:"api_flavor"`
	ExtraHeaders  map[string]string         `json:"extra_headers"`
	ExtraJsonBody map[string]interface{}    `json:"extra_json_body"`
	Properties    ServiceProviderProperties `json:"properties"`
}

func (sp *ServiceProviderInfo) String() string {
	return fmt.Sprintf("{Id: %s, Flavor: %s, Endpoint: %s}", sp.Id, sp.Flavor, sp.Method+" "+sp.Url)
}

type ServiceProviderProperties struct {
	MaxInputTokens        int      `json:"max_input_tokens"`
	SupportedResponseMode []string `json:"supported_response_mode"`
	ModeIsChangable       bool     `json:"mode_is_changable"`
	Models                []string `json:"models"`
	XPU                   []string `json:"xpu"`
}

type ServiceInfo struct {
	Id               string            `json:"-"`
	ServiceProviders map[string]string `json:"service_providers"`
	HybridPolicy     string            `json:"hybrid_policy"`
}

type PlatformInfo struct {
	APIMajorVersion    string                         `json:"version"`
	ServiceProviderIds map[string]ServiceProviderInfo `json:"service_providers"`
	Services           map[string]ServiceInfo         `json:"services"`
}

func (p *PlatformInfo) GetServiceInfo(id string) *ServiceInfo {
	if s, exists := p.Services[id]; exists {
		return &s
	}
	return nil
}

func (p *PlatformInfo) GetServiceProviderInfoById(s string) *ServiceProviderInfo {
	if sp, exists := p.ServiceProviderIds[s]; exists {
		return &sp
	}
	return nil
}

func (p *PlatformInfo) GetServiceProviderInfo(serviceId string, location string) *ServiceProviderInfo {
	service := p.GetServiceInfo(serviceId)
	if service == nil {
		return nil
	}
	if sp, exists := service.ServiceProviders[location]; exists {
		return p.GetServiceProviderInfoById(sp)
	}
	return nil
}

var platform *PlatformInfo

func LoadPlatformInfo(f string) error {
	var Platform PlatformInfo
	// Open the JSON file
	slog.Info("[Platform] Loading platform configuration from", "file", f)
	jsonFile, err := os.Open(f)
	if err != nil {
		return fmt.Errorf("error opening opac configuration JSON file! %s", err.Error())
	}
	defer jsonFile.Close()

	// Read the JSON file into a byte array
	byteValue, _ := io.ReadAll(jsonFile)

	// Unmarshal the byte array into the Config struct
	err = json.Unmarshal(byteValue, &Platform)
	if err != nil {
		return fmt.Errorf("error unmarshalling opac configuration JSON file! %s", err.Error())
	}

	for id, sp := range Platform.ServiceProviderIds {
		sp.Id = id
		Platform.ServiceProviderIds[id] = sp
		if sp.Flavor == "" {
			return fmt.Errorf("service provider %s has no api_flavor provider", id)
		}
		sp.Method = strings.ToUpper(sp.Method)
		if sp.Method != "POST" && sp.Method != "GET" && sp.Method != "PUT" && sp.Method != "DELETE" {
			return fmt.Errorf("service provider %s has no or incorrect HTTP Method %s provided", id, sp.Method)
		}
		if sp.Url == "" {
			return fmt.Errorf("service provider %s has no URL provided", id)
		}
		slog.Debug("[Platform] Service Provider", "id", id, "flavor", sp.Flavor, "URL", sp.Url)
	}

	for id, s := range Platform.Services {
		s.Id = id
		for k, v := range s.ServiceProviders {
			if _, exists := Platform.ServiceProviderIds[v]; !exists {
				return fmt.Errorf("service provider ID %s not found for %s of Service %s", v, k, id)
			}
		}
		Platform.Services[id] = s
		slog.Debug("[Platform] Service", "id", id, "Providers", s.ServiceProviders)
	}

	platform = &Platform

	return nil
}

func GetPlatformInfo() *PlatformInfo {
	if GetEnv().ForceReload {
		err := LoadPlatformInfo(GetEnv().ConfigFile)
		if err != nil {
			slog.Error("[Platform] Failed to reload platform configuration", "error", err.Error())
			// such error cannot recover
			panic("[Platform] Failed to reload platform configuration: " + err.Error())
		}
	}
	return platform
}
