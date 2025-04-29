import { Modal, Button } from 'antd';
import ModelCard from '../model-card';

interface IModelDetailModalProps {
  onDetailClose: () => void;
}

export default function ModelDetailModal(props: IModelDetailModalProps) {
  const { onDetailClose } = props;
  return (
    <Modal title="模型详情" open maskClosable footer={null} onCancel={onDetailClose}>
      <ModelCard isDetail={true} />
    </Modal>
  );
}
