import React, { useState, useRef, useEffect } from 'react';
import { Input, InputNumber, Button, Form, Checkbox } from 'antd';
import styles from './index.module.scss';
import CodeInput from '@/pages/Login/components/codeInput';
import PhoneNumberInput from '@/pages/Login/components/phoneNumberInput';
import AgreedCheckBox from '@/pages/Login/components/agreedCheckBox';
import { useLoginView } from '@/pages/Login/useLoginView.ts';
import useLoginStore from '@/store/loginStore.ts';

export interface LoginFormValues {
  phoneNumber: string;
  code: string;
  agreed: boolean;
}

interface LoginFormProps {
  type?: 'login' | 'bind';
  onOk?: (values: LoginFormValues) => void;
  showAgreed?: boolean;
}

const LoginForm: React.FC<LoginFormProps> = () => {
  const [form] = Form.useForm<LoginFormValues>();
  const { setPersonPhoneData, personPhoneData } = useLoginStore();
  const { loginWithPhone } = useLoginView();
  const onFinish = async (values: LoginFormValues) => {
    console.log('Received values of form: ', values);
    await loginWithPhone(values.phoneNumber, values.code, values.agreed);
  };

  // 处理表单值变化的函数
  const handleValuesChange = (_: any, allValues: LoginFormValues) => {
    console.log('Form values changed: ', allValues);
    setPersonPhoneData(allValues);
  };

  useEffect(() => {
    if (personPhoneData) {
      form.setFieldsValue(personPhoneData);
    }
  }, []);

  return (
    <Form
      form={form}
      name="loginForm"
      autoComplete="off"
      onFinish={onFinish}
      onValuesChange={handleValuesChange}
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

      <AgreedCheckBox
        form={form}
        codeFiled={'agreed'}
      />

      <Form.Item style={{ marginBottom: 0 }}>
        <Button
          htmlType="submit"
          type="primary"
          className="loginButton"
        >
          登录/注册
        </Button>
      </Form.Item>
    </Form>
  );
};

export default LoginForm;
