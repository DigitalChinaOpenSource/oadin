import { Checkbox, Form } from 'antd';
import React from 'react';

const AgreedCheckBox = ({ form, codeFiled }: { form: any; codeFiled: string }) => {
  return (
    <Form.Item
      name={codeFiled}
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
      <Checkbox
        onChange={(e) => {
          console.log('Checkbox changed:', e.target.checked);
          form.setFieldsValue({ [codeFiled]: e.target.checked });
        }}
      >
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
  );
};

export default AgreedCheckBox;
