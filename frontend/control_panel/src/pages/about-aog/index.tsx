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
import { Col, Row, Flex, Button, Radio } from 'antd';
import refreshIcon from '@/components/icons/refresh.svg';
import ProductInfo from './components/product-info';
import type { Status } from './components/product-info';
import bg from '@/assets/list-title-bg.svg';
import InfoIcon from '@/components/icons/info.svg?react';
import { PlusOutlined } from '@ant-design/icons';
import { useState } from 'react';
import chatSvg from '@/assets/chat.svg';
import textToImageSvg from '@/assets/text-to-image.svg';
import { Outlet } from 'react-router-dom';
import useServiceStore from '@/store/useServiceStore';
import Installing from '@/components/installing';

export default function AboutAOG() {
  const { installing } = useServiceStore();

  const [checkCompleted, setCheckCompleted] = useState(true);
  const [checkStatus, setCheckStatus] = useState<Status>('default');
  return (
    <>
      <div
        className={styles.aboutWrapper}
        style={{
          display: installing ? 'none' : 'flex',
        }}
      >
        <ProductInfo />
        <div className={styles.listTitle}>
          <img
            src={bg}
            alt=""
          />
          <span>AOG Service List</span>
        </div>
        <Outlet />
      </div>
      {installing && <Installing />}
    </>
  );
}
