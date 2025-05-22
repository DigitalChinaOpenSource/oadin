import { useState, useEffect } from 'react';
import { Modal, Input, Form } from 'antd';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import styles from './index.module.scss';

interface IModelPathModalProps {
  modalPath?: string;
  onModalPathClose: () => void;
}

interface IModelPathSpaceRes {
  data: {
    free_size: number;
    total_size: number;
    usage_size: number;
  };
}

export default function ModelPathModal(props: IModelPathModalProps) {
  const [form] = Form.useForm();
  const formValues = Form.useWatch([], form);

  const { modalPath, onModalPathClose } = props;
  const [currentPathSpace, setCurrentPathSpace] = useState<IModelPathSpaceRes['data']>({} as IModelPathSpaceRes['data']);
  const [isFormValid, setIsFormValid] = useState(false);

  useEffect(() => {
    if (!modalPath) return;
    form.setFieldsValue({ modelPath: modalPath });
  }, [modalPath]);

  useEffect(() => {
    form
      .validateFields({ validateOnly: true })
      .then(() => setIsFormValid(true))
      .catch(() => setIsFormValid(false));
  }, [formValues, form]);

  const { run: onChechPathSpace } = useRequest(
    async (path: string) => {
      const res = await httpRequest.get<IModelPathSpaceRes>('/control_panel/path/space', { path });
      return res?.data || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        setCurrentPathSpace(data);
      },
    },
  );

  const handleToSavePath = () => {
    form.submit();
  };

  const onFinish = (values: { modelPath: string }) => {
    console.log('Success:', values);
    // TODO 提交接口
    onModalPathClose();
  };

  return (
    <Modal
      centered
      title="修改模型存储路径"
      width={480}
      open
      okButtonProps={{
        disabled: !isFormValid,
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
              <span>{currentPathSpace?.usage_size}GB</span> / <span>{currentPathSpace?.total_size}GB</span>，<span className={styles.diskCanUse}>{currentPathSpace?.free_size}GB 可用</span>
            </div>
          )}
        </div>
      </div>
    </Modal>
  );
}
