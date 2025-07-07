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

import styles from './index.module.scss';
import { Flex } from 'antd';
import localModels from '@/assets/本地模型.svg';
import remoteModels from '@/assets/数据模型.svg';
import hybridScheduling from '@/assets/混合调度.svg';
import CustomBreadcrumb from '@/components/custom-breadcrumb';
import { NavLink, Outlet, useLocation } from 'react-router-dom';

export default function ChooseService() {
  const location = useLocation();

  if (location.pathname.includes('/local-models') || location.pathname.includes('/remote-models') || location.pathname.includes('/hybrid-scheduling')) {
    return <Outlet />;
  }
  return (
    <Flex
      gap={24}
      vertical
    >
      <CustomBreadcrumb />
      <Flex gap={20}>
        <NavLink
          to="local-models"
          className={`${styles.serviceItem} ${styles.serviceModule}`}
        >
          <img
            src={localModels}
            alt="Local Models"
          />
          <span>Local Models</span>
        </NavLink>
        <NavLink
          to="remote-models"
          className={`${styles.serviceItem} ${styles.serviceModule}`}
        >
          <img
            src={remoteModels}
            alt="Remote Models"
          />
          <span>Remote Models</span>
        </NavLink>
        <NavLink
          to="hybrid-scheduling"
          className={`${styles.serviceItem} ${styles.serviceModule}`}
        >
          <img
            src={hybridScheduling}
            alt="Hybrid Scheduling"
          />
          <span>Hybrid Scheduling</span>
        </NavLink>
      </Flex>
    </Flex>
  );
}
