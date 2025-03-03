// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package opac

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// mode is usually set to "default". And set to "stream" if it is using stream mode
type APIFlavor interface {
	Name() string
	InstallRoutes(*gin.Engine)

	// In stream mdoe, some flavor may ask for some packets to be send first
	// or at the end, in addition to normal contents. For example, OpenAI
	// needs to send an additional "data: [DONE]" after everything is done.
	GetStreamResponseProlog(service string) []string
	GetStreamResponseEpilog(service string) []string

	// This should cover the 6 conversion methods below
	Convert(service string, conversion string, content HTTPContent, ctx ConvertContext) (HTTPContent, error)

	ConvertRequestToOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error)
	ConvertRequestFromOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error)
	ConvertResponseToOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error)
	ConvertResponseFromOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error)
	ConvertStreamResponseToOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error)
	ConvertStreamResponseFromOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error)
}

var allFlavors map[string]APIFlavor = make(map[string]APIFlavor)

func RegisterAPIFlavor(f APIFlavor) {
	allFlavors[f.Name()] = f
}

func AllAPIFlavors() map[string]APIFlavor {
	return allFlavors
}

func GetAPIFlavor(name string) (APIFlavor, error) {
	flavor, ok := allFlavors[name]
	if !ok {
		return nil, fmt.Errorf("[Flavor] API Flavor %s not found", name)
	}
	return flavor, nil
}

//------------------------------------------------------------

// NOTE: we use YAML insteald of JSON here because it's easier to read and write
// In particular, it supporrts multiline strings which greatly help write
// jsonata templates
type ConversionStepDef struct {
	Converter string `yaml:"converter"`
	Config    any    `yaml:"config"`
}

type FlavorConversionDef struct {
	Prologue   []string            `yaml:"prologue"`
	Epilogue   []string            `yaml:"epilogue"`
	Conversion []ConversionStepDef `yaml:"conversion"`
}

type ModelSelector struct {
	ModelInRequest  string `yaml:"request"`
	ModelInResponse string `yaml:"response"`
}
type FlavorServiceDef struct {
	Endpoints              []string            `yaml:"endpoints"`
	InstallRawRoutes       bool                `yaml:"install_raw_routes"`
	ModelSelector          ModelSelector       `yaml:"model_selector"`
	RequestToOPAC          FlavorConversionDef `yaml:"request_to_opac"`
	RequestFromOPAC        FlavorConversionDef `yaml:"request_from_opac"`
	ResponseToOPAC         FlavorConversionDef `yaml:"response_to_opac"`
	ResponseFromOPAC       FlavorConversionDef `yaml:"response_from_opac"`
	StreamResponseToOPAC   FlavorConversionDef `yaml:"stream_response_to_opac"`
	StreamResponseFromOPAC FlavorConversionDef `yaml:"stream_response_from_opac"`
}

type FlavorDef struct {
	Version  string                      `yaml:"version"`
	Name     string                      `yaml:"name"`
	Services map[string]FlavorServiceDef `yaml:"services"`
}

var allConversions = []string{"request_to_opac", "request_from_opac", "response_to_opac", "response_from_opac",
	"stream_response_to_opac", "stream_response_from_opac"}

func EnsureConversionNameValid(conversion string) {
	for _, p := range allConversions {
		if p == conversion {
			return
		}
	}
	panic("[Flavor] Invalid Conversion Name: " + conversion)
}

// Not all elements are defined in the YAML file. So need to handle and return nil
// Example: getConversionDef("chat", "request_to_opac")
func (f *FlavorDef) getConversionDef(service, conversion string) *FlavorConversionDef {
	EnsureConversionNameValid(conversion)
	if serviceDef, exists := f.Services[service]; exists {
		var def FlavorConversionDef
		switch conversion {
		case "request_to_opac":
			def = serviceDef.RequestToOPAC
		case "request_from_opac":
			def = serviceDef.RequestFromOPAC
		case "response_to_opac":
			def = serviceDef.ResponseToOPAC
		case "response_from_opac":
			def = serviceDef.ResponseFromOPAC
		case "stream_response_to_opac":
			def = serviceDef.StreamResponseToOPAC
		case "stream_response_from_opac":
			def = serviceDef.StreamResponseFromOPAC
		default:
			panic("[Flavor] Invalid Conversion Name: " + conversion)
		}
		return &def
	}
	return nil
}

func LoadFlavorDef(flavor string) (FlavorDef, error) {
	env := GetEnv()
	fp := env.GetAbsolutePath("flavor_config/"+flavor+".yaml", env.RootDir)
	data, err := os.ReadFile(fp)
	if err != nil {
		return FlavorDef{}, err
	}
	var def FlavorDef
	err = yaml.Unmarshal(data, &def)
	if err != nil {
		return FlavorDef{}, err
	}
	if def.Name != flavor {
		return FlavorDef{}, fmt.Errorf("flavor name %s does not match file name %s", def.Name, flavor)
	}
	return def, err
}

var allFlavorDefs map[string]FlavorDef = make(map[string]FlavorDef)

