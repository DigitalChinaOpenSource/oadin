package api

import (
	"context"
	"oadin/internal/api"
	"oadin/internal/datastore"
	"oadin/version"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// OadinExtensionServer 使用代理模式来扩展 OadinCoreServer 的功能
type OadinExtensionServer struct {
	// 持有父类实例来实现代理扩展
	CoreServer *api.OADINCoreServer

	// 继承父类属性和方法
	api.OADINCoreServer
	Router    *gin.Engine
	DataStore datastore.Datastore

	// 扩展的路由根聚合
	RootRouter *gin.RouterGroup
}

// NewOadinExtensionServer 扩展类实例化构造, 初始化父类 OadinCoreServer 的实例
func NewOadinExtensionServer() *OadinExtensionServer {
	// 初始化父类 OADINCoreServer
	coreServer := api.NewOADINCoreServer()

	// 扩展类共享 父类的公共gin实例对象
	return &OadinExtensionServer{
		Router:     coreServer.Router,
		DataStore:  datastore.GetDefaultDatastore(),
		CoreServer: coreServer,
		RootRouter: coreServer.Router.Group("/oadin/" + version.OADINVersion),
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

	// 实例化组件并注入扩展的路由
	e.injectRoute()
}

// InjectRouter 接口定义了扩展类需要实现的路由注入方法
type InjectRouter interface {
	// InjectRoutes  注入扩展的路由
	InjectRoutes(api *gin.RouterGroup)
}

var validate = validator.New()

// injectRoute 注入扩展的路由
func (e *OadinExtensionServer) injectRoute() {
	// 创建mcp的API组件实例, 并注入到扩展的路由组中
	mcpApi := NewMcpApi()
	mcpApi.InjectRoutes(e.RootRouter.Group("/mcp"))

	// 其他扩展路由注册
	modelPanelApi := NewModelApi()
	modelPanelApi.InjectRoutes(e.RootRouter.Group("/model_panel"))

	controlPanelApi := NewControlPanelApi()
	controlPanelApi.InjectRoutes(e.RootRouter.Group("/control_panel"))

	playgroundApi := NewPlaygroundApi()
	playgroundApi.InjectRoutes(e.RootRouter.Group("/playground"))

	systemApi := NewSystemlApi()
	systemApi.InjectRoutes(e.RootRouter.Group("/system"))
}
