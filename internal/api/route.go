// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"intel.com/aog/config"
	"intel.com/aog/version"
)

func InjectRouter(e *AOGCoreServer) {
	e.Router.Handle(http.MethodGet, "/", rootHandler)

	r := e.Router.Group("/aog/" + version.AOGVersion)

	r.Handle(http.MethodGet, "/health", healthHeader)

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

	slog.Info("Gateway started", "host", config.GlobalAOGEnvironment.ApiHost)
}

func rootHandler(c *gin.Context) {
	c.String(http.StatusOK, "Open Platform for AIPC")
}

func healthHeader(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"status": "UP"})
}
