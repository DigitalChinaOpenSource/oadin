import React from 'react';
import { Modal } from 'antd';
import styles from '@/pages/AppManagement/AppList/index.module.scss';

export interface DeleteAppModalProps {
  setDeletingApp: (deleteApp?: any) => void;
  onDelete: () => Promise<void>;
  deleteApp?: any;
  deleteLoading?: boolean;
}

const DeleteAppModal: React.FC<DeleteAppModalProps> = ({ deleteApp, setDeletingApp, onDelete, deleteLoading }) => {
  return (
    <Modal
      title="删除应用"
      open={!!deleteApp}
      onCancel={() => {
        setDeletingApp(null);
      }}
      confirmLoading={deleteLoading}
      onOk={onDelete}
      width={480}
      className={styles['delete-modal']}
    >
      <div className={styles['delete-modal-content']}>
        <p>
          <strong>删除操作将同时移除以下所有关联数据：</strong>
        </p>
        <ul>
          <li>应用配置文件（含模型参数与 MCP 配置）</li>
          <li>APPID 与 Secret Key 凭证信息</li>
          <li>日志、统计与调用记录等应用数据</li>
        </ul>
        <p>一旦删除，基于该应用构建的服务将停止运行，且数据将无法恢复。</p>
        <p>
          <strong>此操作为不可逆，请谨慎操作。</strong>
        </p>
      </div>
    </Modal>
  );
};

export default DeleteAppModal;
