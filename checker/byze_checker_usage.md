# AogChecker 使用指南

## 目录

- Go
- Python
- C++
- C#
- Node.js

## Go

### 编译 Go 程序为 DLL

在 Go 程序目录中运行以下命令：

```sh
go build -o AogChecker.dll -buildmode=c-shared AogChecker.go
```

### 使用示例

在 Go 中，您可以直接调用 Go 函数，而不需要通过 DLL。

## Python

### 安装 `ctypes` 库

`ctypes` 是 Python 标准库的一部分，无需额外安装。

### 使用示例

创建一个新的 Python 文件，例如 `testChecker.py`：

```python
from ctypes import CDLL

# 加载 DLL 文件
lib = CDLL('./AogChecker.dll')

# 设置函数原型
lib.ByzeInit.argtypes = []
lib.ByzeInit.restype = None

# 调用 ByzeInit 函数
lib.ByzeInit()
print("ByzeInit called successfully")
```

运行 Python 脚本：

```sh
python testChecker.py
```

## C++

### 创建头文件 AogChecker.h

```cpp
#ifndef ByzeCHECKER_H
#define ByzeCHECKER_H

#ifdef __cplusplus
extern "C" {
#endif

void ByzeInit();

#ifdef __cplusplus
}
#endif

#endif // ByzeCHECKER_H
```

### 使用示例

创建一个新的 C++ 文件，例如 `main.cpp`：

```cpp
#include <iostream>
#include "AogChecker.h"

int main()
{
    // 调用 ByzeInit 函数
    ByzeInit();
    std::cout << "ByzeInit called successfully" << std::endl;
    return 0;
}
```

编译和运行 C++ 程序：

```sh
g++ -o AogCheckerApp main.cpp -L. -lAogChecker
./AogCheckerApp
```

## C#

### 创建包装类

创建一个新的 C# 文件，例如 `AogChecker.cs`：

```csharp
using System;
using System.Runtime.InteropServices;

namespace ByzeChecker
{
    public static class AogChecker
    {
        // 导入 ByzeInit 函数
        [DllImport("AogChecker.dll", CallingConvention = CallingConvention.Cdecl)]
        public static extern void ByzeInit();
    }
}
```

### 使用示例

创建一个新的 C# 控制台应用程序，并在其中引用并调用 Go DLL 中的 `ByzeInit` 函数。

```sh
dotnet new console -n AogCheckerApp
cd AogCheckerApp
```

编辑 `Program.cs` 文件：

```csharp
using System;
using ByzeChecker;

class Program
{
    static void Main(string[] args)
    {
        // 调用 ByzeInit 函数
        AogChecker.ByzeInit();
        Console.WriteLine("ByzeInit called successfully");
    }
}
```

将生成的 `AogChecker.dll` 文件复制到 C# 项目的输出目录（例如 `bin/Debug/net6.0`）。

运行 C# 程序：

```sh
dotnet run
```

## Node.js

### 安装 `ffi-napi` 和 `ref-napi`

此方法要求Node.js版本不高于16。

```sh
npm install ffi-napi ref-napi
```

### 使用示例

创建一个新的 JavaScript 文件，例如 `index.js`：

```javascript
const ffi = require('ffi-napi');
const ref = require('ref-napi');

// 定义返回类型和参数类型
const voidType = ref.types.void;

// 加载 DLL 文件
const AogChecker = ffi.Library('./AogChecker.dll', {
    'ByzeInit': [voidType, []]
});

// 调用 ByzeInit 函数
AogChecker.ByzeInit();
console.log("ByzeInit called successfully");
```

运行 Node.js 脚本：

```sh
node index.js
```

### 安装 node-addon-api

```sh
npm install node-addon-api 
```

按照步骤创建完后，最终的文件结构应该是这样的：

```sh
/your_project_root/
│── /src/                         # 你的 Node.js 项目源码目录
│   ├── test.js                    # 你的调用 `initByze()` 的Node.js 代码
│── /native/                       # 存放 Go 和 C++ 代码的目录
│   ├── AogChecker.go              # Go 代码
│   ├── AogChecker.dll             # Go 编译出的 DLL（Windows 动态库）
│   ├── addon.cpp                  # C++ 代码（Node.js 调用 DLL）
│   ├── binding.gyp                 # Node.js 的 `node-gyp` 配置文件
│   ├── build/                     # 编译 C++ 插件后生成的文件
│       ├── Release/
│           ├── addon.node         # Node.js 可调用的 C++ 插件
│── package.json                   # Node.js 依赖文件
│── node_modules/                   # npm 安装的依赖
```

编辑 `addon.cpp` 文件：

```cpp
#include <napi.h>
#include <windows.h>

typedef void (*ByzeInitFunc)();

HINSTANCE hDLL;
ByzeInitFunc ByzeInit;

Napi::Value InitByze(const Napi::CallbackInfo& info) {
    if (ByzeInit) {
        ByzeInit(); 
    }
    return info.Env().Undefined();
}

Napi::Object Init(Napi::Env env, Napi::Object exports) {
    hDLL = LoadLibrary("path/to/AogChecker.dll");           // 确保替换为 AogChecker.dll 的路径
    if (!hDLL) throw Napi::Error::New(env, "无法加载 AogChecker.dll");

    ByzeInit = (ByzeInitFunc)GetProcAddress(hDLL, "ByzeInit");
    if (!ByzeInit) throw Napi::Error::New(env, "无法找到函数 ByzeInit");

    exports.Set(Napi::String::New(env, "initByze"), Napi::Function::New(env, InitByze));
    return exports;
}

NODE_API_MODULE(addon, Init)
```

编写binding.gyp文件：

```json
{
  "targets": [
    {
      "target_name": "addon",
      "sources": ["addon.cpp"],
      "include_dirs": [
        "<!(node -p \"require('node-addon-api').include\")",
        "path/to/node_modules/node-addon-api"       // 确保替换为 node-addon-api 的路径
      ],
      "dependencies": [
        "<!(node -p \"require('node-addon-api').gyp\")"
      ],
      "defines": ["NODE_ADDON_API_CPP_EXCEPTIONS", "NODE_ADDON_API_DISABLE_CPP_EXCEPTIONS"]
    }
  ]
}
```

切换到 `/native/` 目录下运行以下命令：

```sh
cd native
node-gyp configure
node-gyp build
```

之后执行你的 Node.js 代码：

```sh
node src/test.js
```