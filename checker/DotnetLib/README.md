本地源添加步骤

```bash
dotnet pack --configuration Release  

mkdir local-nuget

cp ./bin/Release/ByzeClient.1.0.0.nupkg ./local-nuget

# 这一步会把这个目录配置到你的dotnet源列表中，dotnet nuget list source可以查看你的源列表，这之后在任何项目都可以通过--source LocalByze来使用这个源中的包
dotnet nuget add source ./local-nuget --name LocalByze

dotnet add package ByzeClient --version 1.0.0 --source LocalByze
dotnet add package ByzeClient --version 1.0.0 --source .
```

```csharp
using ByzeClient;

var client = new ByzeClient();


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

```