import { useState } from 'react';
import { Modal, Input } from 'antd';
import styles from './index.module.scss';

interface IModelPathModalProps {
  modalPath?: string;
  onModalPathClose?: () => void;
}

export default function ModelPathModal(props: IModelPathModalProps) {
  const { modalPath, onModalPathClose } = props;
  // 模型存储路径
  const [saveModelPath, setSaveModalPath] = useState<string>(modalPath || '');

  const onModelPathChange = (val: string) => {
    setSaveModalPath(val);
  };

  const handleToSavePath = () => {
    // 如果没有修改, 则直接关闭
    if (modalPath === saveModelPath) {
      onModalPathClose?.();
      return;
    }
    // TODO 调接口保存
    onModalPathClose?.();
  };

  return (
    <Modal centered title="修改模型存储路径" width={480} open onOk={handleToSavePath} onCancel={onModalPathClose} className={styles.modelPathModal} okText="确认">
      <div className={styles.modelPathModal}>
        <div className={styles.tips}>
          <p>若本地模型正在工作中，该操作可能会造成业务的中断。</p>
          <p>
            在模型完成新路径迁移前,基于本地模型的所有功能将不可用。
            <span className={styles.mark}>为保障业务使用，请先将具体应用的模型切换至云端模型。</span>
          </p>
        </div>
        <div className={styles.modelPathInput}>
          <div className={styles.pathTitle}>
            <span className={styles.required}>*</span>
            <span className={styles.label}>模型存储路径：</span>
          </div>

          <Input className={styles.pathInput} allowClear value={saveModelPath} onChange={(e) => onModelPathChange(e.target.value)} />
          <div className={styles.diskSpace}>
            <span>
              （<span>100GB</span>/<span>948GB</span>，<span className={styles.diskCanUse}>666GB可用</span>）
            </span>
          </div>
        </div>
      </div>
    </Modal>
  );
}
