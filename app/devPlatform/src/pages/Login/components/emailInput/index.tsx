import { Form, Input } from 'antd';

const EmailInput = ({ form, codeFiled, placeHolder }: { form: any; codeFiled: string; placeHolder?: string }) => {
  return (
    <Form.Item
      name={codeFiled}
      validateTrigger={['onChange', 'onBlur']}
      rules={[
        {
          validator: (_: unknown, value: string) => {
            if (!value) {
              return Promise.reject(new Error('请输入邮箱地址'));
            }
            if (!/^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/.test(value)) {
              return Promise.reject(new Error('邮箱格式错误，请重新输入！'));
            }
            return Promise.resolve();
          },
          validateTrigger: ['onBlur'],
        },
      ]}
    >
      <Input
        style={{ width: '100%', height: '40px' }}
        placeholder={placeHolder ? placeHolder : '请输入邮箱地址'}
        onChange={(e) => form.setFieldsValue({ [codeFiled]: e.target.value })} // 控制输入
      />
    </Form.Item>
  );
};

export default EmailInput;
