import React, { useCallback } from 'react';
import { Button, Tooltip } from 'antd';
import uploadSvg from '@/components/icons/upload.svg';
// import { useDownLoad } from '@/hooks/useDownload';

const EmbedDownloadButton = React.memo(() => {
  // const { fetchDownloadStart } = useDownLoad();

  const handleDownload = useCallback(() => {
    // fetchDownloadStart({
    //   name: 'quentinz/bge-large-zh-v1.5:f16',
    //   service_name: 'embed',
    //   source: 'local',
    //   service_provider_name: 'local_ollama_embed',
    //   id: 'bc8ca0995fcd651',
    // } as any);
  }, []);

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
