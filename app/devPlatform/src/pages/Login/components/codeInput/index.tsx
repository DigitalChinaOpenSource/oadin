import { Form, Input } from 'antd';
import React, { useRef, useState } from 'react';
import styles from '@/pages/Login/loginForm/index.module.scss';

const CodeInput = ({ form, validateFiled, codeFiled, placeHolder }: { form: any; validateFiled: string; codeFiled: string; placeHolder?: string }) => {
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
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>, type: string) => {
    const value = e.target.value.replace(/\D/g, ''); // 移除非数字字符
    form.setFieldsValue({ [type]: value });
  };
  // 获取验证码
  const getCode = async (e: React.MouseEvent<HTMLDivElement>) => {
    e.stopPropagation();
    e.preventDefault();
    if (countTime !== 60) {
      return;
    }
    try {
      form.resetFields([codeFiled]);
      await form.validateFields([validateFiled]);
      setShowTimeStatus(2);
      createTimer();
      // 这里是获取验证码
      //   const { message } = await System.getPhoneCode({ phone: form.getFieldValue('phoneNumber') })
      //   message.success(message || '验证码已发送')
    } catch (error) {
      console.log('error', error);
    }
  };

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
    <Form.Item
      name={codeFiled}
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
        placeholder={placeHolder ? placeHolder : '请输入验证码'}
        suffix={<CodeAfter />}
        allowClear
        onChange={(e) => handleInputChange(e, codeFiled)} // 控制输入
      />
    </Form.Item>
  );
};

export default CodeInput;
