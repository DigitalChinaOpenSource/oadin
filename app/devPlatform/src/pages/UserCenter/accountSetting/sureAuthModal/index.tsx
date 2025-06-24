import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, Button, message } from 'antd';
import { CloseOutlined } from '@ant-design/icons';
import styles from './index.module.scss';
import PhoneNumberInput from '@/pages/Login/components/phoneNumberInput';
import CodeInput from '@/pages/Login/components/codeInput';
import EmailInput from '@/pages/Login/components/emailInput';

interface ISureAuthModalProps {
  visible: boolean;
  onCancel: () => void;
  onConfirm: () => void;
  userType: 'person' | 'enterprise';
}

/**
 * 身份验证弹窗组件
 */
const SureAuthModal: React.FC<ISureAuthModalProps> = ({
  visible,
  onCancel,
  onConfirm,
  userType = 'person', // 默认为个人用户
}) => {
  const [form] = Form.useForm();

  // 重置表单和倒计时
  useEffect(() => {
    if (!visible) {
      form.resetFields();
    }
  }, [visible, form]);

  // 提交表单
  const handleSubmit = async () => {
    await form.validateFields();

    // 提交验证请求
    // const formData = form.getFieldsValue();
    // const res = await verifyIdentity(formData);

    message.success('身份验证成功');
    onConfirm();
  };

  // 个人用户表单
  const renderPersonForm = () => (
    <Form
      form={form}
      layout="vertical"
    >
      <PhoneNumberInput
        form={form}
        codeFiled={'phoneNumber'}
      />

      <CodeInput
        form={form}
        validateFiled={'phoneNumber'}
        codeFiled={'code'}
      />
    </Form>
  );

  // 企业用户表单
  const renderEnterpriseForm = () => (
    <Form
      form={form}
      layout="vertical"
    >
      <Form.Item
        name="companyName"
        rules={[{ required: true, message: '请输入企业名称' }]}
      >
        <Input
          className="formInput"
          placeholder="请输入企业名称"
        />
      </Form.Item>

      <EmailInput
        form={form}
        codeFiled="email"
      />
      <CodeInput
        form={form}
        validateFiled={'email'}
        codeFiled={'emailCode'}
      />

      <PhoneNumberInput
        form={form}
        codeFiled={'phoneNumber'}
      />

      <CodeInput
        form={form}
        validateFiled={'phoneNumber'}
        codeFiled={'phoneCode'}
      />
    </Form>
  );

  return (
    <Modal
      title={
        <div className={styles.headerTitle}>
          <span>身份验证</span>
        </div>
      }
      open={visible}
      onCancel={onCancel}
      footer={null}
      closeIcon={<CloseOutlined />}
      centered
      className={styles.authModal}
      width={480}
      maskClosable={false}
    >
      <div className={styles.description}>为确保账号安全，注销前需进行身份验证</div>

      <div className={styles.formContent}>{userType === 'person' ? renderPersonForm() : renderEnterpriseForm()}</div>

      <div className={styles.buttonGroup}>
        <Button onClick={onCancel}>取消</Button>
        <Button
          type="primary"
          onClick={handleSubmit}
          danger
        >
          确认注销
        </Button>
      </div>
    </Modal>
  );
};

export default SureAuthModal;
