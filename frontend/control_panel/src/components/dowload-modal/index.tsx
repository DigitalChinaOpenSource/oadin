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

import { Flex, Modal } from 'antd';
import PauseIcon from '@/components/icons/pause.svg?react';
import ContinueIcon from '@/components/icons/continue.svg?react';
import LoadingIcon from '@/components/icons/loading.svg?react';
import CloseIcon from '@/components/icons/close.svg?react';
import styles from './index.module.scss';
import useModelDownloadStore from '@/store/useModelDownloadStore';
import { useDownLoad } from '@/hooks/useDownload';
import { useEffect } from 'react';

type DownloadModalProps = {
  isModalOpen: boolean;
  setIsModalOpen: (isOpen: boolean) => void;
};
const DownloadModal: React.FC<DownloadModalProps> = ({ isModalOpen, setIsModalOpen }) => {
  const { downloadList, setDownloadList } = useModelDownloadStore();
  const { fetchDownLoadAbort, fetchDownloadStart } = useDownLoad();
  useEffect(() => {
    if (downloadList.length === 0) {
      setIsModalOpen(false);
    }
  }, [downloadList]);
  return (
    <Modal
      title="Download Progress"
      open={isModalOpen}
      footer={null}
      mask={false}
      maskClosable={true}
      onCancel={() => {
        setIsModalOpen(false);
      }}
      width={444}
      style={{
        position: 'fixed',
        margin: 0,
        bottom: '24px',
        right: '64px',
        top: 'auto',
        paddingBottom: '0',
      }}
    >
      <Flex
        gap={16}
        vertical
        style={{ paddingBottom: '16px' }}
      >
        {downloadList.map((item) => (
          <div
            className={styles.downloadItem}
            key={item.id}
          >
            <Flex
              justify="space-between"
              align="center"
            >
              <Flex
                gap={8}
                align="center"
              >
                <img
                  className={styles.avatar}
                  src={item.avatar}
                  alt={item.name}
                />
                <span>{item.name}</span>
              </Flex>
              <Flex gap={12}>
                {item.status === 1 && (
                  <PauseIcon
                    onClick={() => {
                      fetchDownLoadAbort({ model_name: item.name }, { id: item.id });
                    }}
                    style={{ cursor: 'pointer' }}
                  />
                )}
                {(item.status === 3 || item.status === 0) && (
                  <ContinueIcon
                    onClick={() => {
                      fetchDownloadStart({
                        ...item,
                      } as any);
                    }}
                    style={{ cursor: 'pointer' }}
                  />
                )}
                <CloseIcon
                  onClick={() => {
                    fetchDownLoadAbort({ model_name: item.name }, { id: item.id }).then((res) => {
                      // console.log(res, 'Download cancelled');

                      setDownloadList((currentList) => currentList.filter((v) => v.id !== item.id));
                    });
                  }}
                  style={{ cursor: 'pointer' }}
                />
              </Flex>
            </Flex>
            <Flex
              style={{ marginTop: '4px', marginBottom: '8px' }}
              justify="space-between"
              align="center"
            >
              {/* <span>{`${item.completedsize || 0}MB/${item.totalsize ? item.totalsize + 'MB' : item.size}`}</span> */}
              <span>{`${item.currentDownload || 0}%`}</span>
              <Flex
                gap={4}
                align="center"
              >
                {item.status === 1 && (
                  <>
                    <LoadingIcon className={styles.loading} /> <span>Downloading</span>
                  </>
                )}
                {(item.status === 0 || item.status === 3) && <span>Suspended</span>}
              </Flex>
            </Flex>
            <div className={styles.progressBar}>
              <div
                className={styles.progress}
                style={{ width: `${item.currentDownload}%` }}
              />
            </div>
          </div>
        ))}
        {downloadList.length === 0 && <span>No downloads in progress</span>}
      </Flex>
    </Modal>
  );
};

export default DownloadModal;
