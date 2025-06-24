import React, { useEffect } from 'react';
import { Checkbox, Form, Modal, Space, Typography } from 'antd';
import styles from './ModelSelectModal.module.scss';
import { IModelDataItem } from '@/types/model.ts';
import { transformedIds2Card } from '@/pages/AppManagement/AppConfig/uitls.ts';
import { IModelSelectCardItem } from '@/pages/AppManagement/AppConfig/types.ts';

// 模型选择弹窗
export interface ModelSelectModalProps {
  open: boolean;
  onCancel: () => void;
  onFinish: (selectedModels: IModelSelectCardItem[]) => void;
  title: string;
  confirmLoading?: boolean;
  modelOptions: IModelDataItem[];
  initialSelectedModels?: string[];
}

const ModelSelectModal: React.FC<ModelSelectModalProps> = ({ open, onCancel, onFinish, title, confirmLoading = false, modelOptions, initialSelectedModels = [] }) => {
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
      const cardList = transformedIds2Card(modelOptions, values.models);
      onFinish(cardList || []);
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
        <Form
          form={form}
          layout="vertical"
        >
          <Form.Item
            name="models"
            rules={[{ required: true, message: '请至少选择一个模型' }]}
          >
            <Checkbox.Group
              onChange={handleCheckboxChange}
              className={styles.checkboxGroup}
            >
              <Space
                direction="vertical"
                className={styles.modelList}
              >
                {modelOptions.map((model) => (
                  <Checkbox
                    key={model.id}
                    value={model.id}
                    className={styles.modelOption}
                  >
                    <Typography.Text>{model.name}</Typography.Text>
                  </Checkbox>
                ))}
              </Space>
            </Checkbox.Group>
          </Form.Item>
        </Form>

        {selectedModels.length > 0 && (
          <div className={styles.selectedCount}>
            <Typography.Text type="secondary">已选择 {selectedModels.length} 个项目</Typography.Text>
          </div>
        )}
      </div>
    </Modal>
  );
};

export default ModelSelectModal;
