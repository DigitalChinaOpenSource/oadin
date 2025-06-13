package console

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

//go:embed dist/*
var distFS embed.FS

type ServerResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Router struct {
	Path     string   `json:"path"`
	Children []Router `json:"children"`
}

const RouteJson = `{
  "path": "/",
  "children": [
    {
      "path": "/model-manage",
      "children": [
        { "path": "/model-list" },
        { "path": "/my-model-list" },
        { "path": "/model-experience" }
      ]
    },
    {
      "path": "/settings",
      "children": [
        { "path": "/model-setting" },
        { "path": "/agent-setting" },
        { "path": "/service-provider-manage" },
        { "path": "/about-us" }
      ]
    },
    {
      "path": "/mcp-service",
      "children": [
        { "path": "/mcp-list" },
        { "path": "/my-mcp-list" }
      ]
    },
    {
      "path": "/mcp-detail"
    }
  ]
}`

func initRoute() (*Router, error) {
	var root Router

	err := json.Unmarshal([]byte(RouteJson), &root)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}
	return &root, nil
}

func handleRoute(router *Router, basePath string, mux *http.ServeMux, fileServer http.Handler) {
	// 拼接路径，确保没有重复斜杠
	fullPath := path.Join(basePath, router.Path)

	// 注册当前路径
	mux.Handle(fullPath+"/", http.StripPrefix(fullPath, fileServer))

	// 处理带参数的路由
	if hasPathParams(fullPath) {
		// 为带参数的路由注册一个特殊的处理器
		mux.HandleFunc(fullPath, func(w http.ResponseWriter, r *http.Request) {
			// 提取路径参数
			params := extractPathParams(fullPath, r.URL.Path)
			// 将参数添加到请求上下文中
			ctx := context.WithValue(r.Context(), "pathParams", params)
			// 使用修改后的上下文处理请求
			fileServer.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// 递归注册子路径
	for i := range router.Children {
		handleRoute(&router.Children[i], fullPath, mux, fileServer)
	}
}

// hasPathParams 检查路径是否包含参数
func hasPathParams(path string) bool {
	return strings.Contains(path, ":")
}

// extractPathParams 从URL中提取路径参数
func extractPathParams(pattern, path string) map[string]string {
	params := make(map[string]string)

	// 分割路径模式
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// 确保路径部分数量匹配
	if len(patternParts) != len(pathParts) {
		return params
	}

	// 提取参数
	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			params[paramName] = pathParts[i]
		}
	}

	return params
}

// StartConsoleServer starts the console server
func StartConsoleServer(ctx context.Context) (*http.Server, error) {
	// 从嵌入的文件系统创建子文件系统
	distContents, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, fmt.Errorf("failed to create sub filesystem: %v", err)
	}

	// 创建文件服务器
	fileServer := http.FileServer(http.FS(distContents))

	// 创建路由处理
	mux := http.NewServeMux()
	routeData, err := initRoute()
	if err != nil {
		return nil, err
	}

	handleRoute(routeData, "", mux, fileServer)

	// 创建服务器
	srv := &http.Server{
		Addr:    ":16699",
		Handler: mux,
	}

	// 在新的 goroutine 中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Console server error: %v\n", err)
		}
	}()

	// 监听上下文取消
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("Console server shutdown error: %v\n", err)
		}
	}()

	return srv, nil
}
