import { Modal } from 'antd';
import ModelCard from '../model-list-content/model-card';
import { ModelDataItem } from '@/types';
interface IModelDetailModalProps {
  onDetailModalVisible: (visible: boolean, selectModelData?: ModelDataItem) => void;
  // 模型数据
  selectModelData: ModelDataItem;
}

export default function ModelDetailModal(props: IModelDetailModalProps) {
  const { onDetailModalVisible, selectModelData } = props;
  return (
    <Modal
      centered
      title="模型详情"
      open
      maskClosable
      footer={null}
      onCancel={() => onDetailModalVisible(false)}
    >
      <ModelCard
        isDetail={true}
        modelData={selectModelData}
      />
    </Modal>
  );
}
