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

import CustomBreadcrumb from '@/components/custom-breadcrumb';
import { Flex, Radio, Button, Spin } from 'antd';
import { useState, useEffect } from 'react';
import { useServices } from '@/hooks/useServices/index.ts';
import { useLocation, useParams } from 'react-router-dom';
import useServiceStore from '@/store/useServiceStore';
export default function HybridScheduling() {
  const location = useLocation();
  const params = useParams();
  const { installedServices } = useServiceStore();
  const [hybridScheduling, setHybridScheduling] = useState('');
  const service_name = params.type as string;
  const { loading, services, updateService } = useServices();
  useEffect(() => {
    const hybrid_policy = installedServices.find((service) => service.service_name === service_name)?.hybrid_policy;
    setHybridScheduling(hybrid_policy || '');
  }, [installedServices, service_name]);
  return (
    <Spin spinning={loading}>
      <Flex
        vertical
        gap={24}
        align="flex-start"
      >
        <CustomBreadcrumb />
        <Radio.Group
          style={{ display: 'flex', gap: 12 }}
          value={hybridScheduling}
          onChange={(e) => {
            setHybridScheduling(e.target.value);
          }}
        >
          <Radio
            style={{ width: '172px' }}
            value="default"
          >
            auto switch
          </Radio>
          <Radio
            style={{ width: '172px' }}
            value="always_local"
          >
            always local
          </Radio>
          <Radio
            style={{ width: '172px' }}
            value="always_remote"
          >
            always remote
          </Radio>
        </Radio.Group>
        <Button
          type="primary"
          size="large"
          onClick={() => {
            if (hybridScheduling) {
              updateService({ service_name, hybrid_policy: hybridScheduling });
            }
          }}
        >
          Save
        </Button>
      </Flex>
    </Spin>
  );
}
