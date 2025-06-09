import { useState, useEffect, memo } from 'react';
import { Modal, Input, Form, message } from 'antd';
import { useRequest, useDebounce } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import useModelDownloadStore from '@/store/useModelDownloadStore';
import useModelPathChangeStore from '@/store/useModelPathChangeStore';
import useByzeServerCheckStore from '@/store/useByzeServerCheckStore';
import { IModelPathSpaceRes } from '../types';
import { DOWNLOAD_STATUS } from '@/constants';
import styles from './index.module.scss';

interface IModelPathModalProps {
  modalPath?: string;
  onModelPathVisible: () => void;
  onModalPathChangeSuccess: () => void;
}

export default memo(function ModelPathModal(props: IModelPathModalProps) {
  const { modalPath, onModelPathVisible, onModalPathChangeSuccess } = props;

  const { downloadList } = useModelDownloadStore();
  console.log('downloadList', downloadList);
  const { IN_PROGRESS } = DOWNLOAD_STATUS;
  const { setMigratingStatus } = useModelPathChangeStore();
  const { checkByzeStatus, setCheckByzeServerLoading } = useByzeServerCheckStore();
  const [form] = Form.useForm();
  const formValues = Form.useWatch([], form);
  const modelPathValue = Form.useWatch('modelPath', form);
  const debouncedModelPath = useDebounce(modelPathValue, { wait: 1000 });
  const [currentPathSpace, setCurrentPathSpace] = useState<IModelPathSpaceRes>({} as IModelPathSpaceRes);
  const [isFormValid, setIsFormValid] = useState(false);
  const [changeModelPathLoading, setChangeModelPathLoading] = useState(false);
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
    if (!checkByzeStatus) {
      setChangeModelPathLoading(false);
      onModelPathVisible();
      setMigratingStatus('failed');
      setCheckByzeServerLoading(false);
    }
  }, [checkByzeStatus]);

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

  const { run: onChangeModelPath } = useRequest(
    async (params: { source_path: string; target_path: string }) => {
      const data = await httpRequest.post('/control_panel/model/filepath', params);
      return data || {};
    },
    {
      manual: true,
      onBefore: () => {
        setChangeModelPathLoading(true);
      },
      onSuccess: (data) => {
        if (data) {
          message.success('模型存储路径修改成功');
        }
        setCurrentPathSpace(data);
        onCheckPathSpace(formValues.modelPath);
        onModalPathChangeSuccess();
        setMigratingStatus('init');
      },
      onError: (error: Error & { handled?: boolean }) => {
        if (!error?.handled) {
          message.error(error?.message || '模型存储路径修改失败');
        }
        setMigratingStatus('failed');
      },
      onFinally: () => {
        setChangeModelPathLoading(false);
        setMigratingStatus('failed');
        onModelPathVisible();
      },
    },
  );

  const handleToSavePath = () => {
    form.submit();
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
    // 修改全局状态，标识模型存储路径正在迁移中
    setMigratingStatus('pending');
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
              <span>总容量 {currentPathSpace?.total_size}GB</span> ｜ 已使用 <span>{currentPathSpace?.usage_size}GB</span>，
              <span className={styles.diskCanUse}>剩余 {currentPathSpace?.free_size}GB</span>
            </div>
          )}
        </div>
      </div>
    </Modal>
  );
});
