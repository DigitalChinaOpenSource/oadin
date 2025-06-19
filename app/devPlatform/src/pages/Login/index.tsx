import React from 'react';
import { Form, Input, Button, Card, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import useAuthStore from '@/store/authStore';

interface LoginFormValues {
  username: string;
  password: string;
}

const Login: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { login } = useAuthStore();

  // 获取重定向来源，如果没有则默认为应用管理页面
  const from = (location.state as { from?: string })?.from || '/app-management';

  const onFinish = (values: LoginFormValues) => {
    // 模拟登录请求
    setTimeout(() => {
      // 模拟成功登录
      if (values.username && values.password) {
        login({
          id: '1',
          name: values.username,
          email: `${values.username}@example.com`,
          token: 'sample-token-123456',
        });

        message.success('登录成功');
        navigate(from, { replace: true });
      } else {
        message.error('用户名或密码错误');
      }
    }, 1000);
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
      <Card
        title="开发平台登录"
        style={{ width: 400 }}
      >
        <Form
          name="login"
          initialValues={{ remember: true }}
          onFinish={onFinish}
          autoComplete="off"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名!' }]}
          >
            <Input
              prefix={<UserOutlined />}
              placeholder="用户名"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码!' }]}
          >
            <Input.Password
              prefix={<LockOutlined />}
              placeholder="密码"
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              block
            >
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default Login;
