import { createBrowserRouter, Navigate } from 'react-router-dom';
import MainLayout from '../components/main-layout';
import ModelManage from '../pages/model-manage';
import ServerManage from '../pages/server-manage';
import McpService from '../pages/mcp-service';
import McpSquareTab from '@/components/mcp-manage/mcp-square-tab';
import MyMcpTab from '@/components/mcp-manage/my-mcp-tab';
import McpDetail from '../components/mcp-manage/mcp-detail';
import ModelChat from '../pages/model-chat'; // 新增对话页面
import ServiceProviderManage from '@/components/service-provider-tab';
import ByzeServiceTab from '@/components/byze-service-tab';
import ModelSetting from '@/components/settings/model-setting';
import AgentSetting from '@/components/settings/agent-setting';
import AboutUs from '@/components/settings/about-us';

const router = createBrowserRouter([
  {
    path: '/',
    element: <MainLayout />,
    children: [
      {
        path: '/',
        element: (
          <Navigate
            to="/model-manage"
            replace
          />
        ),
      },
      // 模型管理
      {
        path: '/model-manage',
        children: [
          {
            path: 'model-list',
            element: <ModelManage />,
          },
          {
            path: 'my-model-list',
            element: <MyMcpTab />,
          },
          {
            path: 'model-experience',
            element: <MyMcpTab />,
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
        path: '/model-chat',
        element: <ModelChat />, // 新增对话页面
      },
    ],
  },
]);

export default router;
