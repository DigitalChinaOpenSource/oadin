import { Modal } from 'antd';
import GeneralCard from '../model-list-content/general-card';
import { IModelDataItem, IModelSourceType } from '@/types';
interface IModelDetailModalProps {
  onDetailModalVisible: (visible: boolean, selectModelData?: IModelDataItem) => void;
  // 模型数据
  selectModelData: IModelDataItem;
  modelSourceVal: IModelSourceType;
}

export default function ModelDetailModal(props: IModelDetailModalProps) {
  const { onDetailModalVisible, selectModelData, modelSourceVal } = props;
  return (
    <Modal
      centered
      title="模型详情"
      open
      maskClosable
      footer={null}
      onCancel={() => onDetailModalVisible(false)}
    >
      <GeneralCard
        isDetail={true}
        modelSourceVal={modelSourceVal}
        modelData={selectModelData}
      />
    </Modal>
  );
}
