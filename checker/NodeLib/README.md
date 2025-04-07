# ByzeLib使用说明

## 1. 介绍

ByzeLib 将协助开发者使用 Byze（白泽模型框架）。

现在 ByzeLib 提供了以下功能：

检查 byze 服务是否存在

检查 byze.exe 是否下载

下载 byze.exe

## 2. 使用

首先在 NodeJS 项目中安装该 Node Module：


``` sh
npm install byze-lib-1.0.0.tgz
```

然后在项目中引入该 Node Module：

``` JavaScript
const ByzeLib = require('byze-lib');

const byze = new ByzeLib();

// 检查 byze 服务是否存在
byze.IsByzeAvailiable().then((result) => {
    console.log(result);
});

// 检查 byze.exe 是否下载
byze.IsByzeExisted().then((result) => {
    console.log(result);
});

// 下载 byze.exe
byze.DownloadByze().then((result) => {
    console.log(result);
});

// 启动 byze 服务
byze.InstallByze().then((result) => {
    console.log(result);
});

// 安装 byze chat服务
byze.InstallChat().then((result) => {
    console.log(result);
});

// 查看当前服务
byze.GetServices().then((result) => {
    console.log(result);
});

// 创建新服务
const data = {
    service_name = "chat/embed/generate/text-to-image",
    service_source = "remote/local",
    hybrid_policy = "default/always_local/always_remote",
    flavor_name = "ollama/openai/...",
    provider_name = "local_ollama_chat/remote_openai_chat/...",
    auth_type = "none/apikey",
    auth_key = "your_api_key",
}; // 必填service_name, service_source, hybrid_policy, flavor_name, provider_name

byze.CreateService(data).then((result) => {
    console.log(result);
});

// 更新服务
const data = {
    service_name = "chat/embed/generate/text-to-image",
    hybrid_policy = "default/always_local/always_remote",
    remote_provider = "",
    local_provider = ""
}; // 必填service_name

byze.UpdateService(data).then((result) => {
    console.log(result);
});

// 查看模型
byze.GetModels().then((result) => {
    console.log(result);
});

// 安装模型
const data = {
    model_name = "llama2",
    service_name = "chat/embed/generate/text-to-image",
    service_source = "remote/local",
    provider_name = "local_ollama_chat/remote_openai_chat/...",
}; // 必填model_name, service_name, service_source

byze.InstallModel(data).then((result) => {
    console.log(result);
});

// 卸载模型
const data = {
    model_name = "llama2",
    service_name = "chat/embed/generate/text-to-image",
    service_source = "remote/local",
    provider_name = "local_ollama_chat/remote_openai_chat/...",
}; // 必填model_name, service_name, service_source

byze.DeleteModel(data).then((result) => {
    console.log(result);
});

// 查看服务提供商
byze.GetServiceProviders().then((result) => {
    console.log(result);
});

// 新增模型提供商
const data = {
    service_name = "chat/embed/generate/text-to-image",
    service_source = "remote/local",
    flavor_name = "ollama/openai/...",
    provider_name = "local_ollama_chat/remote_openai_chat/...",
    desc = "",
    method = "",
    auth_type = "none/apikey",
    auth_key = "your_api_key",
    models = ["qwen2:7b", "deepseek-r1:7b", ...],
    extra_headers = {},
    extra_json_body = {},
    properties = {}
}; // 必填service_name, service_source, flavor_name, provider_name
bzye.InstallserviceProvider(data).then((result) => {
    console.log(result);
});

// 更新模型提供商
const data = {
    service_name = "chat/embed/generate/text-to-image",
    service_source = "remote/local",
    flavor_name = "ollama/openai/...",
    provider_name = "local_ollama_chat/remote_openai_chat/...",
    desc = "",
    method = "",
    auth_type = "none/apikey",
    auth_key = "your_api_key",
    models = ["qwen2:7b", "deepseek-r1:7b", ...],
    extra_headers = {},
    extra_json_body = {},
    properties = {}
}; // 必填service_name, service_source, flavor_name, provider_name

bzye.updateServiceProvider(data).then((result) => {
    console.log(result);
});

// 删除服务提供商
const data = {
    provider_name = ""
};

byze.DeleteServiceProvider(data).then((reult) => {
    console.log(result);
});

// 导入配置文件
byze.ImportConfig("path/to/.byze").then((result) => {
    console.log(result);
});

// 导出配置文件
const data = {
    service_name = "chat/embed/generate/text-to-image"
};

byze.ExportConfig(data).then((result) => { // 不填data则导出全部
    console.log(result);
});

// 获取模型列表
byze.GetModelsAvailiable().then((result) => {
    console.log(result);
});

// 获取推荐模型列表
byze.GetModelsRecommended().then((result) => {
    console.log(result);
});

// 获取支持模型列表
byze.GetModelsSurpported().then((result) => {
    console.log(result);
});

// Chat服务（流式）
const data = {
    model = "deepseek-r1:7b",
    stream = true,
    messages = [
        {
            role = "user",
            content = "你好"
        }
    ],
    temperature = 0.7,
    max_tokens = 100,
}

byze.Chat(data).then((chatStream) => {
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
    model = "deepseek-r1:7b",
    stream = false,
    messages = [
        {
            role = "user",
            content = "你好"
        }
    ],
    temperature = 0.7,
    max_tokens = 100,
}

byze.Chat(data).then((result) => {
    console.log(result);
});

// 生文服务（流式）
const data = {
    model = "deepseek-r1:7b",
    stream = true,
    prompt = "你好",
}
byze.Generate(data).then((generateStream) => {
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
    model = "deepseek-r1:7b",
    stream = false,
    prompt = "你好",
}
byze.Generate(data).then((result) => {
    console.log(result);
});

// 文生图服务
const data = {
    model = "wanx2.1-t2i-turbo",
    prompt = "一间有着精致窗户的花店，漂亮的木质门，摆放着花朵",
}

byze.TextToImage(data).then((result) => {
    console.log(result);
});

```



