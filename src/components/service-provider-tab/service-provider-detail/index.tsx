import styles from './index.module.scss';
import { Modal, Pagination } from 'antd';
import { useViewModel } from './view-model';
import modelPng from '@/assets/modelLogo.png';
import { IServiceProviderDataItem } from '../types';

interface ServiceProviderDetailProps {
  selectedRow: IServiceProviderDataItem;
  onCancel: () => void;
}

export default function ServiceProviderDetail({ selectedRow, onCancel }: ServiceProviderDetailProps) {
  const vm = useViewModel();
  const { baseInfo, modelList, pagination, handlePagChange } = vm;

  return (
    <Modal
      centered
      open
      width={860}
      footer={null}
      title={<div className={styles.modalTitle}>查看服务提供商详情</div>}
      onCancel={onCancel}
      okText="确认"
    >
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
          <div className={baseInfo.statusCode === '1' ? styles.readyStatus : styles.disabledStatus}>
            <div className={styles.dot}></div>
            {baseInfo.statusCode === '1' ? '可用' : '禁用'}
          </div>
        </div>
        <div className={styles.infoRow}>
          <span className={styles.infoLabel}>鉴权信息:</span>
          {baseInfo.name}
        </div>
      </div>

      <div className={styles.pageBlock}>
        <div className={styles.modelTitle}>支持的模型列表</div>
        <Pagination
          current={pagination.current}
          showSizeChanger={false}
          total={pagination.total}
          onChange={handlePagChange}
          show-less-items
        />
      </div>
      <div className={styles.modelList}>
        {modelList.map((model, index) => (
          <div
            className={styles.modelItem}
            key={index}
          >
            <div className={styles.modelLeft}>
              <img
                src={modelPng}
                alt="modelLogo"
              />
              <span className={styles.modelBaseInfo}>
                <span className={styles.modelLabel}>模型名称: </span>
                {model.name}
              </span>
              {model.tags.map((tag, tagIndex) => (
                <span
                  key={tagIndex}
                  className={styles.tagItem}
                >
                  {tag}
                </span>
              ))}
            </div>
            <div className={styles.modelRight}>
              {model.text}
              <div className={styles.line}></div>
              上下文长度: {model.size}k<div className={styles.line}></div>
              {model.status === 1 ? '已下载' : '未下载'}
            </div>
          </div>
        ))}
      </div>
      <div className={styles.timeInfo}>
        <span>创建时间: {baseInfo.createTime}</span>
        <span>更新时间: {baseInfo.updateTime}</span>
      </div>
    </Modal>
  );
}
