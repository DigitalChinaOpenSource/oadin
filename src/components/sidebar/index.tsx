import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import styles from './index.module.scss';
import favicon from '../../assets/favicon.png';
import { SiderDownloadIcon } from '../icons';
import DownloadListBox from '../download-list-box';
import mm from '../icons/mm.svg';
import mmac from '../icons/mmac.svg';
import sm from '../icons/sm.svg';
import smac from '../icons/smac.svg';

const Sidebar = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [selectedKey, setSelectedKey] = useState('model-manage');
  // 是否展示下载列表弹窗
  const [isDownloadListOpen, setIsDownloadListOpen] = useState(false);

  useEffect(() => {
    const path = location.pathname.split('/')[1] || 'model-manage';
    setSelectedKey(path);
  }, [location]);

  const menuItems = [
    {
      key: 'model-manage',
      activeIcon: (
        <img
          src={mmac}
          alt="模型管理"
        />
      ),
      inactiveIcon: (
        <img
          src={mm}
          alt="模型管理"
        />
      ),
      label: '模型管理',
    },
    {
      key: 'server-manage',
      activeIcon: (
        <img
          src={smac}
          alt="服务管理"
        />
      ),
      inactiveIcon: (
        <img
          src={sm}
          alt="服务管理"
        />
      ),
      label: '服务管理',
    },
  ];

  const handleMenuClick = (key: string) => {
    navigate(`/${key}`);
    setSelectedKey(key);
  };

  const handleDownload = () => {
    setIsDownloadListOpen(!isDownloadListOpen);
  };

  return (
    <div className={styles.sidebar}>
      <div className={styles.logo}>
        <img
          src={favicon}
          alt="Logo"
        />
      </div>
      <div className={styles.menuContainer}>
        {menuItems.map((item) => (
          <div
            key={item.key}
            className={`${styles.menuItem} ${selectedKey === item.key ? styles.selected : ''}`}
            onClick={() => handleMenuClick(item.key)}
          >
            <div className={styles.icon}>{selectedKey === item.key ? item.activeIcon : item.inactiveIcon}</div>
            <div
              className={styles.label}
              style={{
                color: selectedKey === item.key ? '#000115' : '#344054',
              }}
            >
              {item.label}
            </div>
          </div>
        ))}
      </div>
      <div className={styles.downloadBtnBox}>
        <div
          className={styles.downloadBtn}
          onClick={handleDownload}
        >
          <SiderDownloadIcon />
        </div>
        {isDownloadListOpen && (
          <DownloadListBox
            className={styles.downloadListWrapper}
            handleDownload={handleDownload}
          />
        )}
      </div>
    </div>
  );
};

export default Sidebar;
