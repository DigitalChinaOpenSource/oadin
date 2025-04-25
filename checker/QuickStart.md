# Byze Quick Start

Byze 是一个模型框架，它能耦 AI PC 上的 AI 应用与它所依赖的 AI 服务。它旨在为开发者提供一个 极其简单易用的基础设施，以便他们在开发环境中安装本地 AI 服务，并发布他们的 AI 应用程序，无需打包自己 的 AI 堆栈和模型。

### 下载和启动

你可以前往 http://120.232.136.73:31619/byzedev/byze.exe 下载 `byze.exe` 到你的用户目录下，并添加到环境变量。

用以下命令启动 Byze 服务：

```bash
byze server start -d
```

如弹框下载和安装 `ollama` ，确认安装即可。

### 快速使用服务( 以 `chat` 服务为例)

接下来我们以chat服务为例，从安装开始来使用服务。

首先通过 api `/service` 安装 `chat` 服务

```
POST /service
```

请求参数如下：

```json
{
    "service_name": "chat", // 必填，服务名，如chat /embed /generate /text-to-image
    "service_source": "remote", // 必填，服务类型，如remote /local
    "service_provider_name": "local_ollama_chat", // 必填，服务提供商名，如local_ollama_chat /local_ollama_generate /remote_openai_chat /remote_tencent_embed /remote_baidu_text-to-image
    "flavor_name": "ollama", // 必填，接口风格，如ollamao /openai /tencent /baidu /deepseek /smartvision /aliyun
    "auth_type": "none/apikey", // 选填，鉴权类型apikey /token /credentials /none
    "auth_key": "", // 选填，服务提供商的鉴权信息
    "method": "POST", // 选填，服务请求方法，默认为POST
    "desc": "", // 选填，服务描述
    "url": "", // 选填，服务提供商的URL
    "skip_model": false, // 选填，安装服务时会默认一起安装一个由 byze 设置的模型，选择是否跳过模型安装，默认为false
    "model_name": "llama2", // 选填，模型名，如果不跳过安装，你可以选择一个模型一起安装
}
```

例如我们用这样的请求，我们可以安装一个`chat`服务，同时为它安装一个 deepseek-r1:7b：
```json
{
    "service_name": "chat",
    "service_source": "remote",
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
POST /services/chat
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

接着就可以在你的应用程序代码中替换 chat() 函数的url使用服务了

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