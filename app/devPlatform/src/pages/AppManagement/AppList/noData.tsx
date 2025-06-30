import noListImage from '@/assets/noList.png';
import styles from './index.module.scss';
import { Button } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import React from 'react';

interface NoDataProps {
  showCreateModal: () => void;
}

export const NoData: React.FC<NoDataProps> = ({ showCreateModal }) => {
  return (
    <div className={styles.noDataWarp}>
      <img
        src={noListImage}
        alt="No Data"
        style={{ width: '120px', height: '120px' }}
      />
      <h2>您还没创建应用，立即创建吧</h2>

      <Button
        className={styles['create-btn']}
        type="primary"
        icon={<PlusOutlined />}
        onClick={showCreateModal}
      >
        创建应用
      </Button>
    </div>
  );
};
