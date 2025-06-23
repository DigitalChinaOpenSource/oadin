import React, { useState, useRef } from 'react';
import { Input, InputNumber, Button, Form, Checkbox } from 'antd';
import styles from './index.module.scss';
import CodeInput from '@/pages/Login/components/codeInput';
import PhoneNumberInput from '@/pages/Login/components/phoneNumberInput';
import AgreedCheckBox from '@/pages/Login/components/agreedCheckBox';

export interface LoginFormValues {
  phoneNumber: string;
  code: string;
  agreed?: boolean;
}

interface LoginFormProps {
  type?: 'login' | 'bind';
  onOk?: (values: LoginFormValues) => void;
  showAgreed?: boolean;
}

const LoginForm: React.FC<LoginFormProps> = ({ type = 'login', onOk, showAgreed = true }) => {
  const [form] = Form.useForm<LoginFormValues>();
  const onFinish = (values: LoginFormValues) => {
    console.log('Received values of form: ', values);
    onOk?.(values);
  };

  return (
    <Form
      form={form}
      name="loginForm"
      autoComplete="off"
      onFinish={onFinish}
      className="loginForm"
    >
      <PhoneNumberInput
        form={form}
        codeFiled={'phoneNumber'}
      />
      <CodeInput
        form={form}
        validateFiled={'phoneNumber'}
        codeFiled={'code'}
      />
      {showAgreed && (
        <AgreedCheckBox
          form={form}
          codeFiled={'agreed'}
        />
      )}
      <Form.Item style={{ marginBottom: 0 }}>
        <Button
          htmlType="submit"
          type="primary"
          className="loginButton"
        >
          {type === 'login' ? '登录/注册' : '确认绑定'}
        </Button>
      </Form.Item>
    </Form>
  );
};

export default LoginForm;
