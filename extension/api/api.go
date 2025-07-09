package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"oadin/extension/server"
	"oadin/internal/api"
	"oadin/internal/datastore"
	"oadin/version"
)

// OadinExtensionServer 使用代理模式来扩展 OadinCoreServer 的功能
type OadinExtensionServer struct {
	// 持有父类实例来实现代理扩展
	CoreServer *api.OadinCoreServer

	// 继承父类属性和方法
	api.OadinCoreServer
	Router    *gin.Engine
	DataStore datastore.Datastore

	// 扩展的路由根聚合
	RootRouter *gin.RouterGroup

	// 扩展的路由api控制器
	McpApi *McpApi
}

// NewOadinExtensionServer 扩展类实例化构造, 初始化父类 OadinCoreServer 的实例
func NewOadinExtensionServer() *OadinExtensionServer {
	// 初始化父类 OadinCoreServer
	coreServer := api.NewOadinCoreServer()

	// 扩展类共享 父类的公共gin实例对象
	return &OadinExtensionServer{
		Router:     coreServer.Router,
		DataStore:  coreServer.DataStore,
		CoreServer: coreServer,
		RootRouter: coreServer.Router.Group("/oadin/" + version.OadinVersion),
	}
}

// Run is the function to start the server
func (e *OadinExtensionServer) Run(ctx context.Context, address string) error {

	// 使用重写覆盖父类run逻辑, 然后扩展类来运行加载所有路由
	return e.Router.Run(address)
}

// Register 扩展注册机制, 保证父类的路由先注册, 然后加载我们的扩展路由
func (e *OadinExtensionServer) Register() {

	// 父类的组件实例化
	e.CoreServer.Register()

	// 扩展类实例化
	e.ExtensionRegister()

	// 注入扩展的路由
	e.injectRouter()
}

// InjectRouter 接口定义了扩展类需要实现的路由注入方法
type InjectRouter interface {
	// InjectRoutes  注入扩展的路由
	InjectRoutes(api *gin.RouterGroup)
}

func (e *OadinExtensionServer) ExtensionRegister() {

	// 扩展API组件实例化
	e.McpApi = NewMcpApi(server.NewMcpService(), e.CoreServer.Playground)

	// 其他扩展组件

}

// injectRoute 注入扩展的路由
func (e *OadinExtensionServer) injectRouter() {
	// Apis related to MCP
	e.McpApi.InjectRoutes(e.RootRouter.Group("/mcp"))

	// 其他扩展路由注册
}
