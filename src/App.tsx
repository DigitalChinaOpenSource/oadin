import { RouterProvider } from 'react-router-dom';
import router from './routes';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import './styles/global.scss';

function App() {
  return (
    <ConfigProvider
      locale={zhCN}
      theme={{ token: { colorPrimary: '#5429ff' } }}
    >
      <RouterProvider router={router} />
    </ConfigProvider>
  );
}

export default App;
