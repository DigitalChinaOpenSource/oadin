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

package console

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"intel.com/aog/version"
)

//go:embed dist/*
var distFS embed.FS

// 注册前端静态资源和SPA路由
func RegisterConsoleRoutes(engine *gin.Engine) error {
	// 提取dist子文件系统
	distContents, err := fs.Sub(distFS, "dist")
	if err != nil {
		return err
	}

	// 静态资源（如 /assets/xxx.js、/assets/xxx.css）
	assetsFS, err := fs.Sub(distContents, "assets")
	if err != nil {
		return err
	}
	engine.StaticFS("/assets", http.FS(assetsFS))
	engine.StaticFile("/favicon.ico", filepath.Join("dist", "favicon.ico"))

	// SPA路由兜底：所有未命中的GET路由都返回index.html
	engine.NoRoute(func(c *gin.Context) {
		// 包含 /aog/ 并且不包含 /aog/{version} 的路由
		if strings.HasPrefix(c.Request.URL.Path, "/aog/") && !strings.Contains(c.Request.URL.Path, "/aog/"+version.AOGVersion) {
			c.Status(http.StatusNotFound)
			return
		}
		// 其它路由兜底返回index.html
		file, err := distContents.Open("index.html")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		defer file.Close()
		c.Header("Content-Type", "text/html; charset=utf-8")

		// 尝试断言为 io.ReadSeeker
		if seeker, ok := file.(io.ReadSeeker); ok {
			http.ServeContent(c.Writer, c.Request, "index.html", fsStatModTime(file), seeker)
			return
		}
		// 否则读到内存再用 bytes.Reader 包装
		data, err := io.ReadAll(file)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		reader := bytes.NewReader(data)
		http.ServeContent(c.Writer, c.Request, "index.html", fsStatModTime(file), reader)
	})

	return nil
}

// 获取文件的modtime
func fsStatModTime(file fs.File) time.Time {
	info, err := file.Stat()
	if err == nil {
		return info.ModTime()
	}
	return time.Now()
}
