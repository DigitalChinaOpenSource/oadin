import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Badge, Menu } from 'antd';
import styles from './index.module.scss';
import { DownloadIcon } from '@phosphor-icons/react';
import DownloadListBox from '../download-list-box';
import useModelDownloadStore from '../../store/useModelDownloadStore';
import mm from '../icons/mm.svg';
import mmac from '../icons/mmac.svg';
import sm from '../icons/sm.svg';
import smac from '../icons/smac.svg';
import mcp from '../icons/mcp.svg';
import mcpac from '../icons/mcpac.svg';
import type { MenuProps } from 'antd';

type MenuItem = Required<MenuProps>['items'][number];

export default function Sidebar({ collapsed }: { collapsed: boolean }) {
  const navigate = useNavigate();
  const location = useLocation();
  // 是否展示下载列表弹窗
  const [isDownloadListOpen, setIsDownloadListOpen] = useState(false);
  const { downloadList } = useModelDownloadStore();

  // 当前选中的菜单项和展开的菜单项
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
  const [openKeys, setOpenKeys] = useState<string[]>([]);

  useEffect(() => {
    if (collapsed) return;
    // 根据当前路由设置选中菜单项和展开的菜单项
    const pathSegments = location.pathname.split('/').filter(Boolean);
    if (location.pathname.startsWith('/mcp-detail')) return;
    if (pathSegments.length > 0) {
      setSelectedKeys([location.pathname]);
      setOpenKeys([...openKeys, pathSegments[0]]);
    }
  }, [location.pathname, collapsed]);

  const menuItems: MenuItem[] = [
    {
      key: 'model-manage',
      label: '模型管理',
      icon: (
        <img
          src={location.pathname.startsWith('/model-manage') ? mmac : mm}
          alt="模型管理"
        />
      ),
      children: [
        { key: '/model-manage/model-list', label: '模型广场' },
        { key: '/model-manage/my-model-list', label: '我的模型' },
        { key: '/model-manage/model-experience', label: '模型体验' },
      ],
    },
    {
      key: 'mcp-service',
      label: 'MCP服务',
      icon: (
        <img
          src={location.pathname.startsWith('/mcp-service') || location.pathname.startsWith('/mcp-detail') ? mcpac : mcp}
          alt="MCP服务"
        />
      ),
      children: [
        { key: '/mcp-service/mcp-list', label: 'MCP广场' },
        { key: '/mcp-service/my-mcp-list', label: '我的MCP' },
      ],
    },
    {
      key: 'settings',
      label: '设置',
      icon: (
        <img
          src={location.pathname.startsWith('/settings') ? smac : sm}
          alt="设置"
        />
      ),
      children: [
        { key: '/settings/model-setting', label: '模型设置' },
        { key: '/settings/agent-setting', label: '代理设置' },
        { key: '/settings/service-provider-manage', label: '服务提供商管理' },
        { key: '/settings/about-us', label: '关于我们' },
      ],
    },
  ];

  const handleMenuClick = ({ key }: { key: string }) => {
    console.log('key------', key);
    navigate(`${key}`);
  };

  const handleDownload = (visible: boolean) => {
    setIsDownloadListOpen(visible);
  };

  return (
    <div className={styles.sidebar}>
      <div className={styles.menuContainer}>
        <Menu
          selectedKeys={selectedKeys} // 动态设置选中项
          openKeys={openKeys} // 动态设置展开项
          mode="inline"
          theme="light"
          items={menuItems}
          onClick={handleMenuClick}
          onOpenChange={(keys) => setOpenKeys(keys)} // 更新展开项
        />
      </div>

      {!!downloadList?.length && (
        <div className={styles.downloadBtnBox}>
          <Badge
            // dot={!!downloadList?.length}
            count={downloadList?.length}
            showZero={false}
            className={styles.badge}
          >
            <div
              className={styles.downloadBtn}
              onClick={() => handleDownload(true)}
            >
              <DownloadIcon
                width={18}
                height={18}
                fill="#ffffff"
              />
            </div>
          </Badge>
          {isDownloadListOpen && downloadList.length > 0 && (
            <DownloadListBox
              className={styles.downloadListWrapper}
              handleDownload={() => handleDownload(false)}
            />
          )}
        </div>
      )}
    </div>
  );
}
