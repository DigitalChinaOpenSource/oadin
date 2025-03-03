// Apache v2 license
// Copyright (C) 2024 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package opac

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func StartGateway(port int, isDebug bool) {
	vSpec := GetEnv().SpecVersion
	if !isDebug {
		gin.SetMode(gin.ReleaseMode)
	}
	gateway := gin.Default()
	// Local gateway shouldn't be passed through proxy so no trusted proxies
	gateway.SetTrustedProxies(nil)
	gateway.GET("/", rootHandler)
	gateway.GET("/opac/"+vSpec+"/health", healthHeader)
	slog.Info("Gateway started", "port", port)

	// Setup flavors
	for _, flavor := range AllAPIFlavors() {
		flavor.InstallRoutes(gateway)
	}

	// NOTE: only listens at local and reject remote access
	gateway.Run("localhost:" + strconv.Itoa(port))
}

func rootHandler(c *gin.Context) {
	c.String(http.StatusOK, "Open Platform for AIPC")
}

func healthHeader(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"status": "UP"})
}
