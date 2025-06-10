import React, { useEffect, useState } from 'react';
import { Form, Input, Tooltip, Button, Space } from 'antd';
import styles from './index.module.scss';
import ModelPathModal from '@/components/model-manage-tab/modelpath-modal';
import { useModelSetting } from '@/components/settings/model-setting/viem-model.ts';
import { IModelPathSpaceRes } from '@/components/model-manage-tab/types.ts';
import useModelPathChangeStore from '@/store/useModelPathChangeStore.ts';

// 表单数据类型定义
interface ModelSettingFormValues {
  modelDownloadUrl: string;
}

const ModelSetting: React.FC = () => {
  // 创建表单实例并指定泛型类型
  const [form] = Form.useForm<ModelSettingFormValues>();
  const {
    fetchModelPath,
    modelPath,
    modalPathVisible,
    onModelPathVisible,
    onChangeModelPath,
    onCheckPathSpace,
    currentPathSpace,
    setCurrentPathSpace,
    modelDownUrl,
    changeModelDownUrl,
    changingModelPath,
    setChangingModelPath,
  } = useModelSetting();

  const { migratingStatus } = useModelPathChangeStore();
  const isMigrating: boolean = migratingStatus === 'pending';

  // 表单提交处理
  const onFinish = (values: ModelSettingFormValues) => {
    // TODO: 调用API保存配置
    changeModelDownUrl(values.modelDownloadUrl);
  };

  // 更改目录按钮点击事件
  const handleChangeDir = () => {
    // TODO: 实现文件选择逻辑
    onModelPathVisible();
  };

  useEffect(() => {
    form.setFieldValue('modelDownloadUrl', modelDownUrl || '');
  }, [modelDownUrl]);

  useEffect(() => {
    if (!modelPath) {
      fetchModelPath();
      setCurrentPathSpace({} as IModelPathSpaceRes);
      return;
    }
    onCheckPathSpace(modelPath);
  }, [modelPath]);

  return (
    <div className={styles.modelSetting}>
      <div className={styles.settingTitle}>模型设置</div>
      <div className={styles.settingContent}>
        <Form
          form={form}
          name="model-setting-form"
          layout="vertical"
          onFinish={onFinish}
        >
          <Form.Item
            label="模型下载源地址"
            name="modelDownloadUrl"
            rules={[{ required: false, message: '请输入模型下载源地址' }]}
            tooltip={'模型下载源地址，所有模型将统一从该入口进行模型的下载'}
          >
            <Input
              placeholder="请输入模型下载源地址，所有模型将统一从该入口进行模型的下载"
              style={{ width: 500 }}
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
            >
              保存
            </Button>
          </Form.Item>
        </Form>
        <div className={styles.modelPath}>
          <div className={styles.pathLabel}>模型存储路径</div>
          <div className={styles.pathContent}>
            <div className={styles.mainLeft}>
              <div className={styles.path}>{modelPath}</div>
              <div className={styles.storageUse}>
                {Object.keys(currentPathSpace).length > 0 && (
                  <div className={styles.diskSpace}>
                    <span>( {currentPathSpace?.total_size}GB</span> / <span>{currentPathSpace?.usage_size}GB</span>，<span className={styles.diskCanUse}> {currentPathSpace?.free_size}GB可用)</span>
                  </div>
                )}
              </div>
            </div>
            <Tooltip title={isMigrating ? changingModelPath : ''}>
              <Button
                onClick={handleChangeDir}
                loading={isMigrating}
              >
                {isMigrating ? '正在修改至新的存储路径' : '更改'}
              </Button>
            </Tooltip>
          </div>
        </div>
      </div>
      {modalPathVisible && (
        <ModelPathModal
          modalPath={modelPath}
          onModelPathVisible={onModelPathVisible}
          updateModelPath={setChangingModelPath}
          onChangeModelPath={onChangeModelPath}
        />
      )}
      ``
    </div>
  );
};

export default ModelSetting;
