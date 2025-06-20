import React, { useRef } from 'react';
import { Button, Tooltip } from 'antd';
import { CopyOutlined, DeleteOutlined, DownloadOutlined, EditOutlined, RobotOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
// import { useIsOverflowed } from '@/hooks/useIsOverflowed';
import styles from './index.module.scss';

interface AppCardProps {
  app: any;
  onEdit: (app: any) => void;
  onDelete: (app: any) => void;
  onCopy: (text: string) => void;
}

const AppCard: React.FC<AppCardProps> = ({ app, onEdit, onDelete, onCopy }) => {
  const nameRef = useRef<HTMLDivElement>(null);
  const isOverflowed = true;
  return (
    <div className={styles['app-card-wrapper']} key={app.id}>
      <div className={styles['card-main-row']}>
        <div className={styles['card-content']}>
          <div className={styles['app-icon-container']}>
            <RobotOutlined className={styles.icon} />
          </div>
          
          <div className={styles['app-content-container']}>
            {/* 应用名称 */}
            <div className={styles['app-header']}>
              <Tooltip
                title={app.name}
                placement="topLeft"
                open={isOverflowed ? undefined : false}
              >
                <div
                  ref={nameRef}
                  className={styles['app-title']}
                  title={app.name}
                >
                  {app.name}
                </div>
              </Tooltip>
            </div>
            
            {/* APP ID 和 Secret Key */}
            <div className={styles['app-info-row']}>
              <div className={styles['app-info-item']}>
                <span>APP ID：</span>
                <span>{app.appId}</span>
                <CopyOutlined
                  className={styles['cursor-pointer']}
                  onClick={() => onCopy(app.appId)}
                />
              </div>
              <div className={styles['app-info-item']}>
                <span>Secret Key：</span>
                <span>******************</span>
                <CopyOutlined
                  className={styles['cursor-pointer']}
                  onClick={() => onCopy(app.secretKey)}
                />
              </div>
            </div>
            
            {/* 元数据 */}
            <div className={styles['app-meta-container']}>
              <div className={styles['app-meta-row']}>
                <div className={styles['meta-item']}>
                  <span>模型：</span>
                  <span>{app.modelCount}</span>
                </div>
                <div className={styles['meta-divider']} />
                <div className={styles['meta-item']}>
                  <span>MCP服务：</span>
                  <span>{app.mcpCount}</span>
                </div>
                <div className={styles['meta-divider']} />
                <div className={styles['meta-item']}>
                  <span>操作系统：</span>
                  <span>{app.osCount}</span>
                </div>
                <div className={styles['meta-divider']} />
                <div className={styles['meta-item']}>
                  <span>更新时间：</span>
                  <span>{dayjs(app.updatedAt).format('YYYY.MM.DD HH:mm')}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        {/* 操作按钮 */}
        <div className={styles['card-actions']}>
          <Button 
            type="text" 
            className={styles['action-button']}
            onClick={() => onEdit(app)}
            icon={<EditOutlined className={styles['button-icon']} />}
          >
            <span className={styles['button-text']}>配置应用</span>
          </Button>
          <Button 
            type="text" 
            className={styles['action-button']}
            icon={<DownloadOutlined className={styles['button-icon']} />}
          >
            <span className={styles['button-text']}>下载配置文件</span>
          </Button>
          <Button 
            type="text" 
            className={styles['action-button']}
            onClick={() => onDelete(app)}
            icon={<DeleteOutlined className={styles['button-icon']} />}
          >
            <span className={styles['button-text']}>删除</span>
          </Button>
        </div>
      </div>
    </div>
  );
};

export default AppCard;
