import React, { Suspense } from 'react';
import { BrowserRouter, Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { ConfigProvider, Layout, Spin } from 'antd';
import Sidebar from './components/Sidebar';
import ProtectedRoute from './components/ProtectedRoute';
import routes from './routes/routes';
import styles from './index.module.scss';
import TopHeader from './components/top-header';
import zhCN from 'antd/locale/zh_CN';

const { Header, Content, Footer } = Layout;

// 主布局组件
const MainLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const location = useLocation();
  // 登录页面不显示侧边栏和头部
  const isLoginPage = location.pathname === '/login';

  if (isLoginPage) {
    return <>{children}</>;
  }

  return (
    <Layout className={styles.mainLayout}>
      <Header className={styles.header}>
        <TopHeader />
      </Header>
      <Layout>
        <Sidebar />
        <Layout>
          <Content className={styles.content}>{children}</Content>
          {/*<Footer style={{ textAlign: 'center' }}>DevPlatform ©{new Date().getFullYear()} Created by Vanta</Footer>*/}
        </Layout>
      </Layout>
    </Layout>
  );
};

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <Suspense
        fallback={
          <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
            <Spin size="large" />
          </div>
        }
      >
        <ConfigProvider
          locale={zhCN}
          theme={{ token: { colorPrimary: '#4f4dff', colorLink: '#4f4dff' } }}
        >
          <MainLayout>
            <Routes>
              {/* 根路由重定向到应用管理 */}
              <Route
                path="/"
                element={
                  <Navigate
                    to="/app-management"
                    replace
                  />
                }
              />

              {/* 动态生成路由 */}
              {renderRoutes(routes)}

              {/* 404页面 */}
              <Route
                path="*"
                element={<div>404 Not Found</div>}
              />
            </Routes>
          </MainLayout>
        </ConfigProvider>
      </Suspense>
    </BrowserRouter>
  );
};

// 递归生成路由配置
const renderRoutes = (routeList: typeof routes) => {
  return routeList.map((route) => {
    const Component = route.component;

    if (route.children?.length) {
      return <React.Fragment key={route.path}>{renderRoutes(route.children)}</React.Fragment>;
    }

    return (
      <Route
        key={route.path}
        path={route.path}
        element={<ProtectedRoute requireAuth={route.requireAuth}>{Component && <Component />}</ProtectedRoute>}
      />
    );
  });
};

export default App;
