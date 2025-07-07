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
import { Flex, Button, notification, FloatButton, Skeleton } from 'antd';
import ModelCard from '@/components/model-card';
import InfoIcon from '@/components/icons/info.svg?react';
import { useState, useEffect } from 'react';

import DownloadModal from '@/components/dowload-modal';
import DownloadIcon from '@/components/icons/download.svg?react';
import { httpRequest } from '@/utils/httpRequest.ts';
import { useRequest } from 'ahooks';
import { useParams } from 'react-router-dom';
import { useDownLoad } from '@/hooks/useDownload';
import useModelDownloadStore from '@/store/useModelDownloadStore';
const LocalModelDetail = () => {
  const params = useParams();
  const [api, contextHolder] = notification.useNotification();
  const { downloadList, setDownloadList } = useModelDownloadStore();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isDownloaded, setIsDownloaded] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const { fetchDownloadStart } = useDownLoad();
  const service_name = params.type as string;
  const { loading, data, run } = useRequest(
    () =>
      httpRequest.get('/control_panel/modellist', {
        service_name,
        service_source: 'local',
        page_size: 1,
        page: 1,
        search_name: params.name,
      }),
    {
      manual: true,
    },
  );
  const { run: setDefaultModel, loading: setting } = useRequest((data) => httpRequest.post('/control_panel/set_default', data), {
    manual: true,
    onSuccess: (res) => {
      run();
    },
  });
  const { run: delteteModel, loading: deleting } = useRequest((data) => httpRequest.del('/model', data), {
    manual: true,
    onSuccess: (res) => {
      api.destroy();
      run();
    },
  });
  const model = data?.data[0];
  useEffect(() => {
    run();
  }, [run]);
  const close = () => {
    console.log('Notification was closed. Either the close button was clicked or duration time elapsed.');
  };
  const showModal = () => {
    setIsModalOpen(true);
  };

  const handleOk = () => {
    setIsModalOpen(false);
  };

  const handleCancel = () => {
    setIsModalOpen(false);
  };

  const openNotification = () => {
    const key = `open${Date.now()}`;
    const actions = (
      <>
        <Button
          style={{ width: '50%', height: '56px' }}
          size="large"
          onClick={() => api.destroy()}
        >
          Cancel
        </Button>
        <Button
          type="primary"
          style={{ width: '50%', height: '56px' }}
          size="large"
          onClick={() => {
            delteteModel({
              model_name: model.name,
              service_name: model.service_name,
              service_source: model.source,
              provider_name: model.service_provider_name,
            });
            api.destroy();
          }}
        >
          Delete
        </Button>
      </>
    );
    api.open({
      icon: <InfoIcon />,
      message: 'Delete Tip',
      description: 'The deletion operation will delete all your conversation records, file information with the model, and terminate the related processes.',
      actions,
      placement: 'bottomRight',
      duration: null,
      key,
      onClose: close,
    });
  };
  useEffect(() => {
    if (!model) return;
    const download = downloadList.find((item) => item.id === model.id);
    // console.log('当前下载状态:', JSON.stringify(downloadList));

    // 下载完成时刷新
    if (download && download.status === 2) {
      run();
      // 你可以在这里把该下载项从 downloadList 移除，避免重复 run
      // setDownloadList(list => list.filter(v => v.id !== model.id));
    }

    // 同步 downloading 状态
    setDownloading(!!(download && download.status === 1));
  }, [downloadList, model, run]);
  return (
    <Flex
      vertical
      gap={24}
    >
      {contextHolder}
      <CustomBreadcrumb />
      <Skeleton
        loading={loading || setting}
        active
        avatar
      >
        {model && (
          <Flex
            vertical
            gap={24}
          >
            <ModelCard
              model={model}
              showSet
              setDefault={() => {
                setDefaultModel({
                  model_name: model.name,
                  service_name: model.service_name,
                  service_source: model.source,
                  provider_name: model.service_provider_name,
                });
              }}
            />
            <Flex justify="end">
              {model.can_select ? (
                <Button
                  loading={deleting}
                  ghost
                  type="primary"
                  onClick={openNotification}
                >
                  Delete
                </Button>
              ) : (
                <Flex
                  gap={8}
                  align="center"
                >
                  <span>size:{model.size}</span>
                  <Button
                    type="primary"
                    size="large"
                    onClick={() => {
                      if (!downloading) {
                        fetchDownloadStart({
                          name: model.name,
                          service_name: model.service_name,
                          source: 'local',
                          service_provider_name: model.service_provider_name,
                          id: model.id,
                          avatar: model.avatar,
                          size: model.size,
                        } as any);
                        setDownloading(true);
                      }
                      showModal();
                    }}
                  >
                    {downloading ? 'Downloading' : 'Download'}
                  </Button>
                </Flex>
              )}
            </Flex>
            {downloadList?.length > 0 && (
              <FloatButton
                onClick={showModal}
                icon={<DownloadIcon />}
                style={{ width: '32px', height: '32px', bottom: '24px' }}
              />
            )}
          </Flex>
        )}
      </Skeleton>
      <DownloadModal
        isModalOpen={isModalOpen}
        setIsModalOpen={setIsModalOpen}
      />
    </Flex>
  );
};
export default LocalModelDetail;
