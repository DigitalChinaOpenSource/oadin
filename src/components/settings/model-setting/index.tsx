import React, { useEffect } from 'react';
import { Form, Input, Tooltip, Button, Space } from 'antd';
import styles from './index.module.scss';
import ModelPathModal from '@/components/model-manage-tab/modelpath-modal';
import { useModelSetting } from '@/components/settings/model-setting/viem-model.ts';

// 表单数据类型定义
interface ModelSettingFormValues {
  modelDownloadUrl: string;
  modelSavePath: string;
}

const ModelSetting: React.FC = () => {
  // 创建表单实例并指定泛型类型
  const [form] = Form.useForm<ModelSettingFormValues>();
  const { modelPath, modalPathVisible, onModelPathVisible, onModalPathChangeSuccess } = useModelSetting();
  // 表单提交处理
  const onFinish = (values: ModelSettingFormValues) => {
    console.log('提交的表单数据:', values);
    // TODO: 调用API保存配置
  };
  console.log('modelPath', modelPath);

  // 更改目录按钮点击事件
  const handleChangeDir = () => {
    // TODO: 实现文件选择逻辑
    onModelPathVisible();
    form.setFieldsValue({ modelSavePath: modelPath });
  };

  useEffect(() => {
    console.log('formform', form.getFieldValue('modelDownloadUrl'));
    console.log('formform2', form.getFieldValue('modelSavePath'));
    form.setFieldsValue({ modelSavePath: modelPath });
  }, [modelPath]);

  return (
    <div className={styles.modelSetting}>
      <div className={styles.settingTitle}>模型设置</div>
      <div className={styles.settingContent}>
        <Form
          form={form}
          name="model-setting-form"
          initialValues={{
            modelDownloadUrl: '123',
            modelSavePath: modelPath,
          }}
          layout="vertical"
          onFinish={onFinish}
        >
          <Form.Item
            label="模型下载源地址"
            name="modelDownloadUrl"
            rules={[{ required: false, message: '请输入模型下载源地址' }]}
            tooltip={'模型下载源地址，所有模型将统一从该入口进行模型的下载'}
          >
            <Space>
              <Input
                placeholder="请输入模型下载源地址，所有模型将统一从该入口进行模型的下载"
                style={{ width: 500 }}
                value={form.getFieldValue('modelDownloadUrl')}
              />
            </Space>
          </Form.Item>

          <Form.Item
            label="模型存储路径"
            name="modelSavePath"
          >
            <Space>
              <Input
                readOnly
                style={{ width: 500 }}
                value={form.getFieldValue('modelSavePath') || modelPath}
              />
              <Button
                type="link"
                onClick={handleChangeDir}
              >
                更改目录
              </Button>
            </Space>
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
      </div>
      {modalPathVisible && (
        <ModelPathModal
          modalPath={modelPath}
          onModelPathVisible={onModelPathVisible}
          onModalPathChangeSuccess={onModalPathChangeSuccess}
        />
      )}
    </div>
  );
};

export default ModelSetting;
