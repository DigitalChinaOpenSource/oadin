import { createBrowserRouter, Navigate } from 'react-router-dom';
import MainLayout from '../components/main-layout';
import ModelManage from '../pages/model-manage';
import ServerManage from '../pages/server-manage';
import McpService from '../pages/mcp-service';
import McpDetail from '../components/mcp-manage/mcp-detail';
import ModelChat from '../pages/model-chat'; // 新增对话页面

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
      {
        path: '/model-manage',
        element: <ModelManage />,
      },
      {
        path: '/server-manage',
        element: <ServerManage />,
      },
      {
        path: '/mcp-service',
        element: <McpService />,
      },
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
