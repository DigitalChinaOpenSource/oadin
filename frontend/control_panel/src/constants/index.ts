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

export const DOWNLOAD_STATUS = {
  // 0：失败  1：下载中   2：下载完成   3. 暂停
  FAILED: 0,
  IN_PROGRESS: 1,
  COMPLETED: 2,
  PAUSED: 3,
};

export const AUTH_TOKEN = 'anythingllm_authToken';

// 本地存储键名常量
export const LOCAL_STORAGE_KEYS = {
  DOWN_LIST: 'downList',
};

export const API_VERSION = 'v0.4';
export const API_PREFIX = `/aog/${API_VERSION}`;
export const API_HEALTH_ENDPOINT = '/health';