func GetFlavorDef(flavor string) FlavorDef {
	// Force reload so changges in flavor config files take effect on the fly
	if _, exists := allFlavorDefs[flavor]; !exists || GetEnv().ForceReload {
		def, err := LoadFlavorDef(flavor)
		if err != nil {
			slog.Error("[Init] Failed to load flavor config", "flavor", flavor, "error", err)
			// This shouldn't happen unless something goes worng
			// Directly panic without recovering
			panic(err)
		}
		allFlavorDefs[flavor] = def
	}
	return allFlavorDefs[flavor]
}

//------------------------------------------------------------

func InitAPIFlavors() error {
	err := InitConverters()
	if err != nil {
		return err
	}
	env := GetEnv()
	fp := env.GetAbsolutePath("flavor_config", env.RootDir)
	files, err := os.ReadDir(fp)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".yaml" {
			baseName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			flavor, err := NewConfigBasedAPIFlavor(GetFlavorDef(baseName))
			if err != nil {
				slog.Error("[Flavor] Failed to create API Flavor", "flavor", baseName, "error", err)
				return err
			}
			RegisterAPIFlavor(flavor)
		}
	}
	return nil
}

// ------------------------------------------------------------
type ConfigBasedAPIFlavor struct {
	Config             FlavorDef
	converterPipelines map[string]map[string]*ConverterPipeline
}

func NewConfigBasedAPIFlavor(config FlavorDef) (*ConfigBasedAPIFlavor, error) {
	flavor := ConfigBasedAPIFlavor{
		Config: config,
	}
	err := flavor.reloadConfig()
	if err != nil {
		return nil, err
	}
	return &flavor, nil
}

// We need to do reload here instead of replace the entire pointer of CongigBasedAPIFlavor
// This is because we don't want to break the existing routes which are already installed
// with the Hanlder using the old pointer to ConfigBasedAPIFlavor
// So we can only update most of the internal states of ConfigBasedAPIFlavor
// NOTE: as stated, the routes etc. defined in the ConfigBasedAPIFlavor are not updated
func (f *ConfigBasedAPIFlavor) reloadConfig() error {
	// Reload the config if needed
	f.Config = GetFlavorDef(f.Config.Name)
	// rebuild the pipelines
	pipelines := make(map[string]map[string]*ConverterPipeline)
	for service := range f.Config.Services {
		pipelines[service] = make(map[string]*ConverterPipeline)
		for _, conv := range allConversions {
			// nil PipelineDef means empty []ConversionStepDef, it still creates a pipeline but
			// its steps are empty slice too
			p, err := NewConverterPipeline(f.Config.getConversionDef(service, conv).Conversion)
			if err != nil {
				return err
			}
			pipelines[service][conv] = p
		}
	}
	f.converterPipelines = pipelines
	// PPprint(">>> Rebuilt Converter Pipelines", f.converterPipelines)
	return nil
}

func (f *ConfigBasedAPIFlavor) GetConverterPipeline(service, conv string) *ConverterPipeline {
	EnsureConversionNameValid(conv)
	return f.converterPipelines[service][conv]
}

func (f *ConfigBasedAPIFlavor) Name() string {
	return f.Config.Name
}

func (f *ConfigBasedAPIFlavor) InstallRoutes(gateway *gin.Engine) {
	vSpec := GetEnv().SpecVersion
	for service, serviceDef := range f.Config.Services {
		for _, endpoint := range serviceDef.Endpoints {
			parts := strings.SplitN(endpoint, " ", 2)
			endpoint = strings.TrimSpace(endpoint)
			if len(parts) != 2 {
				slog.Error("[Flavor] Invalid endpoint format", "endpoint", endpoint)
				panic("[Flavor] Invalid endpoint format: " + endpoint)
			}
			method := parts[0]
			path := parts[1]
			method = strings.TrimSpace(method)
			path = strings.TrimSpace(path)
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			handler := makeServiceRequestHandler(f, service)

			// raw routes which doesn't have any opac prefix
			if serviceDef.InstallRawRoutes {
				gateway.Handle(method, path, handler)
				slog.Debug("[Flavor] Installed raw route", "flavor", f.Name(), "service", service, "route", method+" "+path)
			}
			// flavor routes in api_flavors or directly under services
			if f.Name() != "opac" {
				opacPath := "/opac/" + vSpec + "/api_flavors/" + f.Name() + path
				gateway.Handle(method, opacPath, handler)
				slog.Debug("[Flavor] Installed flavor route", "flavor", f.Name(), "service", service, "route", method+" "+opacPath)
				opacPath = "/opac/v0.1/opac/api_flavors/" + f.Name() + path
				gateway.Handle(method, opacPath, handler)
				slog.Debug("[Flavor] Installed v0.1 flavor route", "flavor", f.Name(), "service", service, "route", method+" "+opacPath)
			} else {
				opacPath := "/opac/" + vSpec + "/services" + path
				gateway.Handle(method, opacPath, makeServiceRequestHandler(f, service))
				slog.Debug("[Flavor] Installed opac route", "flavor", f.Name(), "service", service, "route", method+" "+opacPath)
				opacPath = "/opac/v0.1/opac/services" + path
				gateway.Handle(method, opacPath, makeServiceRequestHandler(f, service))
				slog.Debug("[Flavor] Installed v0.1 opac route", "flavor", f.Name(), "service", service, "route", method+" "+opacPath)
			}
		}
		slog.Info("[Flavor] Installed routes", "flavor", f.Name(), "service", service)
	}

}

