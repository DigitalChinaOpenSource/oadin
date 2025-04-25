import styles from './index.module.scss';
import { Modal } from 'antd';
import { useViewModel } from './view-model';
     
interface ServiceProviderDetailProps {
  visible: boolean;
  onCancel: () => void;
  onSubmit: () => void;
}

export default function ServiceProviderDetail({ 
  visible, 
  onCancel, 
  onSubmit 
}: ServiceProviderDetailProps) {
  const vm = useViewModel();
  const { baseInfo, modelList } = vm;

  return (
    <Modal
      open={visible}
      width={860}
      title={<div className={styles.modalTitle}>查看服务提供商详情</div>}
      onCancel={onCancel}
      onOk={() => onSubmit}
    >
      <div className={styles.infoName}>基础信息</div>
      <div className={styles.infoBlock}>
        <div className={styles.infoItem}>
          <span className={styles.infoLabel}>服务提供商名称:</span>
          {baseInfo.name}
        </div>
        <div className={styles.infoItem}>
          <span className={styles.infoLabel}>服务提供商厂商名称:</span>
          {baseInfo.name}
        </div>
        <div className={styles.infoItem}>
          <span className={styles.infoLabel}>服务来源:</span>
          {baseInfo.name}
        </div>
        <div className={styles.infoItem}>
          <span className={styles.infoLabel}>服务名称:</span>
          {baseInfo.name}
        </div>
        <div className={styles.infoItem}>
          <span className={styles.infoLabel}>服务提供商状态:</span>
          {baseInfo.name}
        </div>
        <div className={styles.infoItem}>
          <span className={styles.infoLabel}>模型上下问长度:</span>
          {baseInfo.name}
        </div>
        <div className={styles.infoRow}>
          <span className={styles.infoLabel}>描述:</span>
          {baseInfo.name}
        </div>
      </div>
      <div className={styles.infoName}>鉴权信息</div>
      <div className={styles.infoBlock}>
        <div className={styles.infoRow}>
          <span className={styles.infoLabel}>请求方法:</span>
          {baseInfo.name}
        </div>
        <div className={styles.infoRow}>
          <span className={styles.infoLabel}>请求URL:</span>
          {baseInfo.name}
        </div><div className={styles.infoRow}>
          <span className={styles.infoLabel}>鉴权类型:</span>
          {baseInfo.name}
        </div><div className={styles.infoRow}>
          <span className={styles.infoLabel}>鉴权信息:</span>
          {baseInfo.name}
        </div>
      </div>

      <div className={styles.modelList}>
        <div className={styles.modelTitle}>支持的模型列表</div>
        {modelList.map((model) => <div className={styles.modelItem} >
          {/* <Image  /> */}
        模型名称: {model.name}
        </div>)}
      </div>
    </Modal>
  );
};