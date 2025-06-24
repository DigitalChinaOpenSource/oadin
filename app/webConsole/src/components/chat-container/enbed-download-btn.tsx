import React, { useCallback, useEffect } from 'react';
import { Button, Tooltip } from 'antd';
import uploadSvg from '@/components/icons/upload.svg';
import { useDownLoad } from '@/hooks/useDownload';
import { httpRequest } from '@/utils/httpRequest';
import { useRequest } from 'ahooks';
import useModelDownloadStore from '@/store/useModelDownloadStore';

const EmbedDownloadButton = React.memo(() => {
  const { fetchDownloadStart } = useDownLoad();
  const isDownloadEmbed = useModelDownloadStore((state) => state.isDownloadEmbed);

  useEffect(() => {
    if (isDownloadEmbed) {
      fetchChooseModelNotify({
        service_name: 'embed',
        local_provider: 'local_ollama_embed',
      });
    }
  }, [isDownloadEmbed]);

  const handleDownload = useCallback(() => {
    fetchDownloadStart({
      name: 'quentinz/bge-large-zh-v1.5:f16',
      service_name: 'embed',
      source: 'local',
      service_provider_name: 'local_ollama_embed',
      id: 'bc8ca0995fcd651',
    } as any);
  }, []);

  // 选择模型后需要将所选择的通知奥丁
  const { run: fetchChooseModelNotify } = useRequest(async (params: { service_name: string; local_provider?: string; remote_provider?: string }) => {
    if (!params?.service_name) return;
    const data = await httpRequest.put('/service', {
      ...params,
    });
    return data || {};
  });

  return (
    <Tooltip
      arrow={false}
      title={
        <>
          该功能需先下载词嵌入模型
          <a onClick={handleDownload}>【立即下载】</a>
        </>
      }
    >
      <Button
        icon={
          <img
            src={uploadSvg}
            alt="上传"
          />
        }
      />
    </Tooltip>
  );
});

export default EmbedDownloadButton;
