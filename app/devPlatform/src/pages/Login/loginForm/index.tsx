import React, { useState, useRef } from 'react';
import { Input, InputNumber, Button, Form, Checkbox } from 'antd';
import styles from './index.module.scss';

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
  // 60s倒计时
  const [countTime, setCountTime] = useState(60);
  const [showTimeStatus, setShowTimeStatus] = useState<1 | 2 | 3>(1); // 1 显示获取验证码 2 显示倒计时 3显示重新获取
  const timerInterval = useRef<NodeJS.Timeout | null>(null);
  const createTimer = () => {
    if (timerInterval.current) {
      clearInterval(timerInterval.current);
    }
    timerInterval.current = setInterval(() => {
      setCountTime((prevCountTime) => {
        if (prevCountTime === 1) {
          setShowTimeStatus(3);
          if (timerInterval.current) {
            clearInterval(timerInterval.current);
          }
          return 60;
        }
        return prevCountTime - 1;
      });
    }, 1000);
  };
  // 获取验证码
  const getCode = async (e: React.MouseEvent<HTMLDivElement>) => {
    e.stopPropagation();
    e.preventDefault();
    if (countTime !== 60) {
      return;
    }
    try {
      form.resetFields(['code']);
      await form.validateFields(['phoneNumber']);
      setShowTimeStatus(2);
      createTimer();
      // 这里是获取验证码
      //   const { message } = await System.getPhoneCode({ phone: form.getFieldValue('phoneNumber') })
      //   message.success(message || '验证码已发送')
    } catch (error) {
      console.log('error', error);
    }
  };

  const delay = (ms: number): Promise<void> => new Promise((resolve) => setTimeout(resolve, ms));

  const checkCode = async (): Promise<{ error: string }> => {
    // await delay(2000)
    return { error: '验证码错误验证码错误' };
  };
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>, type: 'phoneNumber' | 'code') => {
    const value = e.target.value.replace(/\D/g, ''); // 移除非数字字符
    form.setFieldsValue({ [type]: value });
  };
  const PhoneBefore: React.ReactNode = <div className={styles.phoneBefore}>+86</div>;

  const CodeAfter: React.FC = () => {
    return (
      <div className={styles.codeAfter}>
        <div
          className={styles.codeContent}
          onClick={getCode}
        >
          {(showTimeStatus === 1 || showTimeStatus === 3) && <div>获取验证码</div>}
          {showTimeStatus === 2 && <div className={styles.codeNotAllow}>{countTime}s</div>}
        </div>
      </div>
    );
  };

  return (
    <Form
      form={form}
      name="loginForm"
      autoComplete="off"
      onFinish={onFinish}
      className={styles.loginForm}
    >
      <Form.Item
        name="phoneNumber"
        validateTrigger={['onChange', 'onBlur']}
        rules={[
          {
            validator: (_: unknown, value: string) => {
              if (!value) {
                return Promise.reject(new Error('请输入手机号'));
              }
              if (!/^1[3-9]\d{9}$/.test(value)) {
                return Promise.reject(new Error('手机号错误，请重新输入！'));
              }
              return Promise.resolve();
            },
            validateTrigger: ['onBlur'],
            // message: '请输入手机号',
          },
        ]}
      >
        <Input
          style={{ width: '100%', height: '40px' }}
          prefix={PhoneBefore}
          allowClear
          maxLength={11}
          placeholder="请输入手机号"
          onChange={(e) => handleInputChange(e, 'phoneNumber')} // 控制输入
        />
      </Form.Item>
      <Form.Item
        name="code"
        validateTrigger={['onBlur']}
        rules={[
          {
            validator: (_: unknown, value: string) => {
              if (!value) {
                return Promise.reject(new Error('请输入验证码'));
              }
              if (value.length !== 6) {
                return Promise.reject(new Error('验证码格式错误！'));
              }
              return Promise.resolve();
            },
          },
        ]}
      >
        <Input
          style={{ width: '100%', height: '40px' }}
          maxLength={6}
          placeholder="请输入验证码"
          suffix={<CodeAfter />}
          allowClear
          onChange={(e) => handleInputChange(e, 'code')} // 控制输入
        />
      </Form.Item>
      {/* 同意协议复选框 */}
      {showAgreed && (
        <Form.Item
          name="agreed"
          valuePropName="checked"
          initialValue={false}
          rules={[
            {
              validator: (_, value) => {
                return value ? Promise.resolve() : Promise.reject(new Error('请阅读并同意用户协议和隐私政策'));
              },
            },
          ]}
        >
          <Checkbox>
            我已阅读并同意 OADIN{' '}
            <a
              href={'#'}
              target="_blank"
              rel="noopener noreferrer"
            >
              用户协议
            </a>{' '}
            和{' '}
            <a
              href={'#'}
              target="_blank"
              rel="noopener noreferrer"
            >
              隐私政策
            </a>
          </Checkbox>
        </Form.Item>
      )}
      <Form.Item style={{ marginBottom: 0 }}>
        <Button
          htmlType="submit"
          type="primary"
          className={styles.loginButton}
        >
          {type === 'login' ? '登录/注册' : '确认绑定'}
        </Button>
      </Form.Item>
    </Form>
  );
};

export default LoginForm;
