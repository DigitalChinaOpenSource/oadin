//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"oadin/config"
	"oadin/internal/constants"
	"oadin/internal/datastore"
	"oadin/internal/provider"
	"oadin/internal/types"
	serverUtils "oadin/internal/utils/server"
	"oadin/version"

	"github.com/gin-gonic/gin"
)

func InjectRouter(e *OADINCoreServer) {
	e.Router.Handle(http.MethodGet, "/", rootHandler)
	e.Router.Handle(http.MethodGet, "/health", healthHeader)
	e.Router.Handle(http.MethodGet, "/engine/health", engineHealthHandler)
	e.Router.Handle(http.MethodGet, "/version", getVersion)
	e.Router.Handle(http.MethodGet, "/engine/version", getEngineVersion)
	e.Router.Handle(http.MethodGet, "/update/status", updateAvailableHandler)
	e.Router.Handle(http.MethodPost, "/update", updateHandler)

	r := e.Router.Group("/" + constants.AppName + "/" + version.OADINSpecVersion)

	// service import / export
	r.Handle(http.MethodPost, "/service/export", e.ExportService)
	r.Handle(http.MethodPost, "/service/import", e.ImportService)

	// Inject the router into the server
	r.Handle(http.MethodPost, "/service/install", e.CreateAIGCService)
	r.Handle(http.MethodPut, "/service", e.UpdateAIGCService)
	r.Handle(http.MethodGet, "/service", e.GetAIGCServices)

	r.Handle(http.MethodGet, "/service_provider", e.GetServiceProviders)
	r.Handle(http.MethodGet, "/service_provider/detail", e.GetServiceProvider)
	r.Handle(http.MethodPost, "/service_provider", e.CreateServiceProvider)
	r.Handle(http.MethodPut, "/service_provider", e.UpdateServiceProvider)
	r.Handle(http.MethodDelete, "/service_provider", e.DeleteServiceProvider)

	r.Handle(http.MethodGet, "/model", e.GetModels)
	r.Handle(http.MethodPost, "/model", e.CreateModel)
	r.Handle(http.MethodDelete, "/model", e.DeleteModel)
	r.Handle(http.MethodPost, "/model/stream", e.CreateModelStream)
	r.Handle(http.MethodPost, "/model/stream/cancel", e.CancelModelStream)
	r.Handle(http.MethodGet, "/model/recommend", e.GetRecommendModels)
	r.Handle(http.MethodGet, "/model/support", e.GetModelList)

	r.Handle(http.MethodGet, "/control_panel/dashboard", e.GetDashBoardHandler)
	r.Handle(http.MethodGet, "/control_panel/modellist", e.GetSupportModelListCombine)
	r.Handle(http.MethodPost, "/control_panel/set_default", e.SetDefaultModelHandler)
	r.Handle(http.MethodGet, "/control_panel/about", e.GetProductInfoHandler)
	r.Handle(http.MethodPost, "/control_panel/modelkey", e.GetModelkeyHandler)

	slog.Info("Gateway started", "host", config.GlobalEnvironment.ApiHost)
}

func rootHandler(c *gin.Context) {
	c.String(http.StatusOK, "AIPC Open Gateway")
}

func healthHeader(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"status": "UP"})
}

func engineHealthHandler(c *gin.Context) {
	data := make(map[string]string)
	for _, modelEngineName := range types.SupportModelEngine {
		engine := provider.GetModelEngine(modelEngineName)
		err := engine.HealthCheck()
		if err != nil {
			data[modelEngineName] = "DOWN"
			continue
		}
		data[modelEngineName] = "UP"
	}
	c.JSON(http.StatusOK, data)
}

func getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"version": version.OADINVersion})
}

func getEngineVersion(c *gin.Context) {
	ctx := c.Request.Context()
	data := make(map[string]string)
	for _, modelEngineName := range types.SupportModelEngine {
		engine := provider.GetModelEngine(modelEngineName)
		var respData types.EngineVersionResponse
		resp, err := engine.GetVersion(ctx, &respData)
		if err != nil {
			data[modelEngineName] = "get version failed"
			continue
		}
		data[modelEngineName] = resp.Version
	}

	c.JSON(http.StatusOK, data)
}

func updateAvailableHandler(c *gin.Context) {
	ctx := c.Request.Context()
	status, updateResp := version.IsNewVersionAvailable(ctx)
	if status {
		c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("Ollama version %s is ready to install", updateResp.UpdateVersion)})
	} else {
		c.JSON(http.StatusOK, map[string]string{"message": ""})
	}
}

func updateHandler(c *gin.Context) {
	// check server
	status := serverUtils.IsServerRunning()
	if status {
		// stop server
		pidFilePath := filepath.Join(config.GlobalEnvironment.RootDir, "oadin.pid")
		err := serverUtils.StopOadinServer(pidFilePath)
		if err != nil {
			c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
		}
	}
	// rm old version file
	oadinFileName := "oadin.exe"
	if runtime.GOOS != "windows" {
		oadinFileName = "oadin"
	}
	oadinFilePath := filepath.Join(config.GlobalEnvironment.RootDir, oadinFileName)
	err := os.Remove(oadinFilePath)
	if err != nil {
		slog.Error("[Update] Failed to remove oadin file %s: %v\n", oadinFilePath, err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	// install new version
	downloadPath := filepath.Join(config.GlobalEnvironment.RootDir, "download", oadinFileName)
	err = os.Rename(downloadPath, oadinFilePath)
	if err != nil {
		slog.Error("[Update] Failed to rename oadin file %s: %v\n", downloadPath, err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	// start server
	logPath := config.GlobalEnvironment.ConsoleLog
	rootDir := config.GlobalEnvironment.RootDir
	err = serverUtils.StartOadinServer(logPath, rootDir)
	if err != nil {
		slog.Error("[Update] Failed to start oadin log %s: %v\n", logPath, err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	ds := datastore.GetDefaultDatastore()
	ctx := c.Request.Context()
	vr := &types.VersionUpdateRecord{}
	sortOption := []datastore.SortOption{
		{Key: "created_at", Order: -1},
	}
	versionRecoreds, err := ds.List(ctx, vr, &datastore.ListOptions{SortBy: sortOption})
	if err != nil {
		slog.Error("[Update] Failed to list versions: %v\n", err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	versionRecord := versionRecoreds[0].(*types.VersionUpdateRecord)
	if versionRecord.Status == types.VersionRecordStatusInstalled {
		versionRecord.Status = types.VersionRecordStatusUpdated
	}
	err = ds.Put(ctx, versionRecord)
	if err != nil {
		slog.Error("[Update] Failed to update versions: %v\n", err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	c.JSON(http.StatusOK, map[string]string{"message": ""})
}
