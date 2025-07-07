//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

import { createBrowserRouter, Navigate } from 'react-router-dom';
import { upFirstLetter } from '../utils';
import MainLayout from '../components/main-layout';
import Dashboard from '@/pages/dashboard';
import AboutAOG from '@/pages/about-aog'; // 新增关于AOG页面
import ChooseService from '@/pages/choose-service'; // 新增选择服务页面
import LocalModels from '@/pages/local-models';
import ChooseModule from '@/pages/choose-service/chooseModule.tsx'; // 新增选择模块页面
import RemoteModels from '@/pages/remote-models';
import LocalModelDetail from '@/pages/local-model-detail';
import RemoteModelDetail from '@/pages/remote-model-detail';
import HybridScheduling from '@/pages/hybrid-scheduling'; // 新增混合调度页面
const router = createBrowserRouter([
  {
    path: '/',
    element: <MainLayout />,
    children: [
      {
        path: '/',
        element: (
          <Navigate
            to="/dashboard"
            replace
          />
        ),
      },
      {
        path: '/dashboard',
        element: <Dashboard />,
      },
      {
        path: '/about-aog',
        element: <AboutAOG />,

        children: [
          {
            index: true,
            element: (
              <Navigate
                to="choose-service"
                replace
              />
            ),
          },
          {
            path: 'choose-service',
            element: <ChooseService />,
            handle: { breadcrumb: 'Choose Service' },
            children: [
              {
                path: ':type',
                element: <ChooseModule />,
                handle: {
                  breadcrumb: ({ params }: { params: { type?: string } }) => {
                    // 你可以根据 params.type 返回不同的中文名
                    // const map: Record<string, string> = {
                    //   chat: 'Chat',
                    //   text_to_image: 'Text-to-image',
                    // };
                    // return map[params.type as string] || params.type;
                    return params.type ? upFirstLetter(params.type) : 'Service';
                  },
                },
                children: [
                  {
                    path: 'local-models',
                    element: <LocalModels />,
                    handle: { breadcrumb: 'Local Models' },
                    children: [
                      {
                        path: ':name',
                        element: <LocalModelDetail />,
                        handle: {
                          breadcrumb: ({ params }: { params: { name?: string } }) => {
                            return params.name;
                          },
                        },
                      },
                    ],
                  },
                  {
                    path: 'remote-models',
                    element: <RemoteModels />,
                    handle: { breadcrumb: 'Remote Models' },
                    children: [
                      {
                        path: ':name',
                        element: <RemoteModelDetail />,
                        handle: {
                          breadcrumb: ({ params }: { params: { name?: string } }) => {
                            return params.name;
                          },
                        },
                      },
                    ],
                  },
                  {
                    path: 'hybrid-scheduling',
                    element: <HybridScheduling />,
                    handle: { breadcrumb: 'Hybrid Scheduling' },
                  },
                ],
              },
            ],
          },
        ],
      },
      {
        path: '*',
        element: (
          <Navigate
            to="/dashboard"
            replace
          />
        ),
      },
    ],
  },
]);
const formatRoutes = (routes: any[]) => {
  return routes.map((route) => {
    const { path, children } = route;
    const formattedRoute: any = { path, children };

    if (children) {
      formattedRoute.children = formatRoutes(children);
    }
    return formattedRoute;
  });
};

export default router;
