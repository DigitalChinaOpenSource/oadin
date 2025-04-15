// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"aipc/byze/config"
	"aipc/byze/version"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InjectRouter(e *ByzeCoreServer) {
	e.Router.Handle(http.MethodGet, "/", rootHandler)

	r := e.Router.Group("/byze/" + version.ByzeVersion)

	r.Handle(http.MethodGet, "/health", healthHeader)
	r.Handle(http.MethodGet, "/version", getVersion)

	// service import / export
	r.Handle(http.MethodPost, "/service/export", e.ExportService)
	r.Handle(http.MethodPost, "/service/import", e.ImportService)

	// Inject the router into the server
	r.Handle(http.MethodPost, "/service", e.CreateAIGCService)
	r.Handle(http.MethodPut, "/service", e.UpdateAIGCService)
	r.Handle(http.MethodGet, "/service", e.GetAIGCServices)

	r.Handle(http.MethodGet, "/service_provider", e.GetServiceProviders)
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

	slog.Info("Gateway started", "host", config.GlobalByzeEnvironment.ApiHost)
}

func rootHandler(c *gin.Context) {
	c.String(http.StatusOK, "Open Platform for AIPC")
}

func healthHeader(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"status": "UP"})
}

func getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"version": version.ByzeVersion})
}
