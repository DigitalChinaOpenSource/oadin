// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package schedule

import (
	"aipc/byze/config"
	"aipc/byze/internal/convert"
	"aipc/byze/internal/event"
	"aipc/byze/internal/provider/template"
	"aipc/byze/internal/types"
	"aipc/byze/internal/utils"
	"aipc/byze/version"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// APIFlavor mode is usually set to "default". And set to "stream" if it is using stream mode
type APIFlavor interface {
	Name() string
	InstallRoutes(server *gin.Engine, options *config.ByzeEnvironment)

	// GetStreamResponseProlog In stream mdoe, some flavor may ask for some packets to be send first
	// or at the end, in addition to normal contents. For example, OpenAI
	// needs to send an additional "data: [DONE]" after everything is done.
	GetStreamResponseProlog(service string) []string
	GetStreamResponseEpilog(service string) []string

	// Convert This should cover the 6 conversion methods below
	Convert(service string, conversion string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error)

	ConvertRequestToByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error)
	ConvertRequestFromByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error)
	ConvertResponseToByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error)
	ConvertResponseFromByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error)
	ConvertStreamResponseToByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error)
	ConvertStreamResponseFromByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error)
}

var allFlavors = make(map[string]APIFlavor)

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

type FlavorConversionDef struct {
	Prologue   []string                  `yaml:"prologue"`
	Epilogue   []string                  `yaml:"epilogue"`
	Conversion []types.ConversionStepDef `yaml:"conversion"`
}

type ModelSelector struct {
	ModelInRequest  string `yaml:"request"`
	ModelInResponse string `yaml:"response"`
}
type FlavorServiceDef struct {
	Endpoints              []string            `yaml:"endpoints"`
	InstallRawRoutes       bool                `yaml:"install_raw_routes"`
	DefaultModel           string              `yaml:"default_model"`
	RequestUrl             string              `yaml:"url"`
	RequestExtraUrl        string              `yaml:"extra_url"`
	AuthType               string              `yaml:"auth_type"`
	AuthApplyUrl           string              `yaml:"auth_apply_url"`
	RequestSegments        int                 `yaml:"request_segments"`
	ExtraHeaders           string              `yaml:"extra_headers"`
	SupportModels          []string            `yaml:"support_models"`
	ModelSelector          ModelSelector       `yaml:"model_selector"`
	RequestToByze          FlavorConversionDef `yaml:"request_to_byze"`
	RequestFromByze        FlavorConversionDef `yaml:"request_from_byze"`
	ResponseToByze         FlavorConversionDef `yaml:"response_to_byze"`
	ResponseFromByze       FlavorConversionDef `yaml:"response_from_byze"`
	StreamResponseToByze   FlavorConversionDef `yaml:"stream_response_to_byze"`
	StreamResponseFromByze FlavorConversionDef `yaml:"stream_response_from_byze"`
}

type FlavorDef struct {
	Version  string                      `yaml:"version"`
	Name     string                      `yaml:"name"`
	Endpoint string                      `yaml:"endpoint"`
	AuthType string                      `yaml:"auth_type"`
	Services map[string]FlavorServiceDef `yaml:"services"`
}

var allConversions = []string{
	"request_to_byze", "request_from_byze", "response_to_byze", "response_from_byze",
	"stream_response_to_byze", "stream_response_from_byze",
}

func EnsureConversionNameValid(conversion string) {
	for _, p := range allConversions {
		if p == conversion {
			return
		}
	}
	panic("[Flavor] Invalid Conversion Name: " + conversion)
}

// Not all elements are defined in the YAML file. So need to handle and return nil
// Example: getConversionDef("chat", "request_to_byze")
func (f *FlavorDef) getConversionDef(service, conversion string) *FlavorConversionDef {
	EnsureConversionNameValid(conversion)
	if serviceDef, exists := f.Services[service]; exists {
		var def FlavorConversionDef
		switch conversion {
		case "request_to_byze":
			def = serviceDef.RequestToByze
		case "request_from_byze":
			def = serviceDef.RequestFromByze
		case "response_to_byze":
			def = serviceDef.ResponseToByze
		case "response_from_byze":
			def = serviceDef.ResponseFromByze
		case "stream_response_to_byze":
			def = serviceDef.StreamResponseToByze
		case "stream_response_from_byze":
			def = serviceDef.StreamResponseFromByze
		default:
			panic("[Flavor] Invalid Conversion Name: " + conversion)
		}
		return &def
	}
	return nil
}

