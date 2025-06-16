// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"

	"byze/config"
	"byze/internal/datastore"
	"byze/internal/provider"
	"byze/internal/types"
	"byze/internal/utils"
	"byze/tray"
	"byze/version"
)

func InjectRouter(e *ByzeCoreServer) {
	e.Router.Handle(http.MethodGet, "/", rootHandler)
	e.Router.Handle(http.MethodGet, "/health", healthHeader)
	e.Router.Handle(http.MethodGet, "/engine/health", engineHealthHandler)
	e.Router.Handle(http.MethodGet, "/version", getVersion)
	e.Router.Handle(http.MethodGet, "/engine/version", getEngineVersion)
	e.Router.Handle(http.MethodGet, "/update/status", updateAvailableHandler)
	e.Router.Handle(http.MethodPost, "/update", updateHandler)

	r := e.Router.Group("/byze/" + version.ByzeVersion)

	// service import / export
	r.Handle(http.MethodPost, "/service/export", e.ExportService)
	r.Handle(http.MethodPost, "/service/import", e.ImportService)

	// Inject the router into the server
	r.Handle(http.MethodPost, "/service", e.CreateAIGCService)
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

	r.Handle(http.MethodGet, "/model/support/smartvision", e.GetSmartVisionSupportModelList)

	r.Handle(http.MethodGet, "/control_panel/model/filepath", e.GetModelFilePathHandler)
	r.Handle(http.MethodPost, "/control_panel/model/filepath", e.ModifyModelFilePathHandler)
	r.Handle(http.MethodGet, "/control_panel/path/space", e.GetPathDiskSizeHandler)
	r.Handle(http.MethodGet, "/control_panel/model/square", e.GetSupportModelListCombine)

	// Apis related to MCP
	mcpApi := r.Group("/mcp")

	mcpApi.POST("/search", e.GetMCPList)
	mcpApi.GET("/:id", e.GetMCPDetail)
	mcpApi.POST("/:id/tools/search", e.GetKits)
	mcpApi.GET("/:id/clients", e.GetClients)
	mcpApi.GET("/categories", e.GetCategories)
	mcpApi.PUT("/:id/download", e.DownloadMCP)
	mcpApi.POST("/mine", e.GetMyMCPList)
	mcpApi.PUT("/:id/auth", e.AuthorizeMCP)
	mcpApi.PUT("/:id/reverse", e.ReverseStatus)
	mcpApi.PUT("/setup", e.SetupFunTool)
	mcpApi.POST("/client/start", e.ClientMcpStart)
	mcpApi.POST("/client/stop", e.ClientMcpStop)
	mcpApi.POST("/client/getTools", e.ClientGetTools)
	mcpApi.POST("/client/runTool", e.ClientRunTool)

	// Apis related to system
	systemApi := r.Group("system")

	systemApi.GET("/about", e.About)
	systemApi.GET("/information", e.SystemSettings)
	systemApi.PUT("/registry", e.ModifyRepositoryURL)
	systemApi.PUT("/proxy", e.SetProxy)
	systemApi.PUT("/proxy/switch", e.ProxySwitch)
	// Playground相关
	playgroundApi := r.Group("playground")
	playgroundApi.POST("/session", e.CreateSession)
	playgroundApi.GET("/sessions", e.GetSessions)
	playgroundApi.DELETE("/session", e.DeleteSession)
	playgroundApi.POST("/message", e.SendMessage)
	playgroundApi.POST("/message/stream", e.SendMessageStream)
	playgroundApi.GET("/messages", e.GetMessages)
	playgroundApi.POST("/file", e.UploadFile)
	playgroundApi.GET("/files", e.GetFiles)
	playgroundApi.DELETE("/file", e.DeleteFile)
	playgroundApi.POST("/file/process", e.ProcessFile)
	playgroundApi.POST("/session/model", e.ChangeSessionModel)

	slog.Info("Gateway started", "host", config.GlobalByzeEnvironment.ApiHost)
}

func rootHandler(c *gin.Context) {
	c.String(http.StatusOK, "Open Platform for AIPC")
}

func healthHeader(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"status": "UP"})
}

func engineHealthHandler(c *gin.Context) {
	providerName := "ollama"
	engine := provider.GetModelEngine(providerName)
	err := engine.HealthCheck()
	status := 1
	if err != nil {
		status = 0
	}
	c.JSON(http.StatusOK, map[string]interface{}{"status": status})
}

func getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"version": version.ByzeVersion + "-" + version.ByzeSubVersion})
}

func getEngineVersion(c *gin.Context) {
	ctx := c.Request.Context()
	providerName := "ollama"
	engine := provider.GetModelEngine(providerName)
	var respData types.EngineVersionResponse
	resp, err := engine.GetVersion(ctx, &respData)
	if err != nil {
		c.JSON(http.StatusOK, map[string]string{"error": err.Error()})
	}

	c.JSON(http.StatusOK, map[string]interface{}{"version": resp.Version})
}

func updateAvailableHandler(c *gin.Context) {
	ctx := c.Request.Context()
	status, updateResp := tray.IsNewVersionAvailable(ctx)
	if status {
		c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("Ollama version %s is ready to install", updateResp.UpdateVersion)})
	} else {
		c.JSON(http.StatusOK, map[string]string{"message": ""})
	}
}

func updateHandler(c *gin.Context) {
	// check server
	status := utils.IsServerRunning()
	if status {
		// stop server
		pidFilePath := filepath.Join(config.GlobalByzeEnvironment.RootDir, "byze.pid")
		err := utils.StopByzeServer(pidFilePath)
		if err != nil {
			c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
		}
	}
	// rm old version file
	byzeFileName := "byze.exe"
	if runtime.GOOS != "windows" {
		byzeFileName = "byze"
	}
	byzeFilePath := filepath.Join(config.GlobalByzeEnvironment.RootDir, byzeFileName)
	err := os.Remove(byzeFilePath)
	if err != nil {
		slog.Error("[Update] Failed to remove byze file %s: %v\n", byzeFilePath, err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	// install new version
	downloadPath := filepath.Join(config.GlobalByzeEnvironment.RootDir, "download", byzeFileName)
	err = os.Rename(downloadPath, byzeFilePath)
	if err != nil {
		slog.Error("[Update] Failed to rename byze file %s: %v\n", downloadPath, err)
		c.JSON(http.StatusOK, map[string]string{"message": err.Error()})
	}
	// start server
	logPath := config.GlobalByzeEnvironment.ConsoleLog
	rootDir := config.GlobalByzeEnvironment.RootDir
	err = utils.StartByzeServer(logPath, rootDir)
	if err != nil {
		slog.Error("[Update] Failed to start byze log %s: %v\n", logPath, err)
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
