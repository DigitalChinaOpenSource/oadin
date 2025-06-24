import React, { FC, useEffect } from 'react';
import { Form, Input, Modal } from 'antd';

export interface CreateAppModalProps {
  open: boolean;
  onCancel: () => void;
  onFinish: (values: { name: string }) => Promise<void>;
  confirmLoading?: boolean;
  initialValues?: { name: string };
  title?: string;
}

const CreateAppModal: FC<CreateAppModalProps> = ({ open, onCancel, onFinish, confirmLoading = false, initialValues, title = '创建应用' }) => {
  const [form] = Form.useForm();

  // 当初始值发生变化时更新表单
  useEffect(() => {
    if (initialValues && open) {
      form.setFieldsValue(initialValues);
    }
  }, [initialValues, form, open]);

  // 当弹窗关闭时重置表单
  useEffect(() => {
    if (!open) {
      form.resetFields();
    }
  }, [open, form]);

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
      title={title}
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
