import { useRef } from 'react';
import { Modal, Input, Form } from 'antd';
import styles from './index.module.scss';
import { IModelAuthorize, IModelAuthType, IModelAuth } from '../types';
import { ModelDataItem } from '@/types';

export interface IModelAuthorizeModalProps {
  // 模型数据
  modelDataItem: ModelDataItem;
  modelAuthType: IModelAuthType;
  modelAuthorize: IModelAuthorize;
  onModelAuthVisible: (data: IModelAuth) => void;
}

export default function ModelAuthorizeModal(props: IModelAuthorizeModalProps) {
  const [form] = Form.useForm();
  const { onModelAuthVisible, modelAuthType, modelDataItem } = props;

  // TODO 还有更新授权的情况
  const MODELTITLE = '模型授权';

  const handleSubmit = async () => {
    // TODO: 处理表单提交
    const result = await form.validateFields(); // 校验表单
    console.log('校验通过，提交数据:', result);
  };

  const handleCancel = () => {
    onModelAuthVisible?.({
      visible: false,
      type: modelAuthType,
      modelData: modelDataItem,
    });
  };

  const selectedCredentialParams = (modelDataItem?.credentialParams || []).filter((param) => (modelDataItem?.credentialParamsId || '').split(',').includes(String(param.id)));

  const renderCredentialParams = () => {
    console.log('modelDataItem', modelDataItem);
    console.log('selectedCredentialParams', selectedCredentialParams);
    return selectedCredentialParams.map((param) => {
      const formName = param.name;
      const label = param.label;
      const placeholder = param.placeholder;
      const isPassword = param.type === 'password';

      return (
        <Form.Item
          key={formName}
          label={label.charAt(0).toUpperCase() + label.slice(1)} // 首字母大写
          name={formName}
          rules={[{ required: Boolean(param.required), message: placeholder }]}
        >
          {isPassword ? <Input.Password allowClear /> : <Input allowClear />}
        </Form.Item>
      );
    });
  };

  return (
    <Modal
      centered
      open
      width={480}
      okText="确认"
      title={MODELTITLE}
      className={styles.ModelAuthorizeModal}
      onOk={handleSubmit}
      onCancel={handleCancel}
    >
      <Form
        name="basic"
        form={form}
        labelCol={{ span: 5 }}
        wrapperCol={{ span: 19 }}
        style={{ maxWidth: 600 }}
        initialValues={{ remember: true }}
        autoComplete="off"
      >
        {renderCredentialParams()}
      </Form>
    </Modal>
  );
}
