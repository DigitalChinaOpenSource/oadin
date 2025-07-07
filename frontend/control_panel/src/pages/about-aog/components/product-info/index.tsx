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
import { Flex, message, Spin, Skeleton } from 'antd';
import RefreshIcon from '@/components/icons/refresh.svg?react';
import CheckingIcon from '@/components/icons/loader.svg?react';
import SuccessIcon from '@/components/icons/checkmark--filled.svg?react';
import { useState } from 'react';
import { httpRequest } from '@/utils/httpRequest.ts';
import { useRequest } from 'ahooks';
import useServerCheckStore from '@/store/useServerCheckStore';

export type Status = 'default' | 'checking' | 'success';
const statusIconObj: Record<Status, React.FC<React.SVGProps<SVGSVGElement>>> = {
  default: RefreshIcon,
  checking: CheckingIcon,
  success: SuccessIcon,
};
export default function ProductInfo() {
  const [checkServerStatus, setCheckServerStatus] = useState<Status>('default');
  const { checkStatus, fetchServerStatus } = useServerCheckStore();
  const { loading, data } = useRequest(() => httpRequest.get('/control_panel/about'));
  const StatusIcon = statusIconObj[checkServerStatus];
  const startCheck = async () => {
    if (checkServerStatus === 'default') {
      setCheckServerStatus('checking');
      try {
        const res = await fetchServerStatus();
        if (res) {
          setCheckServerStatus('success');
          message.success('Check succeeded');
        }

        setTimeout(() => {
          setCheckServerStatus('default');
        }, 3000);
      } catch (error) {
        setCheckServerStatus('default');
      }
    }
  };

  return (
    <div className={styles.productWrapper}>
      <Skeleton
        loading={loading}
        active
        avatar
      >
        <Flex
          gap={24}
          flex={1}
        >
          <div className={styles.productImage}>icon</div>
          <Flex
            vertical
            gap={8}
            flex={1}
          >
            <Flex
              justify="space-between"
              align="center"
            >
              <Flex
                gap={8}
                align="center"
              >
                <span className={styles.productName}>{data?.productname}</span>
                <span className={styles.productVersion}>{data?.version}</span>
              </Flex>
              <Flex
                align="center"
                gap={4}
                className={styles.checkStatus}
                onClick={startCheck}
              >
                <StatusIcon className={checkServerStatus === 'checking' ? styles.checking : ''} />
                <span>Check the status</span>
              </Flex>
            </Flex>

            <Flex gap={12}>
              <span>Service Status:</span>
              <Flex align="center">
                <span
                  className={styles.statusDot}
                  style={{ backgroundColor: checkStatus ? '#00CC39' : '#F5222D' }}
                ></span>
                <span style={{ color: checkStatus ? '#00CC39' : '#F5222D' }}>{checkStatus ? 'Available' : 'Unavailable'}</span>
              </Flex>
            </Flex>
            <div>
              <span>Product Description:</span>
              <span>{data?.description}</span>
            </div>
          </Flex>
        </Flex>
      </Skeleton>
    </div>
  );
}