func LoadFlavorDef(flavor, rootDir string) (FlavorDef, error) {
	data, err := template.FlavorTemplateFs.ReadFile(flavor + ".yaml")
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

var allFlavorDefs = make(map[string]FlavorDef)

func GetFlavorDef(flavor string) FlavorDef {
	// Force reload so changes in flavor config files take effect on the fly
	if _, exists := allFlavorDefs[flavor]; !exists {
		def, err := LoadFlavorDef(flavor, config.GlobalByzeEnvironment.RootDir)
		if err != nil {
			slog.Error("[Init] Failed to load flavor config", "flavor", flavor, "error", err)
			// This shouldn't happen unless something goes wrong
			// Directly panic without recovering
			panic(err)
		}
		allFlavorDefs[flavor] = def
	}
	return allFlavorDefs[flavor]
}

//------------------------------------------------------------

func InitAPIFlavors() error {
	err := convert.InitConverters()
	if err != nil {
		return err
	}
	files, err := template.FlavorTemplateFs.ReadDir(".")
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
	converterPipelines map[string]map[string]*convert.ConverterPipeline
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

// We need to do reload here instead of replace the entire pointer of ConfigBasedAPIFlavor
// This is because we don't want to break the existing routes which are already installed
// with the Handler using the old pointer to ConfigBasedAPIFlavor
// So we can only update most of the internal states of ConfigBasedAPIFlavor
// NOTE: as stated, the routes etc. defined in the ConfigBasedAPIFlavor are not updated
func (f *ConfigBasedAPIFlavor) reloadConfig() error {
	// Reload the config if needed
	f.Config = GetFlavorDef(f.Config.Name)
	// rebuild the pipelines
	pipelines := make(map[string]map[string]*convert.ConverterPipeline)
	for service := range f.Config.Services {
		pipelines[service] = make(map[string]*convert.ConverterPipeline)
		for _, conv := range allConversions {
			// nil PipelineDef means empty []ConversionStepDef, it still creates a pipeline but
			// its steps are empty slice too
			p, err := convert.NewConverterPipeline(f.Config.getConversionDef(service, conv).Conversion)
			if err != nil {
				return err
			}
			pipelines[service][conv] = p
		}
	}
	f.converterPipelines = pipelines
	// PPrint(">>> Rebuilt Converter Pipelines", f.converterPipelines)
	return nil
}

func (f *ConfigBasedAPIFlavor) GetConverterPipeline(service, conv string) *convert.ConverterPipeline {
	EnsureConversionNameValid(conv)
	return f.converterPipelines[service][conv]
}

func (f *ConfigBasedAPIFlavor) Name() string {
	return f.Config.Name
}

func (f *ConfigBasedAPIFlavor) InstallRoutes(gateway *gin.Engine, options *config.ByzeEnvironment) {
	vSpec := version.ByzeVersion
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

			// raw routes which doesn't have any byze prefix
			if serviceDef.InstallRawRoutes {
				gateway.Handle(method, path, handler)
				slog.Debug("[Flavor] Installed raw route", "flavor", f.Name(), "service", service, "route", method+" "+path)
			}
			// flavor routes in api_flavors or directly under services
			if f.Name() != "byze" {
				byzePath := "/byze/" + vSpec + "/api_flavors/" + f.Name() + path
				gateway.Handle(method, byzePath, handler)
				slog.Debug("[Flavor] Installed flavor route", "flavor", f.Name(), "service", service, "route", method+" "+byzePath)
			} else {
				byzePath := "/byze/" + vSpec + "/services" + path
				gateway.Handle(method, byzePath, makeServiceRequestHandler(f, service))
				slog.Debug("[Flavor] Installed byze route", "flavor", f.Name(), "service", service, "route", method+" "+byzePath)
			}
		}
		slog.Info("[Flavor] Installed routes", "flavor", f.Name(), "service", service)
	}
}

func (f *ConfigBasedAPIFlavor) GetStreamResponseProlog(service string) []string {
	return f.Config.getConversionDef(service, "stream_response_from_byze").Prologue
}

func (f *ConfigBasedAPIFlavor) GetStreamResponseEpilog(service string) []string {
	return f.Config.getConversionDef(service, "stream_response_from_byze").Epilogue
}

func (f *ConfigBasedAPIFlavor) Convert(service, conversion string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error) {
	pipeline := f.GetConverterPipeline(service, conversion)
	slog.Debug("[Flavor] Converting", "flavor", f.Name(), "service", service, "conversion", conversion, "content", content)
	return pipeline.Convert(content, ctx)
}

func (f *ConfigBasedAPIFlavor) ConvertRequestToByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error) {
	return f.Convert(service, "request_to_byze", content, ctx)
}

func (f *ConfigBasedAPIFlavor) ConvertRequestFromByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error) {
	return f.Convert(service, "request_from_byze", content, ctx)
}

func (f *ConfigBasedAPIFlavor) ConvertResponseToByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error) {
	return f.Convert(service, "response_to_byze", content, ctx)
}

