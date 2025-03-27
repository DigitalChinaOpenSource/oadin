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
npm install byze-lib-1.0.2.tgz
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

```

