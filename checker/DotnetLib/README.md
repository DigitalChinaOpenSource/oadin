本地源添加步骤

```bash
dotnet pack --configuration Release  

mkdir local-nuget

cp ./bin/Release/OadinClient.1.0.0.nupkg ./local-nuget

# 这一步会把这个目录配置到你的dotnet源列表中，dotnet nuget list source可以查看你的源列表，这之后在任何项目都可以通过--source LocalOadin来使用这个源中的包
dotnet nuget add source ./local-nuget --name LocalOadin

dotnet add package OadinClient --version 1.0.0 --source LocalOadin
dotnet add package OadinClient --version 1.0.0 --source .
```

```csharp
using OadinClient;

var client = new OadinClient();


// 获取服务列表
var services = await client.GetServicesAsync();
Console.WriteLiine(services);

// 创建服务
var requestData = new
{
    service_name = "chat/embed/generate/text-to-image",
    service_source = "remote/local",
    service_provider_name = "local_ollama_chat",
    api_flavor = "ollama/openai/...",
    auth_type = "none/apikey/token/credentials",
    method = "GET/POST",
    desc = "服务描述",
    url = "",
    auth_key = "your_api_key",
    skip_model = false,
    model_name = "llama2",
};
var result = await client.InstallServiceAsync(requestData);
Console.WriteLine(result);

// 更新服务
var requestData = new
{
    service_name = "chat/embed/generate/text-to-image",
    hybrid_policy = "default/always_local/always_remote",
    remote_provider = "remote_openai_chat",
    local_provider = "local_ollama_chat"
};
var result = await client.UpdateServiceAsync(requestData);
Console.WriteLine(result);

// 查看模型
var models = await client.GetModelsAsync();
Console.WriteLine(models);

// 下载模型
var requestData = new
{
    model_name = "llama2",
    service_name = "chat/embed/generate/text-to-image",
    service_source = "remote/local",
    provider_name = "local_ollama_chat/remote_openai_chat/..."
};
var result = await client.InstallModelAsync(requestData);
Console.WriteLine(result);

// 流式下载模型
var requestData = new
{
    model_name = "nomic-embed-text",
    service_name = "embed",
    service_source = "local",
    provider_name = "local_ollama_embed"
};
await client.InstallModelStreamAsync(
    requestData,
    onData: (json) => Console.WriteLine("流数据: " + json),
    onError: (error) => Console.WriteLine("错误: " + error),
    onEnd: () => Console.WriteLine("流式安装完成")
);

// 取消流式下载模型
var requestData = new
{
    model_name = "nomic-embed-text"
};
await client.CancelInstallModelAsync(requestData);

// 卸载模型
var requestData = new
{
    model_name = "llama2",
    service_name = "chat/embed/generate/text-to-image",
    service_source = "remote/local",
    provider_name = "local_ollama_chat/remote_openai_chat/..."
};
var result = await client.DeleteModelAsync(requestData);
Console.WriteLine(result);

// 查看服务提供商
var serviceProviders = await client.GetServiceProvidersAsync();
Console.WriteLine(serviceProviders);

// 新增模型提供商
var requestData = new
{
    service_name = "chat/embed/generate/text-to-image",
    service_source = "remote/local",
    flavor_name = "ollama/openai/...",
    provider_name = "local_ollama_chat/remote_openai_chat/...",
    desc = "提供商描述",
    method = "GET/POST",
    url = "https://api.example.com",
    auth_type = "none/apikey/token/credentials",
    auth_key = "your_api_key",
    models = new[] { "qwen2:7b", "deepseek-r1:7b" },
    extra_headers = new { },
    extra_json_body = new {  },
    properties = new { }
};
var result = await client.AddServiceProviderAsync(requestData);
Console.WriteLine(result);

// 更新模型提供商
var requestData = new
{
    service_name = "chat/embed/generate/text-to-image",
    service_source = "remote/local",
    flavor_name = "ollama/openai/...",
    provider_name = "local_ollama_chat/remote_openai_chat/...",
    desc = "更新后的描述",
    method = "GET/POST",
    url = "https://api.example.com",
    auth_type = "none/apikey/token/credentials",
    auth_key = "your_api_key",
    models = new[] { "qwen2:7b", "deepseek-r1:7b" },
    extra_headers = new { },
    extra_json_body = new {  },
    properties = new { }
};
var result = await client.UpdateServiceProviderAsync(requestData);

// 删除模型提供商
var requestData = new
{
    provider_name = "local_ollama_chat/remote_openai_chat/..."
};
var result = await client.DeleteServiceProviderAsync(requestData);

// 获取模型列表(从引擎)
var models = await client.GetModelAvailiableAsync();
Console.WriteLine(models);

// 获取推荐模型列表
var models = await client.GetModelRecommendedAsync();
Console.WriteLine(models);

// 获取支持模型列表
var models = await client.GetModelSupportedAsync();
Console.WriteLine(models);

//获取问学平台模型列表
var requestData = new
{
    env_type = "dev/product",
};
var models = await client.GetModelListAsync(requestData);
console.WriteLine(models);

// 导入配置文件
var result = await client.ImportConfigAsync("path/to/.oadin");
Console.WriteLine(result);

// 导出配置文件
var result = await client.ExportConfigAsync();
Console.WriteLine(result);

// 流式Chat
var requestData = new { 
    model = "deepseek-r1:7b", 
    stream = true,
    messages = new[] { 
        new { role = "user", content = "你是谁？" } 
    }
};
await client.ChatAsync(
    requestData,
    isStream: true,
    onData: (data) => Console.WriteLine("流数据: " + data),
    onError: (error) => Console.WriteLine("错误: " + error),
    onEnd: () => Console.WriteLine("流式请求结束")
);

// 非流式Chat
var requestData = new { 
    model = "deepseek-r1:7b", 
    stream = false,
    messages = new[] { 
        new { role = "user", content = "你是谁？" } 
    }
};
var result = await client.ChatAsync(requestData);
Console.WriteLine(result);

// embed
var requestData = new { 
    model = "nomic-embed-text",
    imput = new[] { 
        "二彪子", 
        "踹皮" 
    },
};
var result = await client.EmbedAsync(requestData);
Console.WriteLine(result);

// text-to-image
var requestData = new { 
    model = "wanx2.1-t2i-turbo",
    prompt = "喜欢玩埃德加蹲草里攒大招的小学生"
};
var result = await client.TextToImageAsync(requestData);
Console.WriteLine(result);

```