func (f *ConfigBasedAPIFlavor) ConvertResponseFromByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error) {
	return f.Convert(service, "response_from_byze", content, ctx)
}

func (f *ConfigBasedAPIFlavor) ConvertStreamResponseToByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error) {
	return f.Convert(service, "stream_response_to_byze", content, ctx)
}

func (f *ConfigBasedAPIFlavor) ConvertStreamResponseFromByze(service string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error) {
	return f.Convert(service, "stream_response_from_byze", content, ctx)
}

func makeServiceRequestHandler(flavor APIFlavor, service string) func(c *gin.Context) {
	return func(c *gin.Context) {
		slog.Info("[Handler] Invoking service", "flavor", flavor.Name(), "service", service)
		event.SysEvents.Notify("start_session", []string{flavor.Name(), service})

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
				if data.Type == types.ServiceResultDone || data.Type == types.ServiceResultFailed {
					isHTTPCompleted = true
				}
				data.WriteBack(w)
				flusher.Flush()
			}
		}
		event.SysEvents.Notify("end_session", []string{flavor.Name(), service})
	}
}

func ConvertBetweenFlavors(from, to APIFlavor, service string, conv string, content types.HTTPContent, ctx convert.ConvertContext) (types.HTTPContent, error) {
	if from.Name() == to.Name() {
		return content, nil
	}

	// need conversion, content-length may change
	content.Header.Del("Content-Length")

	firstConv := conv + "_to_byze"
	secondConv := conv + "_from_byze"
	EnsureConversionNameValid(firstConv)
	EnsureConversionNameValid(secondConv)
	if from.Name() != "byze" {
		var err error
		content, err = from.Convert(service, firstConv, content, ctx)
		if err != nil {
			return types.HTTPContent{}, err
		}
	}
	if from.Name() != "byze" && to.Name() != "byze" {
		if strings.HasPrefix(conv, "request") {
			event.SysEvents.NotifyHTTPRequest("request_converted_to_byze", "<n/a>", "<n/a>", content.Header, content.Body)
		} else {
			event.SysEvents.NotifyHTTPResponse("response_converted_to_byze", -1, content.Header, content.Body)
		}
	}
	if to.Name() != "byze" {
		var err error
		content, err = to.Convert(service, secondConv, content, ctx)
		if err != nil {
			return types.HTTPContent{}, err
		}
	}
	return content, nil
}

type ServiceDefaultInfo struct {
	Endpoints       []string `json:"endpoints"`
	DefaultModel    string   `json:"default_model"`
	RequestUrl      string   `json:"url"`
	RequestExtraUrl string   `json:"request_extra_url"`
	AuthType        string   `json:"auth_type"`
	RequestSegments int      `json:"request_segments"`
	ExtraHeaders    string   `json:"extra_headers"`
	SupportModels   []string `json:"support_models"`
	AuthApplyUrl    string   `json:"auth_apply_url"`
}

var FlavorServiceDefaultInfoMap = make(map[string]map[string]ServiceDefaultInfo)

func InitProviderDefaultModelTemplate(flavor APIFlavor) {
	def, err := LoadFlavorDef(flavor.Name(), "/")
	if err != nil {
		slog.Error("[Provider]Failed to load file", "provider_name", flavor, "error", err.Error())
	}
	ServiceDefaultInfoMap := make(map[string]ServiceDefaultInfo)
	for service, serviceDef := range def.Services {
		ServiceDefaultInfoMap[service] = ServiceDefaultInfo{
			Endpoints:       serviceDef.Endpoints,
			DefaultModel:    serviceDef.DefaultModel,
			RequestUrl:      serviceDef.RequestUrl,
			RequestExtraUrl: serviceDef.RequestExtraUrl,
			RequestSegments: serviceDef.RequestSegments,
			AuthType:        serviceDef.AuthType,
			ExtraHeaders:    serviceDef.ExtraHeaders,
			SupportModels:   serviceDef.SupportModels,
			AuthApplyUrl:    serviceDef.AuthApplyUrl,
		}
	}
	FlavorServiceDefaultInfoMap[flavor.Name()] = ServiceDefaultInfoMap
}

func GetProviderServiceDefaultInfo(flavor string, service string) ServiceDefaultInfo {
	serviceDefaultInfo := FlavorServiceDefaultInfoMap[flavor][service]
	return serviceDefaultInfo
}

type SignParams struct {
	SecretId      string           `json:"secret_id"`
	SecretKey     string           `json:"secret_key"`
	RequestBody   string           `json:"request_body"`
	RequestUrl    string           `json:"request_url"`
	RequestMethod string           `json:"request_method"`
	RequestHeader http.Header      `json:"request_header"`
	CommonParams  SignCommonParams `json:"common_params"`
}

type SignCommonParams struct {
	Version string `json:"version"`
	Action  string `json:"action"`
	Region  string `json:"region"`
}

