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

import { Col, Row, Flex, Button, Radio, Spin } from 'antd';
import styles from './index.module.scss';
import { PlusOutlined } from '@ant-design/icons';
import { useState } from 'react';
import Installing from '@/components/installing';
import chatSvg from '@/assets/chat.svg';
import textToImageSvg from '@/assets/text-to-image.svg';
import { Models } from './components/models';
import InstallService from '@/components/install-service';
import type { Service } from '@/hooks/useServices/index.ts';
import { useServices } from '@/hooks/useServices/index.ts';
import { upFirstLetter } from '@/utils';
import useServiceStore from '@/store/useServiceStore';
const serviceImgObj = {
  chat: chatSvg,
  'text-to-image': textToImageSvg,
};
export default function Dashboard() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const { loading, mutate, services, models, updateService, install, installing, unInstalledSerivices } = useServices();
  const { installedServices } = useServiceStore();
  const showModal = () => {
    setIsModalOpen(true);
  };

  return installing ? (
    <Installing />
  ) : (
    <Spin spinning={loading}>
      <Row gutter={4}>
        <Col span={6}>
          <Flex
            vertical
            gap={4}
          >
            <div className={styles.categoryTitle}>AOG Services</div>
            {installedServices.map((service: Service, index: number) => (
              <div
                className={styles.card}
                key={index}
              >
                <Flex
                  vertical
                  align="center"
                  gap={8}
                >
                  <img
                    src={service.avatar || serviceImgObj[service.service_name as keyof typeof serviceImgObj] || chatSvg}
                    alt={service.service_name}
                    className={styles.serviceIcon}
                  />
                  <div className={styles.serviceName}>{upFirstLetter(service.service_name)}</div>
                </Flex>
              </div>
            ))}
            <div className={styles.card}>
              <Button
                type="primary"
                size="large"
                onClick={showModal}
                icon={<PlusOutlined />}
              >
                Install More Service
              </Button>
            </div>
          </Flex>
        </Col>
        <Col span={6}>
          <Flex
            vertical
            gap={4}
          >
            <div className={styles.categoryTitle}>Hybrid Scheduling</div>
            {installedServices.map((service: Service, index: number) => (
              <div
                className={styles.card}
                key={index}
              >
                <Flex
                  vertical
                  gap={8}
                >
                  <Radio.Group
                    style={{ display: 'flex', flexDirection: 'column', gap: 12 }}
                    value={service.hybrid_policy}
                    onChange={(e) => {
                      mutate((oldData: any) => {
                        const newData = { ...oldData };
                        const serviceIndex = newData.services.findIndex((item: Service) => item.service_name === service.service_name);
                        if (serviceIndex !== -1) {
                          newData.services[serviceIndex].hybrid_policy = e.target.value;
                          updateService({ service_name: service.service_name, hybrid_policy: e.target.value });
                        }
                        return newData;
                      });
                    }}
                  >
                    <Radio value="default">auto switch</Radio>
                    <Radio value="always_local">always local</Radio>
                    <Radio value="always_remote">always remote</Radio>
                  </Radio.Group>
                </Flex>
              </div>
            ))}
            <div className={styles.card}></div>
          </Flex>
        </Col>
        <Col span={6}>
          <Flex
            vertical
            gap={4}
          >
            <div className={styles.categoryTitle}>Local Models</div>
            {installedServices.map((service: Service, index: number) => (
              <div
                className={styles.card}
                key={index}
              >
                <Models
                  models={models.filter((model) => model.service_name === service.service_name && model.service_source === 'local')}
                  serviceName={service.service_name}
                  modelType="local-models"
                />
              </div>
            ))}
            <div className={styles.card}></div>
          </Flex>
        </Col>
        <Col span={6}>
          <Flex
            vertical
            gap={4}
          >
            <div className={styles.categoryTitle}>Remote Models</div>
            {installedServices.map((service: Service, index: number) => (
              <div
                className={styles.card}
                key={index}
              >
                <Models
                  models={models.filter((model) => model.service_name === service.service_name && model.service_source === 'remote')}
                  serviceName={service.service_name}
                  modelType="remote-models"
                />
              </div>
            ))}
            <div className={styles.card}></div>
          </Flex>
        </Col>
      </Row>
      <InstallService
        isModalOpen={isModalOpen}
        setIsModalOpen={setIsModalOpen}
        install={install}
        services={unInstalledSerivices}
      />
    </Spin>
  );
}
