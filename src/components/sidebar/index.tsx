import { useState, useEffect } from 'react';
import { Button } from 'antd';
import { AppstoreOutlined, CloudServerOutlined, DownloadOutlined } from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import styles from './index.module.scss';
import favicon from '../../assets/favicon.png';

const Sidebar = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [selectedKey, setSelectedKey] = useState('model-manage');

  useEffect(() => {
    const path = location.pathname.split('/')[1] || 'model-manage';
    setSelectedKey(path);
  }, [location]);

  const menuItems = [
    {
      key: 'model-manage',
      icon: <AppstoreOutlined />,
      label: '模型管理',
    },
    {
      key: 'server-manage',
      icon: <CloudServerOutlined />,
      label: '服务管理',
    },
  ];

  const handleMenuClick = (key: string) => {
    navigate(`/${key}`);
    setSelectedKey(key);
  };

  const handleDownload = () => {
    console.log('下载模型');
    // 实现模型下载逻辑
  };

  return (
    <div className={styles.sidebar}>
      <div className={styles.logo}>
        {/* 替换logo */}
        <img src={favicon} alt="Logo" />
      </div>
      <div className={styles.menuContainer}>
        {menuItems.map(item => (
          <div 
            key={item.key}
            className={`${styles.menuItem} ${selectedKey === item.key ? styles.selected : ''}`}
            onClick={() => handleMenuClick(item.key)}
          >
            <div className={styles.icon}>{item.icon}</div>
            <div className={styles.label}>{item.label}</div>
          </div>
        ))}
      </div>
      <div className={styles.downloadBtn}>
        <Button
          type="primary"
          icon={<DownloadOutlined />}
          onClick={handleDownload}
        />
      </div>
    </div>
  );
};

export default Sidebar;