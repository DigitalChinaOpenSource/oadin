# OadinLib使用说明

## 1. 介绍

OadinLib 将协助开发者使用 Oadin（白泽模型框架）。

现在 OadinLib 提供了以下功能：

检查 oadin 服务是否存在

检查 oadin.exe 是否下载

下载 oadin.exe

## 2. 使用

首先在 NodeJS 项目中安装该 Node Module：


``` sh
npm install oadin-lib-1.2.66.tgz
```

然后在项目中引入该 Node Module：

``` JavaScript
const OadinLib = require('oadin-lib');

const oadin = new OadinLib();

// 检查 oadin 服务是否存在
oadin.IsOadinAvailiable().then((result) => {
    console.log(result);
});

// 检查 oadin.exe 是否下载
oadin.IsOadinExisted().then((result) => {
    console.log(result);
});

// 下载 oadin.exe
oadin.DownloadOadin().then((result) => {
    console.log(result);
});

// 启动 oadin 服务
oadin.InstallOadin().then((result) => {
    console.log(result);
});

// 安装 oadin chat服务
oadin.InstallChat().then((result) => {
    console.log(result);
});

// 查看当前服务
oadin.GetServices().then((result) => {
    console.log(result);
});

// 创建新服务
const data = {
    service_name: "chat/embed/generate/text-to-image",
    service_source: "remote/local",
    hybrid_policy: "default/always_local/always_remote",
    flavor_name: "ollama/openai/...",
    provider_name: "local_ollama_chat/remote_openai_chat/...",
    auth_type: "none/apikey",
    auth_key: "your_api_key",
}; // 必填service_name, service_source, hybrid_policy, flavor_name, provider_name

oadin.CreateService(data).then((result) => {
    console.log(result);
});

// 更新服务
const data = {
    service_name: "chat/embed/generate/text-to-image",
    hybrid_policy: "default/always_local/always_remote",
    remote_provider: "",
    local_provider: ""
}; // 必填service_name

oadin.UpdateService(data).then((result) => {
    console.log(result);
});

// 查看模型
oadin.GetModels().then((result) => {
    console.log(result);
});

// 安装模型
const data = {
    model_name: "llama2",
    service_name: "chat/embed/generate/text-to-image",
    service_source: "remote/local",
    provider_name: "local_ollama_chat/remote_openai_chat/...",
}; // 必填model_name, service_name, service_source

oadin.InstallModel(data).then((result) => {
    console.log(result);
});

// 卸载模型
const data = {
    model_name: "llama2",
    service_name: "chat/embed/generate/text-to-image",
    service_source: "remote/local",
    provider_name: "local_ollama_chat/remote_openai_chat/...",
}; // 必填model_name, service_name, service_source

oadin.DeleteModel(data).then((result) => {
    console.log(result);
});

// 查看服务提供商
oadin.GetServiceProviders().then((result) => {
    console.log(result);
});

// 新增模型提供商
const data = {
    service_name: "chat/embed/generate/text-to-image",
    service_source: "remote/local",
    flavor_name: "ollama/openai/...",
    provider_name: "local_ollama_chat/remote_openai_chat/...",
    desc: "",
    method: "",
    auth_type: "none/apikey",
    auth_key: "your_api_key",
    models: ["qwen2:7b", "deepseek-r1:7b", ...],
    extra_headers: {},
    extra_json_body: {},
    properties: {}
}; // 必填service_name, service_source, flavor_name, provider_name
bzye.InstallserviceProvider(data).then((result) => {
    console.log(result);
});

// 更新模型提供商
const data = {
    service_name: "chat/embed/generate/text-to-image",
    service_source: "remote/local",
    flavor_name: "ollama/openai/...",
    provider_name: "local_ollama_chat/remote_openai_chat/...",
    desc: "",
    method: "",
    auth_type: "none/apikey",
    auth_key: "your_api_key",
    models: ["qwen2:7b", "deepseek-r1:7b", ...],
    extra_headers: {},
    extra_json_body: {},
    properties: {}
}; // 必填service_name, service_source, flavor_name, provider_name

bzye.updateServiceProvider(data).then((result) => {
    console.log(result);
});

// 删除服务提供商
const data = {
    provider_name: ""
};

oadin.DeleteServiceProvider(data).then((reult) => {
    console.log(result);
});

// 导入配置文件
oadin.ImportConfig("path/to/.oadin").then((result) => {
    console.log(result);
});

// 导出配置文件
const data = {
    service_name: "chat/embed/generate/text-to-image"
};

oadin.ExportConfig(data).then((result) => { // 不填data则导出全部
    console.log(result);
});

// 获取模型列表（查看ollama的模型）
oadin.GetModelsAvailiable().then((result) => {
    console.log(result);
});

// 获取推荐模型列表
oadin.GetModelsRecommended().then((result) => {
    console.log(result);
});

// 获取支持模型列表
const data = {
    service_source: "remote/local",
    flavor: "ollama/openai/..." // local 则默认为ollama
}; // 必填service_source, flavor
oadin.GetModelsSurpported().then((result) => {
    console.log(result);
});

// Chat服务（流式）
const data = {
    model: "deepseek-r1:7b",
    stream: true,
    messages: [
        {
            role: "user",
            content: "你好"
        }
    ],
    temperature: 0.7,
    max_tokens: 100,
}

oadin.Chat(data).then((chatStream) => {
    chatStream.on('data', (data) => {
        console.log(data);
    });
    chatStream.on('error', (error) => {
        console.error(error);
    });
    chatStream.on('end', () => {
        console.log('Chat stream ended');
    });
});

// Chat服务（非流式）
const data = {
    model: "deepseek-r1:7b",
    stream: false,
    messages: [
        {
            role: "user",
            content: "你好"
        }
    ],
    temperature: 0.7,
    max_tokens: 100,
}

oadin.Chat(data).then((result) => {
    console.log(result);
});

// 生文服务（流式）
const data = {
    model: "deepseek-r1:7b",
    stream: true,
    prompt: "你好",
}
oadin.Generate(data).then((generateStream) => {
    generateStream.on('data', (data) => {
        console.log(data);
    });
    generateStream.on('error', (error) => {
        console.error(error);
    });
    generateStream.on('end', () => {
        console.log('Generate stream ended');
    });
});

// 生文服务（非流式）
const data = {
    model: "deepseek-r1:7b",
    stream: false,
    prompt: "你好",
}
oadin.Generate(data).then((result) => {
    console.log(result);
});

// 文生图服务
const data = {
    model: "wanx2.1-t2i-turbo",
    prompt: "一间有着精致窗户的花店，漂亮的木质门，摆放着花朵",
}

oadin.TextToImage(data).then((result) => {
    console.log(result);
});

## 8. Playground 功能

Oadin Playground 功能提供了一个与下载的 AI 模型交互的聊天界面，包含以下特性：

- 会话管理
- 消息发送和接收
- 文档处理和 RAG（检索增强生成）功能
- 流式响应
- 思考模式

### 8.1 创建会话

```javascript
const createSessionData = {
    title: "测试会话",
    modelId: "deepseek-r1:7b",
    embedModelId: "deepseek-r1:7b",
};

