import { Modal, Input, Form } from 'antd';
import styles from './index.module.scss';
import { IModelAuthorize, IModelAuthType } from '../types';

interface IModelAuthorizeModalProps {
  modelAuthType: IModelAuthType;
  modelAuthorize: IModelAuthorize;
  onSetModelAuthorize?: (authData: IModelAuthorize) => void;
  onModelAuthVisible?: (visible: boolean, type: IModelAuthType) => void;
}

export default function ModelAuthorizeModal(props: IModelAuthorizeModalProps) {
  const { onModelAuthVisible, onSetModelAuthorize, modelAuthType } = props;
  const modalTitle = '模型授权';

  const onFinish = (values: IModelAuthorize) => {
    if (onSetModelAuthorize) {
      onSetModelAuthorize(values);
    }
  };
  return (
    <Modal
      centered
      open
      width={480}
      okText="确认"
      title={modalTitle}
      className={styles.ModelAuthorizeModal}
      onOk={() => onModelAuthVisible?.(false, modelAuthType)}
      onCancel={() => onModelAuthVisible?.(false, modelAuthType)}
    >
      <Form
        name="basic"
        labelCol={{ span: 5 }}
        wrapperCol={{ span: 19 }}
        style={{ maxWidth: 600 }}
        initialValues={{ remember: true }}
        onFinish={onFinish}
        autoComplete="off"
      >
        <Form.Item<IModelAuthorize>
          label="API-Host"
          name="apiHost"
          rules={[{ required: true, message: '请输入API Host' }]}
        >
          <Input allowClear />
        </Form.Item>

        <Form.Item<IModelAuthorize>
          label="API-Key"
          name="apiKey"
          rules={[{ required: true, message: '请输入API Key' }]}
        >
          <Input.Password allowClear />
        </Form.Item>
      </Form>
    </Modal>
  );
}
