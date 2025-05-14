import { createBrowserRouter, Navigate } from 'react-router-dom';
import MainLayout from '../components/main-layout';
import ModelManage from '../pages/model-manage';
import ServerManage from '../pages/server-manage';
import McpManage from '../pages/mcp-manage';

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
        path: '/mcp-manage',
        element: <McpManage />,
      },
    ],
  },
]);

export default router;
