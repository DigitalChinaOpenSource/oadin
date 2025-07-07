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

import useServerCheckStore from '@/store/useServerCheckStore';

/**
 * 在每次请求前先检查健康状态
 * @param requestFn 需要执行的请求函数
 * @returns Promise<any>
 */
export async function requestWithHealthCheck<T>(requestFn: () => Promise<T>): Promise<T | undefined> {
  const { fetchServerStatus, checkStatus } = useServerCheckStore.getState();

  // 先发起健康检查
  await fetchServerStatus();

  // 检查健康状态
  if (!useServerCheckStore.getState().checkStatus) {
    // 健康检查未通过，直接中断
    return Promise.reject(new Error('服务不可用，已拦截请求'));
  }

  // 健康检查通过，执行后续请求
  return requestFn();
}
