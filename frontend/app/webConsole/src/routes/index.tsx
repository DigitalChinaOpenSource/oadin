import { createBrowserRouter, Navigate, useLocation } from 'react-router-dom';
import { useEffect } from 'react';
import MainLayout from '../components/main-layout';
import McpSquareTab from '@/components/mcp-manage/mcp-square-tab';
import MyMcpTab from '@/components/mcp-manage/my-mcp-tab';
import McpDetail from '../components/mcp-manage/mcp-detail';
import ModelChat from '../pages/model-chat';
import ServiceProviderManage from '@/components/service-provider-tab';
import ModelSetting from '@/components/settings/model-setting';
import AgentSetting from '@/components/settings/agent-setting';
import AboutUs from '@/components/settings/about-us';
import { ModelSquare } from '@/components/choose-model-dialog/modelSquare.tsx';
import { MyModel } from '@/components/choose-model-dialog/myModel.tsx';

// 自定义组件：处理尾部斜杠重定向
const TrailingSlashHandler: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const location = useLocation();

  useEffect(() => {
    // 如果路径以斜杠结尾且不是根路径，则重定向到不带斜杠的版本
    if (location.pathname !== '/' && location.pathname.endsWith('/')) {
      const newPath = location.pathname.slice(0, -1) + location.search + location.hash;
      window.history.replaceState(null, '', newPath);
    }
  }, [location]);

  return <>{children}</>;
};

const WrappedMainLayout: React.FC = () => {
  return (
    <TrailingSlashHandler>
      <MainLayout />
    </TrailingSlashHandler>
  );
};

const router = createBrowserRouter([
  {
    path: '/',
    element: <WrappedMainLayout />,
    children: [
      {
        path: '/',
        element: (
          <Navigate
            to="/model-manage/model-list"
            replace
          />
        ),
      },
      //体验中心
      {
        path: '/model-experience',
        element: <ModelChat />,
      },
      // 模型管理
      {
        path: '/model-manage',
        children: [
          {
            path: 'model-list',
            element: <ModelSquare isDialog={false} />,
          },
          {
            path: 'my-model-list',
            element: <MyModel isDialog={false} />,
          },
        ],
      },
      // 设置
      {
        path: '/settings',
        // element: <ServerManage />,
        children: [
          { path: 'model-setting', element: <ModelSetting /> }, // 模型设置
          { path: 'agent-setting', element: <AgentSetting /> }, // 代理设置
          {
            path: 'service-provider-manage',
            element: <ServiceProviderManage />, // 服务提供商管理
          },
          {
            path: 'about-us',
            element: <AboutUs />, // 关于我们
          },
        ],
      },
      // MCP服务
      {
        path: '/mcp-service',
        children: [
          {
            path: 'mcp-list',
            element: <McpSquareTab />,
          },
          {
            path: 'my-mcp-list',
            element: <MyMcpTab />,
          },
        ],
      },
      //MCP详情
      {
        path: '/mcp-detail',
        element: <McpDetail />,
      },
      {
        path: '*',
        element: (
          <Navigate
            to="/model-manage/model-list"
            replace
          />
        ),
      },
    ],
  },
]);

export default router;
