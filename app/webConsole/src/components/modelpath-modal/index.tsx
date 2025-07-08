import { useState, useEffect, memo } from 'react';
import { Modal, Input, Form, message } from 'antd';
import { useRequest, useDebounce } from 'ahooks';
import { httpRequest, healthRequest } from '@/utils/httpRequest';
import useModelDownloadStore from '@/store/useModelDownloadStore';
import { useModelPathChangeStore } from '@/store/useModelPathChangeStore';
import useOadinServerCheckStore from '@/store/useOadinServerCheckStore';
import useChatStore from '../chat-container/store/useChatStore';
import { IModelPathSpaceRes } from '../model-manage-tab/types';
import { DOWNLOAD_STATUS } from '@/constants';
import styles from './index.module.scss';

interface IModelPathModalProps {
  modalPath?: string;
  onModelPathVisible: () => void;
  updateModelPath: (path: string) => void;
  // 在外部调接口，同时关闭弹窗，接口状态由外部控制
  onChangeModelPath: (params: { source_path: string; target_path: string }) => void;
}

export default memo(function ModelPathModal(props: IModelPathModalProps) {
  const { modalPath, onModelPathVisible, onChangeModelPath, updateModelPath } = props;

  const downloadList = useModelDownloadStore((state) => state.downloadList);
  const { IN_PROGRESS } = DOWNLOAD_STATUS;
  const { setMigratingStatus } = useModelPathChangeStore();
  const { checkOadinStatus, setCheckOadinServerLoading } = useOadinServerCheckStore();
  const isChatLoading = useChatStore.getState().isLoading;
  const [form] = Form.useForm();
  const formValues = Form.useWatch([], form);
  const modelPathValue = Form.useWatch('modelPath', form);
  const debouncedModelPath = useDebounce(modelPathValue, { wait: 1000 });
  const [currentPathSpace, setCurrentPathSpace] = useState<IModelPathSpaceRes>({} as IModelPathSpaceRes);
  const [isFormValid, setIsFormValid] = useState(false);
  const [oadinHealth, setOadinHealth] = useState<boolean | null>(null);

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

  useEffect(() => {
    if (!checkOadinStatus) {
      onModelPathVisible();
      setMigratingStatus('failed');
      setCheckOadinServerLoading(false);
    } else {
      setMigratingStatus('init');
    }
  }, [checkOadinStatus]);

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

  const checkOadinHealth = async () => {
    const data = await healthRequest.get('/health');
    try {
      if (data?.status === 'UP') {
        return true;
      } else {
        return false;
      }
    } catch (error) {
      setOadinHealth(false);
      message.error('检查 Oadin 服务状态失败，请稍后重试');
    }
  };
  const handleToSavePath = () => {
    checkOadinHealth().then(() => {
      if (oadinHealth === false) {
        message.error('Oadin 服务不可用，请检查 Oadin 服务状态');
        return;
      }
      form.submit();
    });
  };

  const onFinish = (values: { modelPath: string }) => {
    // 如果当前输入的路径与传入的modalPath相同，则不调接口，直接关闭弹窗
    if (values.modelPath === modalPath) {
      message.info('路径未发生变化');
      onModelPathVisible();
      return;
    }
    // 检查是否有正在下载中的模型
    const hasDownloadingModel = downloadList.some((item) => item.status === IN_PROGRESS);
    if (hasDownloadingModel) {
      message.error('请等待模型下载完成后再进行操作');
      return;
    }
    updateModelPath(values.modelPath);
    onChangeModelPath({
      source_path: modalPath || '',
      target_path: values.modelPath,
    });
    onModelPathVisible();
  };

  return (
    <Modal
      centered
      title="修改模型存储路径"
      width={480}
      open
      okButtonProps={{
        disabled: !isFormValid || isChatLoading,
      }}
      onOk={handleToSavePath}
      onCancel={onModelPathVisible}
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
              已使用 <span>{currentPathSpace?.usage_size}GB</span> ｜ <span>总容量 {currentPathSpace?.total_size}GB</span> ，
              <span className={styles.diskCanUse}>{currentPathSpace?.free_size}GB 可用</span>
            </div>
          )}
        </div>
      </div>
    </Modal>
  );
});
