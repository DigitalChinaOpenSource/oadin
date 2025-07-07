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

import { RouterProvider } from 'react-router-dom';
import router from './routes';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import './styles/global.scss';

function App() {
  return (
    <ConfigProvider
      // locale={zhCN}
      theme={{
        token: { colorPrimary: '#0068B5', colorLink: '#0068B5', colorLinkHover: '#004A86', colorText: '#262626', borderRadius: 0, fontSize: 16 },
        components: {
          Button: {
            /* 这里是你的组件 token */
            defaultBg: '#404040',
            defaultBorderColor: '#404040',
            defaultColor: '#ffffff',
            defaultHoverColor: '#ffffff',
            defaultHoverBg: '#565656',
            defaultHoverBorderColor: '#565656',
            defaultActiveBg: '#333333',
            defaultActiveBorderColor: '#333333',
            defaultActiveColor: '#ffffff',
          },
        },
      }}
    >
      <RouterProvider router={router} />
    </ConfigProvider>
  );
}

export default App;
