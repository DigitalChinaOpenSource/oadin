import { Form, Input } from 'antd';
import React from 'react';
import styles from '@/pages/Login/loginForm/index.module.scss';

const PhoneNumberInput = ({ form, codeFiled, placeholder }: { form: any; codeFiled: string; placeholder?: string }) => {
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>, type: string) => {
    const value = e.target.value.replace(/\D/g, ''); // 移除非数字字符
    form.setFieldsValue({ [type]: value });
  };

  const PhoneBefore: React.ReactNode = <div className={styles.phoneBefore}>+86</div>;

  return (
    <Form.Item
      name={codeFiled}
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
        className="formInput"
        prefix={PhoneBefore}
        allowClear
        maxLength={11}
        placeholder={placeholder ? placeholder : '请输入手机号'}
        onChange={(e) => handleInputChange(e, 'phoneNumber')} // 控制输入
      />
    </Form.Item>
  );
};

export default PhoneNumberInput;
