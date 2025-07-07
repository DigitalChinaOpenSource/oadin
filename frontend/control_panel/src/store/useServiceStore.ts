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
import { persist, createJSONStorage } from 'zustand/middleware';
export type Service = {
  avatar: string;
  service_name: string; // 服务名称
  hybrid_policy: string; // 混合策略
  remote_provider: string; // 远程提供者
  local_provider: string; // 本地提供者
  status: number; // 服务状态
  created_at: string; // 创建时间
  updated_at: string; // 更新时间
};
interface ServiceStore {
  installing: boolean;
  setInstalling: (val: boolean) => void;
  installedServices: Service[];
  setInstalledServices: (services: Service[]) => void;
}

const useServiceStore = create<ServiceStore>()(
  persist(
    (set) => ({
      installedServices: [],
      installing: false,
      setInstalling: (val: boolean) => set({ installing: val }),
      setInstalledServices: (services: Service[]) => set({ installedServices: services }),
    }),
    {
      name: 'service-storage',
      storage: createJSONStorage(() => sessionStorage),
      partialize: (state) => ({ installedServices: state.installedServices }),
    },
  ),
);

export default useServiceStore;
