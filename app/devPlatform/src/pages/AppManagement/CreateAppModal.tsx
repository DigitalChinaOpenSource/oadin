import React from 'react';
import { Form, Input, Modal } from 'antd';

export interface CreateAppModalProps {
  open: boolean;
  onCancel: () => void;
  onFinish: (values: { name: string }) => Promise<void>;
  confirmLoading?: boolean;
  initialValues?: { name: string };
  title?: string;
}

const CreateAppModal: React.FC<CreateAppModalProps> = ({ open, onCancel, onFinish, confirmLoading = false }) => {
  const [form] = Form.useForm();

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      await onFinish(values);
    } catch (error) {
      console.error('Form validation failed:', error);
    }
  };

  return (
    <Modal
      title={'创建表单'}
      open={open}
      onCancel={onCancel}
      width={480}
      onOk={handleSubmit}
      confirmLoading={confirmLoading}
    >
      <div>
        <Form form={form}>
          <Form.Item
            name="name"
            label="应用名称"
            rules={[
              { required: true, message: '请输入应用名称' },
              {
                pattern: /^[\u4e00-\u9fa5a-zA-Z0-9]+$/,
                message: '支持汉字、数字、字母',
              },
              {
                max: 50,
                message: '不超过50个字符',
              },
            ]}
          >
            <Input
              placeholder="请输入"
              maxLength={50}
            />
          </Form.Item>
        </Form>
      </div>
    </Modal>
  );
};

export default CreateAppModal;
