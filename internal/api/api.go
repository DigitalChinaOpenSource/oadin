package api

import (
	"context"
	"fmt"

	"oadin/internal/datastore"
	"oadin/internal/server"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type OadinCoreServer struct {
	Router          *gin.Engine
	AIGCService     server.AIGCService
	Model           server.Model
	ServiceProvider server.ServiceProvider
	MCP             server.MCPServer
	System          server.System
	Playground      server.Playground
	DataStore       datastore.Datastore
}

// NewOadinCoreServer is the constructor of the server structure
func NewOadinCoreServer() *OadinCoreServer {
	g := gin.Default()
	g.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))
	err := g.SetTrustedProxies(nil)
	if err != nil {
		fmt.Println("SetTrustedProxies failed")
		return nil
	}
	return &OadinCoreServer{
		Router: g,
	}
}

// Run is the function to start the server
func (t *OadinCoreServer) Run(ctx context.Context, address string) error {
	return t.Router.Run(address)
}

func (t *OadinCoreServer) Register() {
	t.AIGCService = server.NewAIGCService()
	t.ServiceProvider = server.NewServiceProvider()
	t.Model = server.NewModel()
	t.MCP = server.NewMCPServer()
	t.System = server.NewSystemImpl()
	t.Playground = server.NewPlayground()
	t.DataStore = datastore.GetDefaultDatastore()
}
