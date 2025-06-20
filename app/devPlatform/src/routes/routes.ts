import { ReactNode, lazy, createElement } from 'react';
import { 
  AppstoreOutlined,
  UserOutlined,
} from '@ant-design/icons';

export interface RouteConfig {
  path: string;
  name: string;
  icon?: ReactNode;
  component?: React.LazyExoticComponent<React.ComponentType<any>>;
  children?: RouteConfig[];
  hideInMenu?: boolean;
  requireAuth?: boolean;
}

const AppManagementPage = lazy(() => import('../pages/AppManagement'));
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
    path: '/user-center',
    name: '用户中心',
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
