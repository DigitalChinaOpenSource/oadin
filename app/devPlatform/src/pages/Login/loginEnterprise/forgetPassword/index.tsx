import styles from './index.module.scss';
import { Button, Checkbox, Form, Input, message } from 'antd';
import useLoginStore from '@/store/loginStore.ts';
import CodeInput from '@/pages/Login/components/codeInput';
import React, { useRef, useState } from 'react';
import EmailInput from '@/pages/Login/components/emailInput';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useLoginView } from '@/pages/Login/useLoginView.ts';

interface IForgetPasswordProp {
  email: string;
  emailCode: string;
}

interface ISetNewFormProps {
  password: string;
  surePassword: string;
}

const SetNewPassword = ({ email }: { email: string }) => {
  const [form] = Form.useForm();
  const { resetPassword } = useLoginView();

  const onFinish = async (values: ISetNewFormProps) => {
    console.log('Received values of form: ', values, email);
    const res = await resetPassword(email, values.password);
    if (res) {
      message.success('密码重置成功，请使用新密码登录');
    }
  };
  return (
    <Form
      noValidate={true}
      name="setNewPasswordForm"
      form={form}
      autoComplete="off"
      onFinish={onFinish}
      className="loginForm"
    >
      <Form.Item
        name={'password'}
        validateTrigger={['onBlur']}
        rules={[
          { required: true, message: '请输入新的登录密码' },
          {
            pattern: /^(?![0-9]+$)(?![a-zA-Z]+$)[0-9A-Za-z!@#$%^&*()_+]{8,16}$/,
            message: '密码必须是8-16位英文字母、数字、字符组合',
          },
        ]}
        extra={<span className={styles.passwordDesc}>密码必须是8-16位英文字母、数字、字符组合(不能是纯数字)</span>}
      >
        <Input.Password
          className="formInput"
          placeholder={'请设置新的登录密码'}
          minLength={8}
          maxLength={16}
          allowClear
        />
      </Form.Item>
      <Form.Item
        name={'surePassword'}
        dependencies={['password']}
        validateTrigger={['onBlur', 'onChange']}
        rules={[
          { required: true, message: '请确认新的登录密码' },
          ({ getFieldValue }) => ({
            validator(_, value) {
              if (!value || getFieldValue('password') === value) {
                return Promise.resolve();
              }
              return Promise.reject(new Error('两次输入的密码不一致'));
            },
          }),
        ]}
      >
        <Input.Password
          className="formInput"
          placeholder={'请再次输入登录密码'}
          minLength={8}
          maxLength={16}
          allowClear
        />
      </Form.Item>
      <Form.Item style={{ marginBottom: 0 }}>
        <Button
          htmlType="submit"
          type="primary"
          className="loginButton"
        >
          确认
        </Button>
      </Form.Item>
    </Form>
  );
};

const ForgetPassword = () => {
  const { setCurrentStep } = useLoginStore();
  const [form] = Form.useForm<IForgetPasswordProp>();
  const { getPassWordWithEmailCode } = useLoginView();
  const onFinish = async (values: IForgetPasswordProp) => {
    const res = await getPassWordWithEmailCode(values.email, values.emailCode);
    emailRef.current = values.email;
    if (res) {
      setShowForget(false);
    }
  };

  const [showForget, setShowForget] = useState<boolean>(true);
  const emailRef: any = useRef(null);

  return (
    <div className={styles.forgetPassword}>
      <div
        className="goBack"
        onClick={() => setCurrentStep('enterpriseAccount')}
      >
        <ArrowLeftOutlined size={24} />
        <div>返回</div>
      </div>
      <div className="header">{showForget ? '找回密码' : '设置新密码'}</div>
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
              codeFiled={'emailCode'}
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
          <SetNewPassword email={emailRef.current} />
        )}
      </div>
    </div>
  );
};

export default ForgetPassword;
