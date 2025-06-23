import styles from './index.module.scss';
import { Button, Checkbox, Form, Input } from 'antd';
import useLoginStore from '@/store/loginStore.ts';
import CodeInput from '@/pages/Login/components/codeInput';
import { useState } from 'react';
import EmailInput from '@/pages/Login/components/emailInput';
import { ArrowLeftOutlined } from '@ant-design/icons';

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
        className="goBack"
        onClick={() => setCurrentStep('enterpriseAccount')}
      >
        <ArrowLeftOutlined size={24} />
        <div>返回</div>
      </div>
      <div className="header">找回密码</div>
      <div className="formContainer">
        {showForget ? (
          <Form
            form={form}
            name="emailForm"
            autoComplete="off"
            onFinish={onFinish}
            className="loginForm"
          >
            <EmailInput
              form={form}
              codeFiled={'email'}
            />
            <CodeInput
              form={form}
              validateFiled={'email'}
              codeFiled={'code'}
            />
            <Form.Item style={{ marginBottom: 0 }}>
              <Button
                htmlType="submit"
                type="primary"
                className="loginButton"
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
