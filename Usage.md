### 爱问学端侧AI应用框架 集成文档

oadin 提供平台级的 AI 服务，因此多个共存 AI 应用无需冗余地部署和启动自己的 AI 栈。这显著减少了应用大小，消除了每个应用重复下载相同 AI 栈和模型的情况，并在执行过程中避免了内存消耗的竞争。

## 1. 集成方式

首先在项目中安装 /checker/NodeLib Node Module：


``` sh
npm install oadin-lib-1.0.0.tgz
```

然后在项目中引入和使用：

``` JavaScript
const OadinLib = require('oadin-lib');

const oadin = new OadinLib();

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

```

将应用程序与 ``.oadin`` 文件一同打包发布。

## 2. 配置方式

您可以通过输入  ``oadin -h`` 来查看命令行工具的帮助信息。

使用以下命令启动和停止 oadin 服务

``` bash
    # 前台启动oadin
    oadin server start

    # 后台启动oadin
    oadin server start -d

    # 停止oadin
    oadin server stop
```

作为开发者，可以通过如下命令来快速安装、导入和配置相应的 oadin 服务和服务提供商

```bash

    # 将 AI 服务安装到本地
    # oadin 将安装必要的 AI 堆栈（如 ollama）和 oadin 推荐的模型
    oadin install chat
    oadin install text-to-image

    # 除了默认的模型之外，您可以在服务中安装更多的模型
    # 当前版本暂仅支持基于 ollama 拉取模型
    # v0.3 版本将支持更多的 AI 堆栈和模型，以及其他服务
    oadin pull <model_name> -for <service_name> --provider <provider_name>

    # 获取服务信息，可查看指定服务，未指定则输出全部服务信息
    oadin get services <service_name>


    # 修改服务配置
    # hybrid_policy 指定具体服务的调度策略，可选值有 always_local, always_remote, default
    # remote_provider 指定远程服务提供商
    # local_provider 指定本地服务提供商
    oadin edit service <service_name> --hybrid_policy always_remote --remote_provider xxx --local_provider xxx


    # 获取服务提供商信息，可设置可选参来获取指定服务提供商信息
    oadin get service_providers --service <service_name> --provider <provider_name> --remote <local/remote>

    # 获取模型信息，可设置可选参获取指定模型信息
    oadin get models --provider <provider_name>

    # 安装服务提供商， 安装过程中会自动拉取模型
    oadin install service_provider -f xx/xxx.json
    # 文件名不作要求，内容需为json格式，示例：
    {
        "provider_name": "local_ollama_chat"
        "service_name": "chat",
        "service_source": "local",
        "desc": "Local ollama chat/completion",
        "api_flavor": "ollama",
        "method": "POST",
        "url": "http://localhost:11434/api/chat",
        "auth_type": "none",
        "auth_key": "",
        "models": [
            "qwen2.5:0.5b",
            "qwen2:0.5b"
        ]
    }

    # 修改服务提供商配置，这里仅可修改服务商配置信息，模型变更需通过拉取模型和删除模型来进行
    oadin edit service_provider <provider_name> -f xxx/xxx.json
    # 示例：
    {
        "provider_name": "local_ollama_chat"
        "service_name": "chat",
        "service_source": "local",
        "desc": "Local ollama chat/completion",
        "api_flavor": "ollama",
        "method": "POST",
        "url": "http://localhost:11434/api/chat",
        "auth_type": "none",
        "auth_key": "",
    }

    # 删除服务提供商
    oadin delete service_provider <provider_name>

    # 删除模型 必选参数：--provider
    oadin delete model <model_name>  --provider <provider_name>
```

通过以下命令进行 oadin 服务的导入导出

```bash
    # 根据指定.oadin文件导入服务配置
    oadin import --file xxx/.oadin

    # 导出当前服务配置到指定位置
    # 可选参：
    #   service 指定服务，未指定则导出全部
    #   provider 指定服务提供商，未指定则导出全部
    #   model 指定模型，未指定则导出全部
    oadin export --service chat --provider local_ollama_chat --model --output ./
```

服务导出的 ``.oadin`` 文件可以直接用作导入使用，也就是说，您可以从设备 ``A`` 上导出 ``.oadin`` 文件然后导入到设备 ``B`` 的 ``oadin`` 中，以实现服务的快速共享。

导出的  ``.oadin``  文件示例如下：

```json

    {
        "version": "v0.2",
        "services": {
            "models": {
                "service_providers": {
                    "local": "local_ollama_models",
                },
                "hybrid_policy": "default"
            },
            "chat": {
                "service_providers": {
                    "local": "local_ollama_chat",
                    "remote": "remote_deepseek_chat"
                },
                "hybrid_policy": "default"
            }
        },
        "service_providers": {
            "local_ollama_chat": {
                "service_name": "chat",
                "service_source": "local",
                "desc": "Local ollama chat/completion",
                "api_flavor": "ollama",
                "method": "POST",
                "url": "http://localhost:11434/api/chat",
                "auth_type": "none",
                "auth_key": "",
                "models": [
                    "qwen2.5:0.5b",
                    "qwen2:0.5b"
                ]
            },
            "remote_deepseek_chat": {
                "desc": "remote deepseek chat/completion",
                "service_name": "chat",
                "service_source": "remote",
                "api_flavor": "ollama",
                "method": "POST",
                "url": "https://api.lkeap.cloud.tencent.com/v1/chat/completions",
                "auth_type": "apikey",
                "auth_key": "xxxxxxxxxx",
                "models": [
                    "deepseek-v3",
                    "deepseek-r1"
                ]
            },
            "local_ollama_models": {
                "desc": "List local ollama models",
                "service_name": "models",
                "service_source": "local",
                "api_flavor": "ollama",
                "method": "GET",
                "url": "http://localhost:11434/api/tags",
                "auth_type": "none",
                "auth_key": ""
            }
        }
    }
```

## 3. 服务使用方式

当前oadin预览提供了基本的 chat、text-to-image 等服务。

您可以使用 curl 在 Windows 上测试聊天服务。

```bash

    curl -X POST http://localhost:16688/oadin/v0.2/services/chat  -X POST -H
    "Content-Type: application/json" -d
    "{\"model\":\"deepseek-r1:7b\",\"messages\":[{\"role\":\"user\",\"content\":\"why is
    the sky blue?\"}],\"stream\":false}"
```

此外，如果您已经使用 OpenAI API 或 ollama API 等的应用程序，您无需重写调用 oadin 的方式以符合其规范。

因为 oadin 能够自动转换这些流行风格的 API，因此您只需更改端点 URL，就可以轻松迁移应用程序。

例如，如果您使用的是 OpenAI 的聊天完成服务，您只需将端点 URL 从 ``https://api.openai.com/v1/chat/completions`` 替换为
``http://localhost:16688/oadin/v0.2/api_flavors/openai/v1/chat/completions``。

**NOTE** 请注意，调用 oadin 的新 URL 位于 ``api_flavors/openai`` ，其余 URL 与原始 OpenAI API 相同，即 ``/v1/chat/completions`` 。

如果您使用 ollama API，可以将端点 URL 从 ``https://localhost:11434/api/chat`` 替换为
``http://localhost:16688/oadin/v0.2/api_flavors/ollama/api/chat`` 。同样，它位于 ``api_flavors/ollama`` ，其余 URL 与原始 ollama API 相同，即 ``/api/chat`` 。

