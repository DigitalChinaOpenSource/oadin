import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import useAuthStore from '../store/authStore';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requireAuth?: boolean;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children, requireAuth = true }) => {
  const { isAuthenticated } = useAuthStore();
  const location = useLocation();

  // 如果需要认证但未登录，重定向到登录页面
  // if (requireAuth && !isAuthenticated) {
  //   return (
  //     <Navigate
  //       to="/login"
  //       state={{ from: location.pathname }}
  //       replace
  //     />
  //   );
  // }

  // 如果已登录且访问登录页，重定向到应用管理
  if (isAuthenticated && location.pathname === '/login') {
    return (
      <Navigate
        to="/app-management"
        replace
      />
    );
  }

  return <>{children}</>;
};

export default ProtectedRoute;
