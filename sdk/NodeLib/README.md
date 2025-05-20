# AogLib使用说明

## 1. 介绍

AogLib 将协助开发者使用 AOG。

## 2. 使用

首先在 NodeJS 项目中安装该 Node Module：


``` sh
npm install aog-lib-1.0.0.tgz
```

然后在项目中引入该 Node Module：

``` JavaScript
const AogLib = require('aog-lib');

const aog = new AogLib();

// check if aog.exe is available
aog.IsAogAvailiable().then((result) => {
    console.log(result);
});

// check if aog.exe is existed
aog.IsAogExisted().then((result) => {
    console.log(result);
});

// download aog.exe
aog.DownloadAog().then((result) => {
    console.log(result);
});

// install aog.exe
aog.InstallAog().then((result) => {
    console.log(result);
});

// run `aog install chat`
aog.InstallChat().then((result) => {
    console.log(result);
});

// aog get services
aog.GetServices().then((result) => {
    console.log(result);
});

// aog install service
const data = {
    service_name: "chat/embed/generate/text-to-image",
    service_source: "remote/local",
    hybrid_policy: "default/always_local/always_remote",
    flavor_name: "ollama/openai/...",
    provider_name: "local_ollama_chat/remote_openai_chat/...",
    auth_type: "none/apikey",
    auth_key: "your_api_key",
}; // required: service_name, service_source, hybrid_policy, flavor_name, provider_name

aog.CreateService(data).then((result) => {
    console.log(result);
});

// aog edit service
const data = {
    service_name: "chat/embed/generate/text-to-image",
    hybrid_policy: "default/always_local/always_remote",
    remote_provider: "",
    local_provider: ""
}; // required: service_name

aog.UpdateService(data).then((result) => {
    console.log(result);
});

// aog get models
aog.GetModels().then((result) => {
    console.log(result);
});

// aog install model
const data = {
    model_name: "llama2",
    service_name: "chat/embed/generate/text-to-image",
    service_source: "remote/local",
    provider_name: "local_ollama_chat/remote_openai_chat/...",
}; // required: model_name, service_name, service_source

aog.InstallModel(data).then((result) => {
    console.log(result);
});

// aog delete model
const data = {
    model_name: "llama2",
    service_name: "chat/embed/generate/text-to-image",
    service_source: "remote/local",
    provider_name: "local_ollama_chat/remote_openai_chat/...",
}; // required: model_name, service_name, service_source

aog.DeleteModel(data).then((result) => {
    console.log(result);
});

// aog get service_providers
aog.GetServiceProviders().then((result) => {
    console.log(result);
});

// aog install serice_provider
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
}; // required: ervice_name, service_source, flavor_name, provider_name
bzye.InstallserviceProvider(data).then((result) => {
    console.log(result);
});

// aog edit serice_provider
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
}; // required: service_name, service_source, flavor_name, provider_name

bzye.updateServiceProvider(data).then((result) => {
    console.log(result);
});

// aog delete serice_provider
const data = {
    provider_name: ""
}; // required: provider_name

aog.DeleteServiceProvider(data).then((reult) => {
    console.log(result);
});

// aog import config
aog.ImportConfig("path/to/.aog").then((result) => {
    console.log(result);
});

// aog export config
const data = {
    service_name: "chat/embed/generate/text-to-image"
};

aog.ExportConfig(data).then((result) => {
    console.log(result);
});

// aog get models available
aog.GetModelsAvailiable().then((result) => {
    console.log(result);
});

// aog get models recommended
aog.GetModelsRecommended().then((result) => {
    console.log(result);
});

// aog get models supported
const data = {
    service_source: "remote/local",
    flavor: "ollama/openai/..." 
}; // required: service_source, flavor
aog.GetModelsSurpported().then((result) => {
    console.log(result);
});

// chat stream
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

aog.Chat(data).then((chatStream) => {
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

// Chat
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

aog.Chat(data).then((result) => {
    console.log(result);
});

// generate stream
const data = {
    model: "deepseek-r1:7b",
    stream: true,
    prompt: "你好",
}
aog.Generate(data).then((generateStream) => {
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

// generate
const data = {
    model: "deepseek-r1:7b",
    stream: false,
    prompt: "你好",
}
aog.Generate(data).then((result) => {
    console.log(result);
});

// text to image
const data = {
    model: "wanx2.1-t2i-turbo",
    prompt: "A beautiful landscape with mountains and a river",
}

aog.TextToImage(data).then((result) => {
    console.log(result);
});

```