func TencentSignGenerate(p SignParams, req http.Request) error {
	secretId := p.SecretId
	secretKey := p.SecretKey
	parseUrl, err := url.Parse(p.RequestUrl)
	if err != nil {
		return err
	}
	host := parseUrl.Host
	service := strings.Split(host, ".")[0]
	algorithm := "TC3-HMAC-SHA256"
	version := p.CommonParams.Version
	action := p.CommonParams.Action
	region := p.CommonParams.Region
	var timestamp int64 = time.Now().Unix()

	// step 1: build canonical request string
	httpRequestMethod := p.RequestMethod
	canonicalURI := "/"
	canonicalQueryString := ""
	canonicalHeaders := ""
	signedHeaders := ""
	for k, v := range p.RequestHeader {
		if strings.ToLower(k) == "content-type" {
			signedHeaders += fmt.Sprintf("%s;", strings.ToLower(k))
			canonicalHeaders += fmt.Sprintf("%s:%s\n", strings.ToLower(k), strings.ToLower(v[0]))
		}
	}
	signedHeaders += "host"
	canonicalHeaders += fmt.Sprintf("%s:%s\n", "host", host)
	signedHeaders = strings.TrimRight(signedHeaders, ";")
	payload := p.RequestBody
	hashedRequestPayload := utils.Sha256hex(payload)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)

	// step 2: build string to sign
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	hashedCanonicalRequest := utils.Sha256hex(canonicalRequest)
	string2sign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithm,
		timestamp,
		credentialScope,
		hashedCanonicalRequest)

	// step 3: sign string
	secretDate := utils.HmacSha256(date, "TC3"+secretKey)
	secretService := utils.HmacSha256(service, secretDate)
	secretSigning := utils.HmacSha256("tc3_request", secretService)
	signature := hex.EncodeToString([]byte(utils.HmacSha256(string2sign, secretSigning)))

	// step 4: build authorization
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		secretId,
		credentialScope,
		signedHeaders,
		signature)

	req.Header.Add("Authorization", authorization)
	req.Header.Add("X-TC-Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Add("X-TC-Version", version)
	req.Header.Add("X-TC-Region", region)
	req.Header.Add("X-TC-Action", action)
	return nil
}

type SignAuthInfo struct {
	SecretId  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
}
type Authenticator interface {
	Authenticate() error
}

type APIKEYAuthenticator struct {
	APIKey string
	Req    http.Request
}

type TencentSignAuthenticator struct {
	AuthInfo     string                `json:"auth_info"`
	Req          http.Request          `json:"request"`
	ProviderInfo types.ServiceProvider `json:"provider_info"`
	ReqBody      string                `json:"req_body"`
}

func (a *APIKEYAuthenticator) Authenticate() error {
	a.Req.Header.Set("Authorization", "Bearer "+a.APIKey)
	return nil
}

func (s *TencentSignAuthenticator) Authenticate() error {
	var authInfoData SignAuthInfo
	err := json.Unmarshal([]byte(s.AuthInfo), &authInfoData)
	if err != nil {
		return err
	}

	commonParams := SignParams{
		SecretId:      authInfoData.SecretId,
		SecretKey:     authInfoData.SecretKey,
		RequestUrl:    s.ProviderInfo.URL,
		RequestBody:   s.ReqBody,
		RequestHeader: s.Req.Header,
		RequestMethod: s.Req.Method,
	}
	if s.ProviderInfo.ExtraHeaders != "" {
		var serviceExtraInfo SignCommonParams
		err := json.Unmarshal([]byte(s.ProviderInfo.ExtraHeaders), &serviceExtraInfo)
		if err != nil {
			return err
		}
		commonParams.CommonParams = serviceExtraInfo
	}

	err = TencentSignGenerate(commonParams, s.Req)
	if err != nil {
		return err
	}
	return nil
}

type AuthenticatorParams struct {
	Request      *http.Request
	ProviderInfo *types.ServiceProvider
	RequestBody  string
}

func ChooseProviderAuthenticator(p *AuthenticatorParams) Authenticator {
	var authenticator Authenticator
	if p.ProviderInfo.AuthType == types.AuthTypeToken {
		switch p.ProviderInfo.Flavor {
		case types.FlavorTencent:
			authenticator = &TencentSignAuthenticator{
				Req:          *p.Request,
				AuthInfo:     p.ProviderInfo.AuthKey,
				ProviderInfo: *p.ProviderInfo,
				ReqBody:      p.RequestBody,
			}
		}
	} else if p.ProviderInfo.AuthType == types.AuthTypeApiKey {
		authenticator = &APIKEYAuthenticator{
			APIKey: p.ProviderInfo.AuthKey,
			Req:    *p.Request,
		}
	}
	return authenticator
}
