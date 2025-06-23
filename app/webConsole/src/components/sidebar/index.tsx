import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Badge, Menu, Popover, Tooltip } from 'antd';
import styles from './index.module.scss';
import { DownloadIcon } from '@phosphor-icons/react';
import DownloadListBox from '../download-list-box';
import useModelDownloadStore from '../../store/useModelDownloadStore';
import mxty from '../icons/mxty.svg';
import mxtyac from '../icons/mxtyac.svg';
import mm from '../icons/mm.svg';
import mmac from '../icons/mmac.svg';
import sm from '../icons/sm.svg';
import smac from '../icons/smac.svg';
import mcp from '../icons/mcp.svg';
import mcpac from '../icons/mcpac.svg';
import type { MenuProps } from 'antd';
import useMcpDownloadStore from '@/store/useMcpDownloadStore.ts';
import McpDownloadBox from '@/components/mcp-download-box';
import mcpAddSvg from '@/components/icons/mcpAdd.svg';
import DownloadSection from './download-section.tsx';

type MenuItem = Required<MenuProps>['items'][number];

export default function Sidebar({ collapsed }: { collapsed: boolean }) {
  const navigate = useNavigate();
  const location = useLocation();
  // 是否展示下载列表弹窗
  const [isDownloadListOpen, setIsDownloadListOpen] = useState(false);
  const downloadList = useModelDownloadStore((state) => state.downloadList);

  // 当前选中的菜单项和展开的菜单项
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
  const [openKeys, setOpenKeys] = useState<string[]>([]);

  // 获取当前下载中和失败的mcp
  const { mcpAddModalShow, setMcpAddModalShow, mcpDownloadList } = useMcpDownloadStore();

  useEffect(() => {
    // 根据当前路由设置选中菜单项和展开的菜单项
    const pathSegments = location.pathname.split('/').filter(Boolean);
    if (location.pathname.startsWith('/mcp-detail')) return;
    if (pathSegments.length > 0) {
      const curPath = location.pathname.endsWith('/') ? location.pathname.slice(0, -1) : location.pathname;
      setSelectedKeys([curPath]);
      setOpenKeys(collapsed ? [] : [pathSegments[0]]);
    }
  }, [location.pathname, collapsed]);

  const menuItems: MenuItem[] = [
    {
      key: '/model-experience',
      label: '体验中心',
      icon: (
        <img
          src={location.pathname.startsWith('/model-experience') ? mxtyac : mxty}
          alt="体验中心"
        />
      ),
    },
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
        // { key: '/model-manage/model-experience', label: '模型体验' },
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
    navigate(`${key}`);
  };

  const handleDownload = (visible: boolean) => {
    setIsDownloadListOpen(visible);
  };

  useEffect(() => {
    if (mcpDownloadList.length === 0) {
      setMcpAddModalShow(false);
    }
  }, [mcpDownloadList]);

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

      {mcpDownloadList.length > 0 && (
        <Popover
          zIndex={1300}
          content={<McpDownloadBox />}
          trigger={'click'}
          placement={'rightTop'}
          arrow={false}
          open={mcpAddModalShow}
          onOpenChange={setMcpAddModalShow}
        >
          <div className={styles.mcpDownloadBox}>
            <div className={styles.mpcDownloadBoxContent}>
              <Tooltip title={collapsed ? '添加中' : ''}>
                <img
                  src={mcpAddSvg}
                  alt=""
                />
              </Tooltip>
              {!collapsed && <div>添加中</div>}
            </div>
          </div>
        </Popover>
      )}
      <DownloadSection collapsed={collapsed} />
    </div>
  );
}
