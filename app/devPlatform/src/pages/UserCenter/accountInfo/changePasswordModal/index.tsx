import { CloseOutlined } from '@ant-design/icons';
import { Form, Input, Modal } from 'antd';
import React from 'react';
import styles from './index.module.scss';

interface IChangePasswordProps {
  visible: boolean;
  onCancel: () => void;
  onConfirm: () => void;
  title?: string;
}

const ChangePasswordModal = ({ title, onCancel, onConfirm, visible }: IChangePasswordProps) => {
  const [form] = Form.useForm();
  const handleOk = async () => {
    await form.validateFields();
    onConfirm();
  };

  return (
    <Modal
      title={
        <div className={styles.headerTitle}>
          <span>{title || '设置登录密码'}</span>
        </div>
      }
      open={visible}
      onCancel={onCancel}
      closeIcon={<CloseOutlined />}
      centered
      width={480}
      maskClosable={true}
      destroyOnHidden={true}
      onOk={handleOk}
    >
      <div className={styles.description}>密码必须是8-16位英文字母、数字、字符组合(不能是纯数字)</div>
      <div className={styles.formContent}>
        <Form
          form={form}
          name="userNameForm"
          autoComplete="off"
          noValidate={true}
          labelCol={{ span: 5 }}
        >
          <Form.Item
            name="oldPassword"
            label="旧密码"
            rules={[{ required: true, message: '请输入原始密码' }]}
          >
            <Input
              placeholder={'请输入原始密码'}
              allowClear
            />
          </Form.Item>
          <Form.Item
            name={'password'}
            label="新密码"
            validateTrigger={['onBlur']}
            rules={[
              { required: true, message: '请输入新的登录密码' },
              {
                pattern: /^(?![0-9]+$)(?![a-zA-Z]+$)[0-9A-Za-z!@#$%^&*()_+]{8,16}$/,
                message: '密码必须是8-16位英文字母、数字、字符组合',
              },
            ]}
          >
            <Input.Password
              placeholder={'请设置新的登录密码'}
              minLength={8}
              maxLength={16}
              allowClear
            />
          </Form.Item>
          <Form.Item
            name={'surePassword'}
            label="再次输入"
            dependencies={['password']}
            validateTrigger={['onBlur', 'onChange']}
            rules={[
              { required: true, message: '请确认新的登录密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password
              placeholder={'请再次输入登录密码'}
              minLength={8}
              maxLength={16}
              allowClear
            />
          </Form.Item>
        </Form>
      </div>
    </Modal>
  );
};

export default ChangePasswordModal;
