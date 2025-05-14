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
