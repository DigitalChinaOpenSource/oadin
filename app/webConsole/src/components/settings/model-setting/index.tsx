import React, { useEffect } from 'react';
import { Button, Form, Input, Tooltip } from 'antd';
import styles from './index.module.scss';
import ModelPathModal from '@/components/modelpath-modal';
import { useModelSetting } from '@/components/settings/model-setting/viem-model.ts';
import realLoadingSvg from '@/components/icons/real-loading.svg';
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
    modelDownUrl,
    changeModelDownUrl,
    changeModelDownUrlLoading,
    changingModelPath,
    setChangingModelPath,
    isCheckingPathSpace,
  } = useModelSetting();

  const { migratingStatus } = useModelPathChangeStore();
  const isMigrating: boolean = migratingStatus === 'pending';

  // 表单提交处理
  const onFinish = (values: ModelSettingFormValues) => {
    changeModelDownUrl(values.modelDownloadUrl);
  };

  // 更改目录按钮点击事件
  const handleChangeDir = async () => {
    // 校验源路径是否可用
    onCheckPathSpace(modelPath)
      .then((res) => {
        onModelPathVisible();
      })
      .catch((err) => {
        console.log('校验源路径失败', err);
      });
  };

  useEffect(() => {
    form.setFieldValue('modelDownloadUrl', modelDownUrl || '');
  }, [modelDownUrl]);

  useEffect(() => {
    fetchModelPath();
  }, []);

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
          {/*<Form.Item*/}
          {/*  style={{ marginBottom: 0, height: '36px' }}*/}
          {/*  label="模型下载源地址"*/}
          {/*  tooltip={'模型下载源地址，所有模型将统一从该入口进行模型的下载'}*/}
          {/*></Form.Item>*/}
          <Form.Item
            label="模型下载源地址"
            name="modelDownloadUrl"
            tooltip={'模型下载源地址，所有模型将统一从该入口进行模型的下载'}
            rules={[{ required: false, message: '请输入模型下载源地址' }]}
          >
            <Input
              placeholder="请输入模型下载源地址"
              style={{ width: 500 }}
              allowClear={true}
              autoComplete={'off'}
            />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={changeModelDownUrlLoading}
            >
              保存
            </Button>
          </Form.Item>
        </Form>
        <div className={styles.modelPath}>
          <div className={styles.pathLabel}>模型存储路径</div>
          <div className={styles.pathContent}>
            <div className={styles.mainLeft}>
              <>
                {isCheckingPathSpace ? (
                  <img
                    style={{ width: 20, height: 20 }}
                    src={realLoadingSvg}
                    alt="loading"
                  />
                ) : (
                  <>
                    <div className={styles.path}>{modelPath}</div>
                    <div className={styles.storageUse}>
                      {Object.keys(currentPathSpace).length > 0 ? (
                        <div className={styles.diskSpace}>
                          <span>( {currentPathSpace?.usage_size}GB</span> / <span>{currentPathSpace?.total_size}GB</span>，
                          <span className={styles.diskCanUse}> {currentPathSpace?.free_size}GB可用)</span>
                        </div>
                      ) : (
                        '暂无数据'
                      )}
                    </div>
                  </>
                )}
              </>
            </div>
            <Tooltip title={isMigrating ? changingModelPath : ''}>
              <Button
                onClick={handleChangeDir}
                disabled={isCheckingPathSpace}
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
    </div>
  );
};

export default ModelSetting;
