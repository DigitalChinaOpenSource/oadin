import { createElement, lazy, ReactNode } from 'react';
import { AppstoreOutlined, UserOutlined } from '@ant-design/icons';

export interface RouteConfig {
  path: string;
  name: string;
  icon?: ReactNode;
  component?: React.LazyExoticComponent<React.ComponentType<any>>;
  children?: RouteConfig[];
  hideInMenu?: boolean;
  requireAuth?: boolean;
  menuPath?: string; // 用于在子路由中指定激活的菜单项
}

const AppManagementPage = lazy(() => import('../pages/AppManagement/AppList'));
const AppConfigPage = lazy(() => import('../pages/AppManagement/AppConfig'));
const UserCenterPage = lazy(() => import('../pages/UserCenter'));
const LoginPage = lazy(() => import('../pages/Login/index'));

const routes: RouteConfig[] = [
  {
    path: '/app-management',
    name: '应用管理',
    icon: createElement(AppstoreOutlined),
    component: AppManagementPage,
    requireAuth: true,
  },
  {
    path: '/app-management/config/:id',
    name: '配置应用',
    component: AppConfigPage,
    requireAuth: true,
    hideInMenu: true,
    menuPath: '/app-management', // 指定在此路由激活时，菜单应该高亮哪个路径
  },
  {
    path: '/user-center',
    name: '账号中心',
    icon: createElement(UserOutlined),
    component: UserCenterPage,
    requireAuth: true,
  },
  {
    path: '/login',
    name: '登录',
    component: LoginPage,
    hideInMenu: true,
    requireAuth: false,
  },
];

export default routes;
