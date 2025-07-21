import { RouterProvider } from 'react-router-dom';
import router from './routes';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import './styles/global.scss';
import '@ant-design/v5-patch-for-react-19';

function App() {
  return (
    <ConfigProvider
      locale={zhCN}
      theme={{ token: { colorPrimary: '#4f4dff', colorLink: '#4f4dff' } }}
    >
      <RouterProvider router={router} />
    </ConfigProvider>
  );
}

export default App;
