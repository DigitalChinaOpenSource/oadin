# Byze Quick Start

Byze 是一个模型框架，它能解耦 AI PC 上的 AI 应用与它所依赖的 AI 服务。它旨在为开发者提供一个 极其简单易用的基础设施，以便他们在开发环境中安装本地 AI 服务，并发布他们的 AI 应用程序，无需打包自己 的 AI 堆栈和模型。

### 系统要求
- windows 10 64位及以上版本/ macOS 14及以上版本

### 下载和启动

你可以前往 http://120.232.136.73:31619/byzedev/byze.exe 下载 `byze.exe` 到你的用户目录下，并添加到环境变量。

用以下命令启动 Byze 服务：

```bash
byze server start -d
```

Byze 启动时会同时检测本地是否安装了 `ollama` ，如果没有安装会弹出下载和安装的提示，你可以选择是否安装，确认安装即可。

### 检查当前 byze 是否启动

通过 api 即向本地的16688端口发送GET请求可以检查 byze 是否正常。响应如下则正常：
```
Open Platform for AIPC
```

### 快速使用服务( 以 `chat` 服务为例)

以下将帮助你更快地使用 byze 的服务，在使用中熟悉 byze。如果你想定制化地使用，可以前往[通过多步骤操作安装chat服务](#通过多步骤操作安装chat服务)。

接下来我们以 `chat` 服务为例，从安装开始来使用服务。

首先通过 api `/service` 一键安装 `chat` 服务

```
POST http:127.0.0.1:16688/byze/v0.2/service
```

请求参数如下：

```json
{
    "service_name": "chat", // 必填，服务名，如chat /embed /generate /text-to-image
    "service_source": "remote", // 必填，服务类型，如remote /local
    "service_provider_name": "local_ollama_chat", // 必填，服务提供商名，如local_ollama_chat /local_ollama_generate /remote_openai_chat /remote_tencent_embed /remote_baidu_text-to-image
    "api_flavor": "ollama", // 必填，接口风格，如ollamao /openai /tencent /baidu /deepseek /smartvision /aliyun
    "auth_type": "none/apikey", // 选填，鉴权类型apikey /token /credentials /none
    "auth_key": "", // 选填，服务提供商的鉴权信息
    "method": "POST", // 选填，服务请求方法，默认为POST
    "desc": "", // 选填，服务描述
    "url": "", // 选填，服务提供商的URL
    "skip_model": false, // 选填，安装服务时会默认一起安装一个由 byze 默认推荐的模型，选择是否跳过模型安装，默认为false
    "model_name": "llama2", // 选填，模型名，如果不跳过安装，你可以指定安装某个模型，不指定则安装默认模型 deepseek-r1:7b
}
```

例如我们用这样的请求，我们可以安装一个`chat`服务，同时为它安装一个 deepseek-r1:7b：
```json
{
    "service_name": "chat",
    "service_source": "local",
    "service_provider_name": "local_ollama_chat",
    "flavor_name": "ollama",
}

```

响应如下：
```json
{
    "business_code": 10000,
    "message": "service interface call success"
}
```

然后我们来了解一下 `chat` 服务的使用方法。

`chat` 服务同样通过 api 调用

```
POST http:127.0.0.1:16688/byze/v0.2/services/chat
```

请求参数如下：
```json
{
    "model": "", // 选填，模型名称
    "messages": "", // 必填，会话列表
    "stream": true, // 选填，是否流式输出，默认为true
}
```

例如我们用这样的请求，可以通过 byze 请求模型的 chat：
```json
{
    "model": "deepseek-r1:7b",
    "stream": true,
    "messages": [
        {
            "role": "user",
            "content": "你好！"
        },
        {
            "role": "assistant",
            "content": "你好！很高兴见到你！"
        },
        {
            "role": "user",
            "content": "你是谁？"
        }
    ]
}
```

流式响应如下：
```json
data:{
    "model": "deepseek-r1:7b",
    "created_at": "2025-04-25T06:53:34.9488435Z",
    "message": {
        "role": "assistant",
        "content": "<think>"
    },
    "done": false
}
data:{
    "model": "deepseek-r1:7b",
    "created_at": "2025-04-25T06:53:35.2907313Z",
    "message": {
        "role": "assistant",
        "content": "我是"
    },
    "done": false
}
......
data:{
    "model": "deepseek-r1:7b",
    "created_at": "2025-04-25T06:53:49.573032Z",
    "message": {
        "role": "assistant",
        "content": "。"
    },
    "done": false
}
data:{
    "model": "deepseek-r1:7b",
    "created_at": "2025-04-25T06:53:49.7700738Z",
    "message": {
        "role": "assistant",
        "content": ""
    },
    "done_reason": "stop",
    "done": true,
    "total_duration": 24976937400,
    "load_duration": 8490355100,
    "prompt_eval_count": 17,
    "prompt_eval_duration": 1258000000,
    "eval_count": 79,
    "eval_duration": 14828000000
}
```

接着就可以在你的应用程序代码中替换 chat() 函数的url使用服务了，下面给出了用 byze 提供的 url 替换原代码的示例：

```javascript
async function chat(messages) {
    // 原代码
    const response = await fetch('http://127.0.0.1:11434/api/chat', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            model: "deepseek-r1:7b",
            stream: true,
            messages: messages
        }),
    });

    // 修改后
    const response = await fetch('http://127.0.0.1:16688/byze/v0.2/services/chat', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            model: "deepseek-r1:7b",
            stream: true,
            messages: messages
        }),
    });

    // 原有处理流式响应
    const eventEmitter = new EventEmitter();
    res.data.on('data', (chunk) => {
    try {
        const rawData = chunk.toString().trim();
        const jsonString = rawData.startsWith('data:') ? rawData.slice(5) : rawData;
        const response = JSON.parse(jsonString);
        eventEmitter.emit('data', response);  // 触发事件，实时传输数据
    } catch (err) {
        eventEmitter.emit('error', `解析流数据失败: ${err.message}`);
    }
    });

    res.data.on('error', (err) => {
    eventEmitter.emit('error', `流式响应错误: ${err.message}`);
    });

    res.data.on('end', () => {
    eventEmitter.emit('end');  // 触发结束事件
    });

    return eventEmitter;
}
```

### 通过导入配置一键安装多个服务

同时，Byze 也支持导入配置文件，导入配置文件功能可以帮助你快速导入多个服务和模型配置。配置文件(.byze)是一个json格式的文件，示例：
```json
{
  "version": "v0.2",
  "services": { // 当前包含的服务信息
    "chat": {
      "service_providers": {
        "local": "local_ollama_chat",
        "remote": "remote_smartvision_chat"
      },
      "hybrid_policy": "always_local"
    },
    "embed": {
      "service_providers": {
        "local": "local_ollama_embed",
        "remote": "remote_smartvision_embed"
      },
      "hybrid_policy": "always_local"
    },
    "text_to_image": {
      "service_providers": {
        "remote": "remote_aliyun_text_to_image"
      },
      "hybrid_policy": "always_remote"
    }
  },
  "service_providers": { // 服务提供商的详细信息
    "local_ollama_chat": {
      "service_name": "chat",
      "service_source": "local",
      "desc": "Local ollama chat/completion",
      "api_flavor": "ollama",
      "method": "POST",
      "auth_type": "none",
      "auth_key": "",
      "models": [
      ]
    },
    "local_ollama_embed": {
      "desc": "Local ollama embed",
      "service_name": "embed",
      "service_source": "local",
      "api_flavor": "ollama",
      "method": "POST",
      "auth_type": "none",
      "auth_key": "",
      "models": []
    },
    "remote_smartvision_chat": {
      "desc": "Remote smartVision chat",
      "service_name": "chat",
      "service_source": "remote",
      "api_flavor": "smartvision",
      "method": "POST",
      "auth_type": "none",
      "auth_key": "{\\\"神州数码|DeepSeek-R1\\\":{\\\"provider\\\":\\\"dcmodel\\\",\\\"model_key\\\":\\\"DeepSeek-R1\\\",\\\"env_type\\\":\\\"production\\\",\\\"credentials\\\":{\\\"api_key\\\":\\\"aGmXnTbWqLpZvYrKsDfVcJhQ\\\",\\\"endpoint_url\\\":\\\"http://120.232.136.137:8100/v1\\\"}}}",
      "models": ["神州数码|DeepSeek-R1"]
    },
    "remote_smartvision_embed": {
      "desc": "Remote smartVision embed",
      "service_name": "embed",
      "service_source": "remote",
      "api_flavor": "smartvision",
      "method": "POST",
      "auth_type": "none",
      "auth_key": "",
      "models": []
    },
    "remote_aliyun_text_to_image": {
      "desc": "Remote aliyun tti",
      "service_name": "text_to_image",
      "service_source": "remote",
      "api_flavor": "aliyun",
      "method": "POST",
      "auth_type": "none",
      "auth_key": "",
      "models": []
    }
  }
}
```

可以通过 api 调用
```
POST http:127.0.0.1:16688/byze/v0.2/service/import
```

将配置文件作为请求，得到的响应示例：
```json
{
    "business_code": 0,
    "message": ""
}
```

### 通过多步骤操作安装chat服务

通过以下的步骤，你可以更灵活地指定模型的调用。

服务提供商的作用是提供模型的调用，例如 
`local_ollama_chat` 可以从本地的 `ollama` 拉取 `chat` 模型；
`remote_baidu_text-to-image` 可以从远程调用 `baidu` 的 `text-to-image` 模型......
所以需要安装服务提供商后才能拉取和调用模型。

#### 安装服务商

首先，你需要在 Byze 中安装一个服务提供商，例如 `local_ollama_chat`，你可以通过 api 调用：
```
POST http:127.0.0.1:16688/byze/v0.2/service_provider
```
请求参数如下：
```json
{
    "service_name": "chat/embed/generate/text-to-image", // 必填，服务名，如chat /embed /generate /text-to-image
    "service_source": "remote/local", // 必填，服务来源，如remote /local
    "api_flavor": "ollama/openai/...", // 必填，服务厂商名称，如ollamao /openai /tencent /baidu /deepseek /smartvision /aliyun
    "provider_name": "local_ollama_chat/remote_openai_chat/...", // 必填，服务提供商名称，如local_ollama_chat /remote_openai_chat /remote_tencent_embed /remote_baidu_text-to-image
    "desc": "", // 选填，服务描述
    "method": "", // 选填，服务请求方法，默认为POST
    "auth_type": "none/apikey", // 选填，鉴权类型none/apikey/token/credentials
    "auth_key": "your_api_key", // 选填，服务提供商的鉴权信息
    "models": ["qwen2:7b", "deepseek-r1:7b", ...], // 选填，服务提供商的模型列表
    "extra_headers": {}, // 选填，服务提供商的额外请求头
    "extra_json_body": {}, // 选填，服务提供商的额外请求体
    "properties": {} // 选填，服务提供商的额外属性
}
```

例如我们用这样的请求，我们可以安装一个`local_ollama_chat`服务提供商：
```json
{
    "service_name": "chat",
    "service_source": "local",
    "api_flavor": "ollama",
    "provider_name": "local_ollama_chat",
}
```
#### 拉取模型

例如 `local_ollama_chat`，你可以通过 api 调用
```
POST http:127.0.0.1:16688/byze/v0.2/model
```
请求参数
```json
{
    "model_name": "llama2", // 必填，模型名称
    "service_name": "chat/embed/generate/text-to-image", // 必填，服务名，如chat /embed /generate /text-to-image
    "service_source": "remote/local", // 必填，服务来源，如remote /local
    "provider_name": "local_ollama_chat/remote_openai_chat/...", // 选填 /remote_openai_chat /remote_tencent_embed /remote_baidu_text-to-image
}
```
响应如下：
```json
{
    "business_code": 30000,
    "message": "service interface call success"
}
```
接下来你就可以使用模型了。

### 停止 byze 服务

在终端运行以下命令即可：
```bash
byze server stop
```