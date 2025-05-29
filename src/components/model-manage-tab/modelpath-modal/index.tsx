import { useState, useEffect } from 'react';
import { Modal, Input, Form, message } from 'antd';
import { useRequest, useDebounce } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import styles from './index.module.scss';

interface IModelPathModalProps {
  modalPath?: string;
  onModalPathClose: () => void;
}

interface IModelPathSpaceRes {
  free_size: number;
  total_size: number;
  usage_size: number;
}

export default function ModelPathModal(props: IModelPathModalProps) {
  const [form] = Form.useForm();
  const formValues = Form.useWatch([], form);
  const modelPathValue = Form.useWatch('modelPath', form);
  const debouncedModelPath = useDebounce(modelPathValue, { wait: 1000 });
  const { modalPath, onModalPathClose } = props;
  const [currentPathSpace, setCurrentPathSpace] = useState<IModelPathSpaceRes>({} as IModelPathSpaceRes);
  const [isFormValid, setIsFormValid] = useState(false);

  useEffect(() => {
    if (!modalPath) return;
    form.setFieldsValue({ modelPath: modalPath });
  }, [modalPath]);

  useEffect(() => {
    if (!debouncedModelPath) {
      setCurrentPathSpace({} as IModelPathSpaceRes);
      return;
    }
    onCheckPathSpace(debouncedModelPath);
  }, [debouncedModelPath]);

  useEffect(() => {
    form
      .validateFields({ validateOnly: true })
      .then(() => setIsFormValid(true))
      .catch(() => setIsFormValid(false));
  }, [formValues, form]);

  const { run: onCheckPathSpace } = useRequest(
    async (path: string) => {
      const data = await httpRequest.get<IModelPathSpaceRes>('/control_panel/path/space', { path });
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        setCurrentPathSpace(data);
      },
      onError: (error) => {
        setCurrentPathSpace({} as IModelPathSpaceRes);
      },
    },
  );

  const { loading: changeModelPathLoading, run: onChangeModelPath } = useRequest(
    async (params: { source_path: string; target_path: string }) => {
      const data = await httpRequest.post('/control_panel/model/filepath', params);
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        if (data) {
          message.success('模型存储路径修改成功');
        }
        setCurrentPathSpace(data);
        onModalPathClose();
      },
      onError: (error) => {
        message.error(error?.message || '模型存储路径修改失败');
      },
    },
  );

  const handleToSavePath = () => {
    form.submit();
  };

  const onFinish = (values: { modelPath: string }) => {
    onChangeModelPath({
      source_path: modalPath || '',
      target_path: values.modelPath,
    });
  };

  return (
    <Modal
      centered
      title="修改模型存储路径"
      width={480}
      open
      okButtonProps={{
        disabled: !isFormValid,
        loading: changeModelPathLoading,
      }}
      onOk={handleToSavePath}
      onCancel={onModalPathClose}
      className={styles.modelPathModal}
      okText="确认"
    >
      <div className={styles.modelPathModal}>
        <div className={styles.tips}>
          <p>若本地模型正在工作中，该操作可能会造成业务的中断。</p>
          <p>在模型完成新路径迁移前,基于本地模型的所有功能将不可用。</p>
          <p className={styles.mark}>为保障业务使用，请先将具体应用的模型切换至云端模型。</p>
        </div>
        <div className={styles.modelPathInput}>
          <Form
            form={form}
            name="basic"
            labelCol={{ span: 6 }}
            wrapperCol={{ span: 18 }}
            initialValues={{ remember: true }}
            onFinish={onFinish}
          >
            <Form.Item
              label="模型存储路径："
              name="modelPath"
              rules={[{ required: true, message: '请输入模型存储路径' }]}
            >
              <Input
                className={styles.pathInput}
                allowClear
              />
            </Form.Item>
          </Form>

          {Object.keys(currentPathSpace).length > 0 && (
            <div className={styles.diskSpace}>
              <span>总容量 {currentPathSpace?.total_size}GB</span> ｜ 已使用 <span>{currentPathSpace?.usage_size}GB</span>，
              <span className={styles.diskCanUse}>剩余 {currentPathSpace?.free_size}GB</span>
            </div>
          )}
        </div>
      </div>
    </Modal>
  );
}
