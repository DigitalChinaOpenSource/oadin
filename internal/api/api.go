package api

import (
	"context"
	"fmt"

	"byze/internal/server"

	"github.com/gin-gonic/gin"
)

type ByzeCoreServer struct {
	Router          *gin.Engine
	AIGCService     server.AIGCService
	Model           server.Model
	ServiceProvider server.ServiceProvider
	MCP             server.MCPServer
}

// NewByzeCoreServer is the constructor of the server structure
func NewByzeCoreServer() *ByzeCoreServer {
	g := gin.Default()
	err := g.SetTrustedProxies(nil)
	if err != nil {
		fmt.Println("SetTrustedProxies failed")
		return nil
	}
	return &ByzeCoreServer{
		Router: g,
	}
}

// Run is the function to start the server
func (t *ByzeCoreServer) Run(ctx context.Context, address string) error {
	return t.Router.Run(address)
}

func (t *ByzeCoreServer) Register() {
	t.AIGCService = server.NewAIGCService()
	t.ServiceProvider = server.NewServiceProvider()
	t.Model = server.NewModel()
	t.MCP = server.NewMCPServer()
}
