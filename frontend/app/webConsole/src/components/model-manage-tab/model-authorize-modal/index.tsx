import { Modal, Input, Form, message } from 'antd';
import { useRequest } from 'ahooks';
import { IModelAuthType, IModelAuth } from '../types';
import { IModelDataItem } from '@/types';
import { httpRequest } from '@/utils/httpRequest';
import { toHttpHeaderFormat } from '@/utils';
import styles from './index.module.scss';

export interface IModelAuthForm {
  auth_type: string;
  service_name: string;
  service_source: string;
  flavor_name?: string;
  api_flavor?: string;
  provider_name: string;
  models: string[];
  auth_key: string;
}

export interface IModelAuthorizeModalProps {
  // 模型数据
  modelDataItem: IModelDataItem;
  modelAuthType: IModelAuthType;
  onModelAuthVisible: (data: IModelAuth) => void;
  onModelAuthSuccess: () => void;
}

export default function ModelAuthorizeModal(props: IModelAuthorizeModalProps) {
  const [form] = Form.useForm();
  const { onModelAuthVisible, modelAuthType, modelDataItem, onModelAuthSuccess } = props;

  // TODO 还有更新授权的情况
  const MODELTITLE = '配置授权';

  const handleSubmit = async () => {
    // TODO: 处理表单提交
    const result = await form.validateFields(); // 校验表单
    submitForm(result);
    console.log('校验通过，提交数据:', result);
  };

  // 配置模型授权
  const { loading: submitLoading, run: fetchSubmitAuth } = useRequest(
    async (params: IModelAuthForm) => {
      const data = await httpRequest.put<IModelAuthForm>('/service_provider', params);
      return data;
    },
    {
      manual: true,
      onSuccess: (data) => {
        handleCancel();
        onModelAuthSuccess();
        message.success('模型配置授权成功');
      },
      onError: (error: Error & { handled?: boolean }) => {
        console.error('模型配置授权失败:', error);
        if (!error?.handled) {
          message.error('模型配置授权失败，请重试或检查数据是否正确');
        }
      },
    },
  );

  const submitForm = (result: any) => {
    // 组装提交数据
    const changeAuthKey = {
      [modelDataItem.name]: {
        credentials: result,
        env_type: 'product',
        provider: modelDataItem.smartvision_provider,
        model_key: modelDataItem?.smartvision_model_key,
      },
    };
    const changeParams = {
      auth_type: 'credentials',
      service_name: 'chat',
      service_source: 'remote',
      // flavor_name: 'smartvision',
      provider_name: 'remote_smartvision_chat',
      api_flavor: modelDataItem?.api_flavor,
      models: [modelDataItem.name],
      auth_key: JSON.stringify(changeAuthKey),
    };

    fetchSubmitAuth(changeParams);
  };

  // 关闭弹窗
  const handleCancel = () => {
    onModelAuthVisible?.({
      visible: false,
      type: modelAuthType,
      modelData: modelDataItem,
    });
  };

  const renderCredentialParams = () => {
    return (modelDataItem?.auth_fields || []).map((field) => {
      return (
        <Form.Item
          key={field}
          label={toHttpHeaderFormat(field)}
          name={field}
          rules={[{ required: true, message: `请输入${toHttpHeaderFormat(field)}` }]}
        >
          {field === 'api_key' ? <Input.Password allowClear /> : <Input allowClear />}
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
      okButtonProps={{ loading: submitLoading }}
    >
      <Form
        name="basic"
        form={form}
        labelCol={{ span: 6 }}
        wrapperCol={{ span: 18 }}
        style={{ maxWidth: 600 }}
        initialValues={{ remember: true }}
        autoComplete="off"
      >
        {renderCredentialParams()}
      </Form>
    </Modal>
  );
}