func (f *ConfigBasedAPIFlavor) GetStreamResponseProlog(service string) []string {
	return f.Config.getConversionDef(service, "stream_response_from_opac").Prologue
}

func (f *ConfigBasedAPIFlavor) GetStreamResponseEpilog(service string) []string {
	return f.Config.getConversionDef(service, "stream_response_from_opac").Epilogue
}

func (f *ConfigBasedAPIFlavor) Convert(service, conversion string, content HTTPContent, ctx ConvertContext) (HTTPContent, error) {
	if GetEnv().ForceReload {
		err := f.reloadConfig()
		if err != nil {
			return HTTPContent{}, err
		}
	}
	pipeline := f.GetConverterPipeline(service, conversion)
	slog.Debug("[Flavor] Converting", "flavor", f.Name(), "service", service, "conversion", conversion, "content", content)
	return pipeline.Convert(content, ctx)
}

func (f *ConfigBasedAPIFlavor) ConvertRequestToOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error) {
	return f.Convert(service, "request_to_opac", content, ctx)
}
func (f *ConfigBasedAPIFlavor) ConvertRequestFromOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error) {
	return f.Convert(service, "request_from_opac", content, ctx)
}
func (f *ConfigBasedAPIFlavor) ConvertResponseToOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error) {
	return f.Convert(service, "response_to_opac", content, ctx)
}
func (f *ConfigBasedAPIFlavor) ConvertResponseFromOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error) {
	return f.Convert(service, "response_from_opac", content, ctx)
}
func (f *ConfigBasedAPIFlavor) ConvertStreamResponseToOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error) {
	return f.Convert(service, "stream_response_to_opac", content, ctx)
}
func (f *ConfigBasedAPIFlavor) ConvertStreamResponseFromOPAC(service string, content HTTPContent, ctx ConvertContext) (HTTPContent, error) {
	return f.Convert(service, "stream_response_from_opac", content, ctx)
}

func makeServiceRequestHandler(flavor APIFlavor, service string) func(c *gin.Context) {
	return func(c *gin.Context) {
		slog.Info("[Handler] Invoking service", "flavor", flavor.Name(), "service", service)
		SysEvents.Notify("start_session", []string{flavor.Name(), service})

		w := c.Writer

		taskid, ch, err := InvokeService(flavor.Name(), service, c.Request)
		if err != nil {
			slog.Error("[Handler] Failed to invoke service", "flavor", flavor.Name(), "service", service, "error", err)
			http.NotFound(w, c.Request)
			return
		}

		closenotifier, ok := w.(http.CloseNotifier)
		if !ok {
			slog.Error("[Handler] Not found http.CloseNotifier")
			http.NotFound(w, c.Request)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.NotFound(w, c.Request)
			return
		}

		isHTTPCompleted := false
	outerLoop:
		for {
			select {
			case <-closenotifier.CloseNotify():
				slog.Warn("[Handler] Client connection disconnected", "taskid", taskid)
				isHTTPCompleted = true
			case data, ok := <-ch:
				if !ok {
					slog.Debug("[Handler] Service task channel closed", "taskid", taskid)
					break outerLoop
				}
				slog.Debug("[Handler] Received service result", "result", data)
				if isHTTPCompleted {
					// skip below statements but do not quit
					// we should exhaust the channel to allow it to be closed
					continue
				}
				if data.Type == ServiceResultDone || data.Type == ServiceResultFailed {
					isHTTPCompleted = true
				}
				data.WriteBack(w)
				flusher.Flush()
			}
		}
		SysEvents.Notify("end_session", []string{flavor.Name(), service})
	}
}

func ConvertBetweenFlavors(from, to APIFlavor, service string, conv string, content HTTPContent, ctx ConvertContext) (HTTPContent, error) {
	if from.Name() == to.Name() {
		return content, nil
	}

	// need conversion, content-length may change
	content.Header.Del("Content-Length")

	firstConv := conv + "_to_opac"
	secondConv := conv + "_from_opac"
	EnsureConversionNameValid(firstConv)
	EnsureConversionNameValid(secondConv)
	if from.Name() != "opac" {
		var err error
		content, err = from.Convert(service, firstConv, content, ctx)
		if err != nil {
			return HTTPContent{}, err
		}
	}
	if from.Name() != "opac" && to.Name() != "opac" {
		if strings.HasPrefix(conv, "request") {
			SysEvents.NotifyHTTPRequest("request_converted_to_opac", "<n/a>", "<n/a>", content.Header, content.Body)
		} else {
			SysEvents.NotifyHTTPResponse("response_converted_to_opac", -1, content.Header, content.Body)
		}
	}
	if to.Name() != "opac" {
		var err error
		content, err = to.Convert(service, secondConv, content, ctx)
		if err != nil {
			return HTTPContent{}, err
		}
	}
	return content, nil
}
