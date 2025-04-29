import { Modal } from 'antd';
import ModelCard from '../model-list-content/model-card';

interface IModelDetailModalProps {
  onDetailModalVisible: (visible: boolean) => void;
}

export default function ModelDetailModal(props: IModelDetailModalProps) {
  const { onDetailModalVisible } = props;
  return (
    <Modal centered title="模型详情" open maskClosable footer={null} onCancel={() => onDetailModalVisible(false)}>
      <ModelCard isDetail={true} />
    </Modal>
  );
}
