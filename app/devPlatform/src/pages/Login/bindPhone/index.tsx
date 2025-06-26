import styles from './index.module.scss';
import { Button, Form } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import React from 'react';
import PhoneNumberInput from '@/pages/Login/components/phoneNumberInput';
import CodeInput from '@/pages/Login/components/codeInput';
import useLoginStore from '@/store/loginStore.ts';
import { useNavigate } from 'react-router-dom';
import { IBasePhoneFormProps } from '@/pages/Login/types';

interface IBindPhoneProps {
  onConfirm: (values: IBasePhoneFormProps) => void;
}

const BindPhone = ({ onConfirm }: IBindPhoneProps) => {
  const [form] = Form.useForm();
  const { setCurrentStep } = useLoginStore();
  const navigate = useNavigate();

  const onFinish = (values: IBasePhoneFormProps) => {
    console.log('Received values of form: ', values);
    onConfirm(values);
  };

  return (
    <div className={styles.bindPhone}>
      <div
        className="goBack"
        onClick={() => {
          // 清除浏览器地址到登录页
          navigate('/login');
          setCurrentStep('personWechat');
        }}
      >
        <ArrowLeftOutlined size={24} />
        <div>返回</div>
      </div>
      <div className="headerSection">
        <div className="tabContainer">
          <div className="tabItem">
            <span className="tabTitle">手机号绑定</span>
          </div>
        </div>
        <span className="subTitle">为了保障您的账号安全和数据互通，请绑定手机号</span>
      </div>
      <div className="formContainer">
        <Form
          name={'bindPhoneForm'}
          autoComplete={'off'}
          onFinish={onFinish}
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
          <Form.Item style={{ marginBottom: 0 }}>
            <Button
              htmlType="submit"
              type="primary"
              className="loginButton"
            >
              确认登录
            </Button>
          </Form.Item>
        </Form>
      </div>
    </div>
  );
};

export default BindPhone;
