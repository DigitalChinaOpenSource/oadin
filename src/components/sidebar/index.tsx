import { useState, useEffect } from 'react';
import { Menu, Button } from 'antd';
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
      <Menu
        mode="inline"
        selectedKeys={[selectedKey]}
        className={styles.menu}
        items={menuItems}
        onClick={({ key }) => handleMenuClick(key)}
      />
      <div className={styles.downloadBtn}>
        <Button
          type="primary"
          icon={<DownloadOutlined />}
          onClick={handleDownload}
        >
          下载模型
        </Button>
      </div>
    </div>
  );
};

export default Sidebar;