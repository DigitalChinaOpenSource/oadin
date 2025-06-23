import styles from './index.module.scss';
import useLoginStore from '@/store/loginStore.ts';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { App, Button, Form, Input, message } from 'antd';
import React from 'react';
import EmailInput from '@/pages/Login/components/emailInput';
import CodeInput from '@/pages/Login/components/codeInput';
import PhoneNumberInput from '@/pages/Login/components/phoneNumberInput';
import AgreedCheckBox from '@/pages/Login/components/agreedCheckBox';
import { useLoginView } from '@/pages/Login/useLoginView.ts';

export interface IEnterpriseFormValues {
  name: string;
  password: string;
  surePassword: string;
  email: string;
  emailCode: string;
  phoneNumber: string;
  phoneCode: string;
  agreed: boolean;
}

const CreateNewAccount = () => {
  const { setCurrentStep } = useLoginStore();
  const { createNewAccount } = useLoginView();
  const [form] = Form.useForm();
  const { message } = App.useApp();

  const onFinish = async (values: IEnterpriseFormValues) => {
    console.log('Received values of form: ', values);
    // 这里可以添加提交表单的逻辑
    // 比如调用API注册新账户
    // onOk?.(values);
    const createRes = await createNewAccount(values);
    if (createRes) {
      message.success('企业账户创建成功');
    }
  };

  return (
    <div className={styles.createNewAccount}>
      <div
        className="goBack"
        onClick={() => setCurrentStep('enterpriseAccount')}
      >
        <ArrowLeftOutlined size={24} />
        <div>返回</div>
      </div>
      <div className="header">企业邮箱注册</div>
      <div className="formContainer">
        <Form
          form={form}
          name={'enterpriseForm'}
          autoComplete={'off'}
          className="loginForm"
          onFinish={onFinish}
          noValidate={true}
        >
          <Form.Item
            name={'name'}
            rules={[{ required: true, message: '请输入企业名称' }]}
          >
            <Input
              className="formInput"
              placeholder={'请输入企业名称'}
            />
          </Form.Item>
          <Form.Item
            name={'password'}
            validateTrigger={['onBlur']}
            rules={[
              { required: true, message: '请设置登录密码' },
              {
                pattern: /^(?![0-9]+$)(?![a-zA-Z]+$)[0-9A-Za-z!@#$%^&*()_+]{8,16}$/,
                message: '密码必须是8-16位英文字母、数字、字符组合',
              },
            ]}
          >
            <Input.Password
              className="formInput"
              placeholder={'请设置登录密码'}
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
              { required: true, message: '请再次输入登录密码' },
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
          <EmailInput
            form={form}
            codeFiled={'email'}
            placeholder="请输入企业邮箱"
          />
          <CodeInput
            form={form}
            validateFiled={'email'}
            codeFiled={'emailCode'}
            placeholder="请输入邮箱验证码"
          />
          <PhoneNumberInput
            form={form}
            codeFiled={'phoneNumber'}
            placeholder="请输入手机号"
          />
          <CodeInput
            form={form}
            validateFiled={'phoneNumber'}
            codeFiled={'phoneCode'}
            placeholder="请输入手机验证码"
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
              确认
            </Button>
          </Form.Item>
        </Form>
      </div>
    </div>
  );
};

export default CreateNewAccount;
