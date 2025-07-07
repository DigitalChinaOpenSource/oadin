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

import { httpRequest } from '@/utils/httpRequest.ts';
import { useRequest } from 'ahooks';
import { useMemo, useEffect } from 'react';
import useServiceStore from '@/store/useServiceStore';
import { message } from 'antd';
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
export type Model = {
  Avatar: string;
  model_name: string;
  provider_name: string;
  status: string;
  service_name: string;
  service_source: string;
  is_default: boolean;
  created_at: string;
  updated_at: string;
};
export const useServices = () => {
  const { installing, setInstalling, setInstalledServices } = useServiceStore();

  const {
    loading,
    data,
    mutate,
    run: getServices,
    runAsync: getServicesAsync,
  } = useRequest<{ services: Service[]; models: Model[] }, any>(() => httpRequest.get('/control_panel/dashboard'), {
    manual: true,
    cacheKey: 'dashboard-services',
    // loadingDelay: 300,
    onSuccess: (result) => {
      setInstalledServices(result.services ? result.services.filter((item) => item.status === 0 || item.status === 1) : []);
    },
  });
  const { run: updateService } = useRequest((data: { service_name: string; hybrid_policy: string }) => httpRequest.put('/service', data), {
    manual: true,
    onSuccess: () => {
      getServicesAsync();
      message.open({
        content: 'Saved successfully',
        type: 'success',
      });
    },
    onError: (error) => {
      console.error('Error updating service:', error);
    },
  });
  const { runAsync: installService } = useRequest(
    (service_name: string) =>
      httpRequest.post('/service/install', {
        service_name,
        skip_model: true,
      }),
    {
      manual: true,
      onSuccess: (result) => {
        // console.log('Service installed successfully', result);
      },
      onError: (error) => {
        // console.error('Error installing service:', error);
      },
    },
  );
  const install = async (service: Service) => {
    setInstalling(true);
    try {
      await installService(service.service_name);
    } catch (error) {
      // console.log('Error installing service:', error);
    } finally {
      getServicesAsync();
      setInstalling(false);
    }
  };
  const unInstalledSerivices = useMemo(() => (data && data.services ? data.services.filter((item) => item.status === -1) : []), [data]);
  const services = useMemo(() => (data && data.services ? data.services.filter((item) => item.status === -1) : []), [data]);
  const models = useMemo(() => (data && data.models ? data.models : []), [data]);
  useEffect(() => {
    getServices();
  }, [getServices]);
  return {
    loading,
    services,
    models,
    getServices,
    getServicesAsync,
    updateService,
    mutate,
    installService,
    install,
    installing,
    unInstalledSerivices,
  };
};
