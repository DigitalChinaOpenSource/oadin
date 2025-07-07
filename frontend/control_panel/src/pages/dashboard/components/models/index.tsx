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

import { Col, Row, Flex, Button, Modal, Select, Radio } from 'antd';
import { PlusOutlined, RightOutlined } from '@ant-design/icons';
import { NavLink } from 'react-router-dom';
import styles from './index.module.scss';
import RightIcon from '@/components/icons/right.svg?react';
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
type ModelsProps = {
  models: Model[];
  serviceName: string;
  modelType: string;
};
export const Models: React.FC<ModelsProps> = (props: ModelsProps) => {
  return props.models.length === 0 ? (
    <NavLink to={`/about-aog/choose-service/${props.serviceName}/${props.modelType}`}>
      <Button
        type="primary"
        size="large"
        className={styles.addModelBtn}
        icon={<PlusOutlined />}
      >
        <span className={styles.addModelText}>Add Model</span>
      </Button>
    </NavLink>
  ) : (
    <Flex
      gap={12}
      vertical
    >
      <Flex
        gap={12}
        vertical
        align="start"
      >
        {props.models.slice(0, 2).map((model: Model, modelIndex: number) => (
          <div
            key={modelIndex}
            className={styles.modelItem}
          >
            <div className={styles.modelImage}>
              <img
                src={model.Avatar}
                alt=""
              />
            </div>
            <NavLink
              title={model.model_name}
              to={`/about-aog/choose-service/${props.serviceName}/${props.modelType}/${encodeURIComponent(model.model_name)}`}
              className={styles.modelName}
            >
              {model.model_name}
            </NavLink>
            {model.is_default && <span className="defaultBtn">Default</span>}
          </div>
        ))}
      </Flex>
      {props.models.length > 2 && (
        <NavLink
          className={styles.showMore}
          to={`/about-aog/choose-service/${props.serviceName}/${props.modelType}`}
        >
          Show more
          <RightIcon />
        </NavLink>
      )}

      <NavLink to={`/about-aog/choose-service/${props.serviceName}/${props.modelType}`}>
        <Button
          type="primary"
          size="large"
          className={styles.addModelBtn}
          icon={<PlusOutlined />}
        >
          <span className={styles.addModelText}>Add Model</span>
        </Button>
      </NavLink>
    </Flex>
  );
};
