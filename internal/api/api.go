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
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"intel.com/aog/internal/server"
)

type AOGCoreServer struct {
	Router          *gin.Engine
	AIGCService     server.AIGCService
	Model           server.Model
	ServiceProvider server.ServiceProvider
}

// NewAOGCoreServer is the constructor of the server structure
func NewAOGCoreServer() *AOGCoreServer {
	g := gin.Default()
	err := g.SetTrustedProxies(nil)
	if err != nil {
		fmt.Println("SetTrustedProxies failed")
		return nil
	}
	return &AOGCoreServer{
		Router: g,
	}
}

// Run is the function to start the server
func (t *AOGCoreServer) Run(ctx context.Context, address string) error {
	return t.Router.Run(address)
}

func (t *AOGCoreServer) Register() {
	t.AIGCService = server.NewAIGCService()
	t.ServiceProvider = server.NewServiceProvider()
	t.Model = server.NewModel()
}
