import { createBrowserRouter, Navigate } from 'react-router-dom';
import MainLayout from '../components/main-layout';
import ModelManage from '../pages/model-manage';
import ServerManage from '../pages/server-manage';

const router = createBrowserRouter([
  {
    path: '/',
    element: <MainLayout />,
    children: [
      {
        path: '/',
        element: <Navigate to="/model-manage" replace />,
      },
      {
        path: '/model-manage',
        element: <ModelManage />,
      },
      {
        path: '/server-manage',
        element: <ServerManage />,
      },
    ],
  },
]);

export default router;