import React, { useState, useEffect } from 'react';
import { Layout, Menu } from 'antd';
import type { MenuProps } from 'antd';
import { useNavigate, useLocation } from 'react-router-dom';
import routes, { RouteConfig } from '../routes/routes';

const { Sider } = Layout;

const Sidebar: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const [collapsed, setCollapsed] = useState(false);
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
  const [openKeys, setOpenKeys] = useState<string[]>([]);
  // 当路径变化时更新选中的菜单项
  useEffect(() => {
    const pathKey = location.pathname;
    
    // 查找当前路径匹配的路由配置
    const findMatchingRoute = (path: string): RouteConfig | null => {
      // 精确匹配
      const exactMatch = routes.find(route => route.path === path);
      if (exactMatch) return exactMatch;
      
      // 参数路径匹配 (如 /app-management/config/123)
      return routes.find(route => {
        if (!route.path) return false;
        const routePathSegments = route.path.split('/');
        const currentPathSegments = path.split('/');
        
        if (routePathSegments.length !== currentPathSegments.length) return false;
        
        for (let i = 0; i < routePathSegments.length; i++) {
          if (routePathSegments[i].startsWith(':')) continue;
          if (routePathSegments[i] !== currentPathSegments[i]) return false;
        }
        return true;
      }) || null;
    };
    
    const matchedRoute = findMatchingRoute(pathKey);
    
    // 如果找到匹配的路由，并且它有menuPath配置，则使用menuPath作为选中项
    if (matchedRoute && matchedRoute.menuPath) {
      setSelectedKeys([matchedRoute.menuPath]);
    } else {
      setSelectedKeys([pathKey]);
    }

    // 找到当前路径所在的父菜单，并自动展开
    const findParentPath = (routeList: RouteConfig[], path: string): string | null => {
      for (const route of routeList) {
        if (route.children) {
          for (const child of route.children) {
            if (child.path === path) {
              return route.path;
            }
          }

          const foundInChildren = findParentPath(route.children, path);
          if (foundInChildren) return route.path;
        }
      }
      return null;
    };

    // 使用真实路径或menuPath查找父路径
    const currentPath = matchedRoute?.menuPath || pathKey;
    const parentPath = findParentPath(routes, currentPath);
    if (parentPath) {
      setOpenKeys([parentPath]);
    }
  }, [location.pathname]);

  // 生成菜单项
  const getMenuItems = (routeList: RouteConfig[]): MenuProps['items'] => {
    return routeList
      .filter((route) => !route.hideInMenu)
      .map((route) => {
        if (route.children?.length) {
          return {
            key: route.path,
            label: route.name,
            icon: route.icon,
            children: getMenuItems(route.children),
          };
        }

        return {
          key: route.path,
          label: route.name,
          icon: route.icon,
          onClick: () => navigate(route.path),
        };
      });
  };

  // 准备菜单项
  const menuItems = getMenuItems(routes);

  return (
    <Sider
      collapsible
      collapsed={collapsed}
      onCollapse={(value) => setCollapsed(value)}
      theme="light"
      style={{ background: 'rgba(255, 255, 255, 0.5)' }}
    >
      <Menu
        mode="inline"
        items={menuItems}
        selectedKeys={selectedKeys}
        openKeys={openKeys}
        onOpenChange={(keys) => setOpenKeys(keys as string[])}
      />
    </Sider>
  );
};

export default Sidebar;