oadin.CreatePlaygroundSession(createSessionData).then((result) => {
    console.log(result);
});
```

### 8.2 获取会话列表

```javascript
oadin.GetPlaygroundSessions().then((result) => {
    console.log(result);

});
```

### 8.3 发送消息

```javascript
const messageData = {
    sessionId: "session-uuid",
    content: "Hello, I'd like to chat about AI."
};

oadin.SendPlaygroundMessage(messageData).then((result) => {
    console.log(result);

});
```

### 8.4 发送流式消息

```javascript
const messageData = {
    sessionId: "session-uuid",
    content: "Tell me about neural networks"
};

const stream = oadin.SendPlaygroundMessageStream(messageData);

// 使用事件监听器接收流式数据
stream.on('data', (chunk) => {
    console.log('收到数据片段:', chunk.content);
    // 处理流式响应的每个数据片段
});

stream.on('complete', (finalMessage) => {
    console.log('完整消息:', finalMessage);
    // 处理完整的最终消息
});

stream.on('error', (error) => {
    console.error('错误:', error);
});

stream.on('end', () => {
    console.log('流结束');
});
```

### 8.5 获取消息历史

```javascript
const sessionId = "session-uuid";

oadin.GetPlaygroundMessages(sessionId).then((result) => {
    console.log(result);
});
```

### 8.6 文件上传与处理（RAG功能）

```javascript
// 上传文件
const sessionId = "session-uuid";
const filePath = "path/to/document.txt";

oadin.UploadPlaygroundFile(sessionId, filePath).then((result) => {
    console.log('文件上传结果:', result);
    
    if (result.code === 200 && result.data) {
        // 处理文件（生成嵌入）
        const fileId = result.data.id;
        const model = "deepseek-r1:7b";  // 用于生成嵌入的模型
        
        return oadin.ProcessPlaygroundFile(fileId, model);
    }
}).then((result) => {
    console.log('文件处理结果:', result);
});
```

### 8.7 获取文件列表

```javascript
const sessionId = "session-uuid";

oadin.GetPlaygroundFiles(sessionId).then((result) => {
    console.log(result);
});
```

### 8.8 删除文件

```javascript
const fileId = "file-uuid";

oadin.DeletePlaygroundFile(fileId).then((result) => {
    console.log(result);
});
```

### 8.9 切换会话模型

```javascript
oadin.ChangePlaygroundSessionModel({
  sessionId: 'your-session-id',
  modelId: 'new-model-id',
  embedModelId: 'your-embed-model-id'
});
```

### 8.10 删除会话

```javascript
const sessionId = "session-uuid";

oadin.DeletePlaygroundSession(sessionId).then((result) => {
    console.log(result);
});
```
