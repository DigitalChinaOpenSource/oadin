import React, { useEffect } from 'react';
import { Modal, Form, Checkbox, Typography, Space } from 'antd';
import styles from './ModelSelectModal.module.scss';

export interface ModelSelectModalProps {
  open: boolean;
  onCancel: () => void;
  onFinish: (selectedModels: string[]) => void;
  title: string;
  confirmLoading?: boolean;
  modelOptions: { value: string; label: string }[];
  initialSelectedModels?: string[];
}

const ModelSelectModal: React.FC<ModelSelectModalProps> = ({
  open,
  onCancel,
  onFinish,
  title,
  confirmLoading = false,
  modelOptions,
  initialSelectedModels = []
}) => {
  const [form] = Form.useForm();
  const [selectedModels, setSelectedModels] = React.useState<string[]>(initialSelectedModels);

  // 当初始值或弹窗打开状态变化时更新选中的模型
  useEffect(() => {
    if (open) {
      setSelectedModels(initialSelectedModels);
      form.setFieldsValue({ models: initialSelectedModels });
    }
  }, [initialSelectedModels, open, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      onFinish(values.models || []);
    } catch (error) {
      console.error('表单验证失败:', error);
    }
  };

  const handleCheckboxChange = (checkedValues: any) => {
    setSelectedModels(checkedValues);
  };

  return (
    <Modal
      title={title}
      open={open}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={confirmLoading}
      width={600}
    >
      <div className={styles.modalContent}>
        <Form form={form} layout="vertical">
          <Form.Item
            name="models"
            rules={[{ required: true, message: '请至少选择一个模型' }]}
          >
            <Checkbox.Group 
              onChange={handleCheckboxChange} 
              className={styles.checkboxGroup}
            >
              <Space direction="vertical" className={styles.modelList}>
                {modelOptions.map((model) => (
                  <Checkbox key={model.value} value={model.value} className={styles.modelOption}>
                    <Typography.Text>{model.label}</Typography.Text>
                  </Checkbox>
                ))}
              </Space>
            </Checkbox.Group>
          </Form.Item>
        </Form>

        {selectedModels.length > 0 && (
          <div className={styles.selectedCount}>
            <Typography.Text type="secondary">
              已选择 {selectedModels.length} 个项目
            </Typography.Text>
          </div>
        )}
      </div>
    </Modal>
  );
};

export default ModelSelectModal;
