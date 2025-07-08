# Oadin 构建与部署指南

## 目录
- [环境要求](#环境要求)
- [构建 Oadin](#构建-oadin)
- [部署 Oadin 服务](#部署-oadin-服务)
- [应用程序集成](#应用程序集成)
- [部署基于 Oadin 的应用](#部署基于-oadin-的应用)
- [配置管理](#配置管理)
- [故障排除](#故障排除)

## 环境要求

### 开发环境要求

#### 必要组件
- **Go 语言环境**: 1.19+ 版本
  - 下载地址: [https://go.dev/](https://go.dev/)
  - 验证安装: `go version`

#### Windows 开发环境额外要求
- **MSYS2**: 提供 Make 等 Unix 工具
  - 下载地址: [https://www.msys2.org](https://www.msys2.org)
  - 安装后添加到 PATH: `C:\msys64\usr\bin`

#### 运行时要求
- **操作系统**: Windows 10/11, Linux, macOS
- **内存**: 最小 4GB，推荐 8GB+
- **存储**: 最小 2GB 可用空间（用于模型存储）
- **网络**: 用于下载模型和 AI 引擎

### 生产环境要求

#### 硬件要求
- **CPU**: 现代多核处理器
- **内存**: 8GB+ (取决于模型大小)
- **GPU**: 可选，支持 CUDA 的 NVIDIA GPU 可提升性能
- **存储**: SSD 推荐，至少 10GB 可用空间

#### 网络要求
- **入站**: 端口 16688 (Oadin API 服务)
- **出站**: HTTPS (443) 用于下载模型和更新

## 构建 Oadin

### 1. 获取源代码

```bash
# 克隆项目
git clone <repository-url> /path_to_oadin
cd /path_to_oadin
```

### 2. 构建命令

```bash
# 构建所有组件
make build-all

# 仅构建主程序
make build

# 构建特定平台
make build-windows
make build-linux
make build-darwin
```

### 3. 构建产物

构建完成后，将生成以下文件：
```
bin/
├── oadin                 # 主程序 (Linux/macOS)
├── oadin.exe            # 主程序 (Windows)
├── OadinChecker.dll     # Windows 集成组件
├── OadinChecker.so      # Linux 集成组件
└── OadinChecker.dylib   # macOS 集成组件
```

### 4. 验证构建

```bash
# 检查版本
./oadin version

# 检查帮助信息
./oadin -h
```

## 部署 Oadin 服务

### 1. 单机部署

#### 安装步骤

```bash
# 1. 复制可执行文件到系统路径
sudo cp oadin /usr/local/bin/

# 2. 启动 Oadin 服务
oadin server start

# 3. 验证服务状态
curl http://localhost:16688/health
```

#### 系统服务配置 (Linux)

创建 systemd 服务文件：

```bash
# 创建服务文件
sudo tee /etc/systemd/system/oadin.service > /dev/null <<EOF
[Unit]
Description=Oadin AI Gateway Service
After=network.target

[Service]
Type=simple
User=oadin
Group=oadin
ExecStart=/usr/local/bin/oadin server start
ExecStop=/usr/local/bin/oadin server stop
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl enable oadin
sudo systemctl start oadin
sudo systemctl status oadin
```

#### Windows 服务配置

```powershell
# 使用 NSSM 创建 Windows 服务
nssm install Oadin "C:\Program Files\Oadin\oadin.exe"
nssm set Oadin Parameters "server start"
nssm set Oadin DisplayName "Oadin AI Gateway Service"
nssm set Oadin Description "AI PC Open Gateway Service"
nssm start Oadin
```

### 2. 容器化部署

#### Dockerfile 示例

```dockerfile
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/bin/oadin .

EXPOSE 16688

CMD ["./oadin", "server", "start"]
```

#### Docker Compose 部署

```yaml
version: '3.8'

services:
  oadin:
    build: .
    ports:
      - "16688:16688"
    volumes:
      - oadin_data:/root/.oadin
    environment:
      - OADIN_LOG_LEVEL=info
    restart: unless-stopped

volumes:
  oadin_data:
```

### 3. 集群部署

#### 负载均衡配置 (Nginx)

```nginx
upstream oadin_backend {
    server 192.168.1.10:16688;
    server 192.168.1.11:16688;
    server 192.168.1.12:16688;
}

server {
    listen 80;
    server_name oadin.example.com;

    location / {
        proxy_pass http://oadin_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 应用程序集成

### 1. C/C++/C# 应用集成

#### 头文件引入

```c
// 在 main() 函数中引入
#include "OadinChecker.h"

int main() {
    // 初始化 Oadin
    int result = OadinInit();
    if (result != 0) {
        printf("Oadin 初始化失败: %d\n", result);
        return 1;
    }
    
    // 您的应用程序逻辑
    // ...
    
    return 0;
}
```

#### 编译链接

```bash
# Linux/macOS
gcc -o myapp main.c -L. -lOadinChecker

# Windows (使用 MSVC)
cl main.c OadinChecker.lib
```

### 2. .oadin 配置文件

#### 基本配置

```json
{
  "version": "0.2",
  "service": {
    "chat": {
      "models": ["qwen2.5:0.5b", "qwen2.5:7b"]
    },
    "embed": {
      "models": ["nomic-embed-text:latest"]
    }
  }
}
```

#### 完整配置示例

```json
{
  "version": "0.2",
  "metadata": {
    "app_name": "MyAIApp",
    "app_version": "1.0.0",
    "description": "基于 Oadin 的 AI 应用"
  },
  "service": {
    "chat": {
      "models": ["qwen2.5:0.5b", "deepseek-r1:1.5b"],
      "required": true,
      "fallback_model": "qwen2.5:0.5b"
    },
    "embed": {
      "models": ["nomic-embed-text:latest"],
      "required": false
    },
    "text-to-image": {
      "models": ["stable-diffusion-1.5-int4"],
      "required": false
    }
  },
  "deployment": {
    "auto_install": true,
    "check_updates": true,
    "offline_mode": false
  }
}
```

### 3. API 调用示例

#### 原生 Oadin API

```bash
# Chat 服务调用
curl -X POST http://localhost:16688/oadin/v0.2/services/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen2.5:0.5b",
    "messages": [
      {"role": "user", "content": "Hello, world!"}
    ],
    "stream": false
  }'
```

#### OpenAI 兼容 API

```bash
# OpenAI 风格调用
curl -X POST http://localhost:16688/oadin/v0.2/api_flavors/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen2.5:0.5b",
    "messages": [
      {"role": "user", "content": "Hello, world!"}
    ]
  }'
```

#### Ollama 兼容 API

```bash
# Ollama 风格调用
curl -X POST http://localhost:16688/oadin/v0.2/api_flavors/ollama/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen2.5:0.5b",
    "messages": [
      {"role": "user", "content": "Hello, world!"}
    ]
  }'
```

## 部署基于 Oadin 的应用

### 1. 应用打包结构

```
MyAIApp/
├── MyAIApp.exe          # 主程序
├── .oadin               # 服务依赖配置
├── OadinChecker.dll     # Oadin 集成组件
├── config/              # 应用配置
│   └── app.conf
└── README.txt           # 部署说明
```

### 2. 部署脚本示例

#### Windows 部署脚本 (deploy.bat)

```batch
@echo off
echo 正在部署 MyAIApp...

REM 检查 .oadin 文件
if not exist ".oadin" (
    echo 错误: 找不到 .oadin 配置文件
    exit /b 1
)

REM 检查 OadinChecker.dll
if not exist "OadinChecker.dll" (
    echo 错误: 找不到 OadinChecker.dll
    exit /b 1
)

REM 启动应用
echo 启动应用...
MyAIApp.exe

echo 部署完成
pause
```

#### Linux 部署脚本 (deploy.sh)

```bash
#!/bin/bash

echo "正在部署 MyAIApp..."

# 检查依赖文件
if [ ! -f ".oadin" ]; then
    echo "错误: 找不到 .oadin 配置文件"
    exit 1
fi

if [ ! -f "libOadinChecker.so" ]; then
    echo "错误: 找不到 libOadinChecker.so"
    exit 1
fi

# 设置权限
chmod +x MyAIApp

# 启动应用
echo "启动应用..."
./MyAIApp

echo "部署完成"
```

### 3. 安装包制作

#### Windows Installer (NSIS)

```nsis
!include "MUI2.nsh"

Name "MyAIApp"
OutFile "MyAIApp-Setup.exe"
InstallDir "$PROGRAMFILES\MyAIApp"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

Section "Main"
    SetOutPath "$INSTDIR"
    File "MyAIApp.exe"
    File ".oadin"
    File "OadinChecker.dll"
    
    CreateShortcut "$DESKTOP\MyAIApp.lnk" "$INSTDIR\MyAIApp.exe"
SectionEnd
```

#### Linux 软件包 (DEB)

```bash
# 创建包结构
mkdir -p myaiapp-1.0.0/DEBIAN
mkdir -p myaiapp-1.0.0/usr/local/bin
mkdir -p myaiapp-1.0.0/usr/share/myaiapp

# 控制文件
cat > myaiapp-1.0.0/DEBIAN/control << EOF
Package: myaiapp
Version: 1.0.0
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Your Name <your.email@example.com>
Description: My AI Application powered by Oadin
EOF

# 复制文件
cp MyAIApp myaiapp-1.0.0/usr/local/bin/
cp .oadin myaiapp-1.0.0/usr/share/myaiapp/
cp libOadinChecker.so myaiapp-1.0.0/usr/share/myaiapp/

# 构建包
dpkg-deb --build myaiapp-1.0.0
```

## 配置管理

### 1. 服务配置

```bash
# 查看所有服务
oadin get services

# 查看特定服务
oadin get services chat

# 修改服务配置
oadin edit service chat --hybrid_policy always_local
```

### 2. 服务提供商管理

```bash
# 查看服务提供商
oadin get service_providers

# 安装新的服务提供商
oadin install service_provider -f provider.json

# 删除服务提供商
oadin delete service_provider local_ollama_chat
```

### 3. 模型管理

```bash
# 查看已安装模型
oadin get models

# 安装新模型
oadin pull qwen2.5:7b -for chat --provider local_ollama_chat

# 删除模型
oadin delete model qwen2.5:0.5b --provider local_ollama_chat
```

## 故障排除

### 1. 常见问题

#### 服务无法启动

```bash
# 检查端口占用
netstat -tulpn | grep 16688

# 查看日志
oadin server logs

# 重置配置
oadin server reset
```

#### 模型下载失败

```bash
# 检查网络连接
curl -I https://ollama.ai

# 手动下载模型
oadin pull qwen2.5:0.5b -for chat --force

# 清理缓存
rm -rf ~/.oadin/cache
```

#### API 调用失败

```bash
# 检查服务状态
curl http://localhost:16688/health

# 验证服务可用性
oadin get services chat

# 检查模型状态
oadin get models --provider local_ollama_chat
```

### 2. 日志分析

#### 日志位置
- **Linux/macOS**: `~/.oadin/logs/`
- **Windows**: `%USERPROFILE%\.oadin\logs\`

#### 日志级别配置

```bash
# 设置详细日志
export OADIN_LOG_LEVEL=debug
oadin server start

# 实时查看日志
tail -f ~/.oadin/logs/oadin.log
```

### 3. 性能调优

#### 内存优化

```bash
# 设置内存限制
export OADIN_MAX_MEMORY=4G

# 配置模型缓存
oadin config set model_cache_size 2G
```

#### 并发调优

```bash
# 设置最大并发连接
oadin config set max_connections 100

# 设置工作线程数
oadin config set worker_threads 8
```

## 监控与维护

### 1. 健康检查

```bash
# API 健康检查
curl http://localhost:16688/health

# 详细状态检查
curl http://localhost:16688/oadin/v0.2/status
```

### 2. 性能监控

```bash
# 查看资源使用情况
oadin server status

# 查看 API 调用统计
oadin server metrics
```

### 3. 备份恢复

```bash
# 备份配置
tar -czf oadin-backup.tar.gz ~/.oadin/

# 恢复配置
tar -xzf oadin-backup.tar.gz -C ~/
```

---

**注意**: 本文档基于 Oadin v0.2.0 预览版编写，后续版本可能会有变化。请及时关注官方文档更新。