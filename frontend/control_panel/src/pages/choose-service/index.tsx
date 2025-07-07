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
import { Col, Row, Flex, Button, Spin } from 'antd';
import { upFirstLetter } from '@/utils';

import InfoIcon from '@/components/icons/info.svg?react';
import { PlusOutlined } from '@ant-design/icons';
import { useState } from 'react';
import chatSvg from '@/assets/chat.svg';
import textToImageSvg from '@/assets/text-to-image.svg';
import { useNavigate, NavLink, Outlet, useLocation, matchPath } from 'react-router-dom';
import useServiceStore from '@/store/useServiceStore';
import { useServices } from '@/hooks/useServices/index.ts';
import InstallService from '@/components/install-service';

const serviceImgObj = {
  chat: chatSvg,
  'text-to-image': textToImageSvg,
};
export default function ChooseService() {
  const [checkCompleted, setCheckCompleted] = useState(false);
  const { loading, services, getServices, getServicesAsync, install, unInstalledSerivices } = useServices();
  const location = useLocation();
  const { installedServices } = useServiceStore();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const showModal = () => {
    setIsModalOpen(true);
  };
  // 判断当前是否正好是 /choose-service 路由
  if (location.pathname !== '/about-aog/choose-service') {
    return <Outlet />;
  }
  return (
    <Spin spinning={loading}>
      <Flex
        gap={24}
        vertical
        align="start"
      >
        <div className={styles.commonTitle}>Choose Service</div>
        {installedServices.length ? (
          <Flex gap={20}>
            {installedServices.map((service) => (
              <NavLink
                key={service.service_name}
                to={service.service_name}
                className={styles.serviceItem}
              >
                <img
                  src={service.avatar || serviceImgObj[service.service_name as keyof typeof serviceImgObj] || chatSvg}
                  alt="Chat Service"
                />
                <span>{upFirstLetter(service.service_name)}</span>
              </NavLink>
            ))}
          </Flex>
        ) : (
          <div className={styles.tipContainer}>
            <InfoIcon />
            <span>No service available. Please install the service first.</span>
          </div>
        )}
        <div className={styles.commonTitle}>Install Services</div>
        <Button
          type="primary"
          size="large"
          icon={<PlusOutlined />}
          onClick={showModal}
        >
          Install More Service
        </Button>
        <InstallService
          isModalOpen={isModalOpen}
          setIsModalOpen={setIsModalOpen}
          install={install}
          services={unInstalledSerivices}
        />
      </Flex>
    </Spin>
  );
}
