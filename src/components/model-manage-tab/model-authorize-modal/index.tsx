import { Modal, Input, Form, message } from 'antd';
import { useRequest } from 'ahooks';
import { IModelAuthorize, IModelAuthType, IModelAuth } from '../types';
import { IModelDataItem } from '@/types';
import { httpRequest } from '@/utils/httpRequest';
import styles from './index.module.scss';

export interface IModelAuthForm {
  auth_type: string;
  service_name: string;
  service_source: string;
  flavor_name: string;
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
  const ENV_TYPE = import.meta.env.VITE_ENV_TYPE;
  const [form] = Form.useForm();
  const { onModelAuthVisible, modelAuthType, modelDataItem } = props;

  console.log('modelDataItem===>', modelDataItem);
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
        console.log('模型配置授权成功', data, submitLoading);
        // if (!submitLoading) {

        // }
        handleCancel();
        message.success('模型配置授权成功');
      },
      onError: (error) => {
        message.error('模型配置授权失败，请重试或检查数据是否正确');
        console.error('模型配置授权失败:', error);
      },
    },
  );

  const submitForm = (result: any) => {
    console.log('submitForm result===>', result);
    // 组装提交数据
    const changeAuthKey = {
      [modelDataItem.name]: {
        credentials: result,
        env_type: ENV_TYPE || 'product',
        provider: modelDataItem.provider,
        model_key: modelDataItem?.modelKey,
      },
    };
    const changeParams = {
      auth_type: 'credentials',
      service_name: 'chat',
      service_source: 'remote',
      flavor_name: 'smartvision',
      provider_name: 'remote_smartvision_chat',
      models: [modelDataItem.name],
      auth_key: JSON.stringify(changeAuthKey),
    };
    console.log('changeParams', changeParams);

    // 调用接口
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

  const selectedCredentialParams = (modelDataItem?.credentialParams || []).filter((param) => (modelDataItem?.credentialParamsId || '').split(',').includes(String(param.id)));

  const renderCredentialParams = () => {
    return selectedCredentialParams.map((param) => {
      const formName = param.name;
      const label = param.label;
      const placeholder = param.placeholder;
      const isPassword = param.type === 'password';

      return (
        <Form.Item
          key={formName}
          label={label.charAt(0).toUpperCase() + label.slice(1)}
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
      okButtonProps={{ loading: submitLoading }}
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
