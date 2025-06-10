package console

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"
)

//go:embed dist/*
var distFS embed.FS

type ServerResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
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
	mux.Handle("/", fileServer)
	// 使用 StripPrefix 处理 /model-manage 路由
	mux.Handle("/model-manage/", http.StripPrefix("/model-manage", fileServer))
	mux.Handle("/server-manage/", http.StripPrefix("/server-manage", fileServer))
	mux.Handle("/mcp-service/", http.StripPrefix("/mcp-service", fileServer))

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
