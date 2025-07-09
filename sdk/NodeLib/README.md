# OADINLib使用说明

## 1. 介绍

OADINLib 将协助开发者使用 OADIN（白泽模型框架）。

现在 OADINLib 提供了以下功能：

检查 oadin 服务是否存在

检查 oadin.exe 是否下载

下载 oadin.exe

## 2. 使用

首先在 NodeJS 项目中安装该 Node Module：


``` sh
npm install oadin-lib-1.0.0.tgz
```

然后在项目中引入该 Node Module：

``` JavaScript
const OADINLib = require('oadin-lib');

const oadin = new OADINLib();

// 检查 oadin 服务是否存在
oadin.isOADINAvailable().then((result) => {
    console.log(result);
});

// 检查 oadin.exe 是否下载
const existed = oadin.isOADINExisted();
console.log(existed);

// 下载 oadin.exe
oadin.downloadOADIN().then((result) => {
    console.log(result);
});

// 启动 oadin 服务
oadin.startOADIN().then((result) => {
    console.log(result);
});

// 查看当前服务
oadin.getServices().then((result) => {
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

oadin.installService(data).then((result) => {
    console.log(result);
});

// 更新服务
const data = {
    service_name: "chat/embed/generate/text-to-image",
    hybrid_policy: "default/always_local/always_remote",
    remote_provider: "",
    local_provider: ""
}; // 必填service_name

oadin.updateService(data).then((result) => {
    console.log(result);
});

// 查看模型
oadin.getModels().then((result) => {
    console.log(result);
});

// 安装模型
const data = {
    model_name: "llama2",
    service_name: "chat/embed/generate/text-to-image",
    service_source: "remote/local",
    provider_name: "local_ollama_chat/remote_openai_chat/...",
}; // 必填model_name, service_name, service_source

oadin.installModel(data).then((result) => {
    console.log(result);
});

// 卸载模型
const data = {
    model_name: "llama2",
    service_name: "chat/embed/generate/text-to-image",
    service_source: "remote/local",
    provider_name: "local_ollama_chat/remote_openai_chat/...",
}; // 必填model_name, service_name, service_source

oadin.deleteModel(data).then((result) => {
    console.log(result);
});

// 查看服务提供商
oadin.getServiceProviders().then((result) => {
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
oadin.installServiceProvider(data).then((result) => {
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

oadin.updateServiceProvider(data).then((result) => {
    console.log(result);
});

// 删除服务提供商
const data = {
    provider_name: ""
};

oadin.deleteServiceProvider(data).then((result) => {
    console.log(result);
});

// 导入配置文件
oadin.importConfig("path/to/.oadin").then((result) => {
    console.log(result);
});

// 导出配置文件
const data = {
    service_name: "chat/embed/generate/text-to-image"
};

oadin.exportConfig(data).then((result) => { // 不填data则导出全部
    console.log(result);
});

// 获取模型列表（查看ollama的模型）
// oadin.getModelsAvailiable() 方法已移除或重命名，请使用 getModels()
oadin.getModels().then((result) => {
    console.log(result);
});

// 获取推荐模型列表
oadin.getModelsRecommended().then((result) => {
    console.log(result);
});

// 获取支持模型列表
const data = {
    service_source: "remote/local",
    flavor: "ollama/openai/..." // local 则默认为ollama
}; // 必填service_source, flavor
oadin.getModelsSupported(data).then((result) => {
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

oadin.chat(data).then((chatStream) => {
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

oadin.chat(data).then((result) => {
    console.log(result);
});

// 生文服务（流式）
const data = {
    model: "deepseek-r1:7b",
    stream: true,
    prompt: "你好",
}
oadin.generate(data).then((generateStream) => {
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
oadin.generate(data).then((result) => {
    console.log(result);
});

// 文生图服务
const data = {
    model: "wanx2.1-t2i-turbo",
    prompt: "一间有着精致窗户的花店，漂亮的木质门，摆放着花朵",
}

oadin.textToImage(data).then((result) => {
    console.log(result);
});

// 语音识别服务
const data = {
    model: "NamoLi/whisper-large-v3-ov",
    audio: "C:/Users/Lenovo/Downloads/中.mp3",
    language: "zh"
}

oadin.speechToText(data).then(response => {
    console.log( response);
});
```
