import { Form, Input } from 'antd';

const EmailInput = ({ form, codeFiled, placeholder }: { form: any; codeFiled: string; placeholder?: string }) => {
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
        className="formInput"
        placeholder={placeholder ? placeholder : '请输入邮箱地址'}
        onChange={(e) => form.setFieldsValue({ [codeFiled]: e.target.value })} // 控制输入
      />
    </Form.Item>
  );
};

export default EmailInput;
