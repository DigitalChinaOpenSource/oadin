import styles from './index.module.scss';
import { Button, Checkbox, Form, Input } from 'antd';
import useLoginStore from '@/store/loginStore.ts';
import CodeInput from '@/pages/Login/codeInput';
import { useState } from 'react';

interface IForgetPasswordProp {
  email: string;
  code: string;
}

const SetNewPassword = () => {
  return <div>123</div>;
};

const ForgetPassword = () => {
  const { setCurrentStep } = useLoginStore();
  const [form] = Form.useForm<IForgetPasswordProp>();
  const onFinish = (values: IForgetPasswordProp) => {
    setShowForget(false);
  };

  const [showForget, setShowForget] = useState<boolean>(true);

  return (
    <div className={styles.forgetPassword}>
      <div
        className={styles.goBack}
        onClick={() => setCurrentStep('enterpriseAccount')}
      >
        返回
      </div>
      <div className={styles.header}>找回密码</div>
      <div className={styles.formContainer}>
        {showForget ? (
          <Form
            form={form}
            name="emailForm"
            autoComplete="off"
            onFinish={onFinish}
            className={styles.loginForm}
          >
            <Form.Item
              name="email"
              validateTrigger={['onChange', 'onBlur']}
              rules={[
                {
                  validator: (_: unknown, value: string) => {
                    if (!value) {
                      return Promise.reject(new Error('请输入邮箱地址'));
                    }
                    if (!/^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/.test(value)) {
                      return Promise.reject(new Error('邮箱格式错误，请重新输入！'));
                    }
                    return Promise.resolve();
                  },
                  validateTrigger: ['onBlur'],
                },
              ]}
            >
              <Input
                style={{ width: '100%', height: '40px' }}
                placeholder="请输入邮箱地址"
              />
            </Form.Item>
            <CodeInput
              form={form}
              validateFiled={'email'}
              codeFiled={'code'}
            />
            <Form.Item style={{ marginBottom: 0 }}>
              <Button
                htmlType="submit"
                type="primary"
                className={styles.loginButton}
              >
                下一步
              </Button>
            </Form.Item>
          </Form>
        ) : (
          <SetNewPassword />
        )}
      </div>
    </div>
  );
};

export default ForgetPassword;
