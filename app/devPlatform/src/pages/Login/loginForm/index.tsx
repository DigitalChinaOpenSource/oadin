import React, { useState, useRef, useEffect } from 'react';
import { Input, InputNumber, Button, Form, Checkbox } from 'antd';
import styles from './index.module.scss';
import CodeInput from '@/pages/Login/components/codeInput';
import PhoneNumberInput from '@/pages/Login/components/phoneNumberInput';
import AgreedCheckBox from '@/pages/Login/components/agreedCheckBox';
import { useLoginView } from '@/pages/Login/useLoginView.ts';
import useLoginStore from '@/store/loginStore.ts';
import useAuthStore, { User } from '@/store/authStore.ts';
import { useNavigate } from 'react-router-dom';

export interface LoginFormValues {
  phone: string;
  verifyCode: string;
  agreed: boolean;
}

const LoginForm = () => {
  const [form] = Form.useForm<LoginFormValues>();
  const { setPersonPhoneData, personPhoneData, setCurrentStep } = useLoginStore();
  const navigate = useNavigate();
  const { loginWithPhone } = useLoginView();
  const { login } = useAuthStore();
  const onFinish = async (values: LoginFormValues) => {
    console.log('Received values of form: ', values);
    const userRes: any = await loginWithPhone(values.phone, values.verifyCode, values.agreed);
    console.log('userRes', userRes);

    if (userRes.data) {
      login(userRes as User);
      if (userRes.needRealName) {
        setCurrentStep('personAuth');
      } else {
        navigate('/app-management');
      }
    }
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
        codeFiled={'phone'}
      />
      <CodeInput
        form={form}
        validateFiled={'phone'}
        codeFiled={'verifyCode'}
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
