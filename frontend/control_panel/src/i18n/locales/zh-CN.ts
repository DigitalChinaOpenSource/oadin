//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

export default {
  errors: {
    '10001': '请检查请求参数是否正确',
    '20001': '请检查请求参数是否正确',
    '30001': '请检查请求参数是否正确',

    '20003': '授权失败，请检查配置授权参数和服务提供商状态',
    '40001': '路径校验失败，请确认路径是否存在或路径填写是否规范',
    '40002': '当前路径下的可用存储空间不足',
    '40003': '当前路径下文件夹不为空',
    '40004': '模型文件迁移失败',
    unknown: '未知错误',
    network: '网络异常，请检查网络连接',
    unavailable: 'AOG服务不可用，请确认AOG服务启动状态',
  },
};
