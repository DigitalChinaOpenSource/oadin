import React from 'react';
import { Modal, Button } from 'antd';
import ModelCard from '../model-card';

interface IModelDetailModalProps {
  onDetailClose: () => void;
}

export default function ModelDetailModal(props: IModelDetailModalProps) {
  const { onDetailClose } = props;
  return (
    <Modal
      title="模型详情"
      open={true}
      // width={800}
      onCancel={onDetailClose}
      footer={[
        <Button key="back" onClick={onDetailClose}>
          关闭
        </Button>,
      ]}
    >
      {/* <div style={{overflow: 'hidden'}}> */}
        <ModelCard isDetail={true} />
      {/* </div> */}
    </Modal>
  )
}