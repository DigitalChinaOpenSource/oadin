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

import { create } from 'zustand';
import { healthRequest } from '@/utils/httpRequest';
import { message } from 'antd';

interface HealthCheckState {
  checkStatus: boolean; // 服务健康状态
  checkServerLoading: boolean; // 请求加载状态
  setCheckServerLoading: (status: boolean) => void;
  fetchServerStatus: () => Promise<boolean>; // 手动触发健康检查
}

const useServerCheckStore = create<HealthCheckState>((set) => ({
  checkStatus: true,
  checkServerLoading: false,

  setCheckServerLoading: (loading: boolean) => {
    set({ checkServerLoading: loading });
  },

  fetchServerStatus: async () => {
    set({ checkServerLoading: true });
    try {
      const data = await healthRequest.get('/health');
      if (data?.status === 'UP') {
        set({ checkStatus: true });
        return true;
      } else {
        set({ checkStatus: false });
        return false;
      }
    } catch (error) {
      set({ checkStatus: false });
      return false;
    } finally {
      set({ checkServerLoading: false });
    }
  },
}));

export default useServerCheckStore;
