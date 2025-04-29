本地源添加步骤

```bash
dotnet pack --configuration Release  

mkdir local-nuget

cp ./bin/Release/ByzeClient.1.0.0.nupkg ./local-nuget

# 这一步会把这个目录配置到你的dotnet源列表中，dotnet nuget list source可以查看你的源列表，这之后在任何项目都可以通过--source LocalByze来使用这个源中的包
dotnet nuget add source ./local-nuget --name LocalByze

dotnet add package ByzeClient --version 1.0.0 --source LocalByze
```