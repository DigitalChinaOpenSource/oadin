'use client';

import React, { useEffect } from 'react';
import { Button, Modal, Form, Input, Select, Pagination } from 'antd';
import styles from './index.module.scss';
import { useViewModel } from './view-model';

interface ServiceProviderEditProps {
  id?: string;
  visible: boolean;
  onCancel: () => void;
}

export default function ServiceProviderEdit({ id = '', visible, onCancel }: ServiceProviderEditProps) {
  const [form] = Form.useForm();
  const vm = useViewModel();
  const { modelList, serviceOrgList } = vm;

  // 提交处理
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();

      onCancel();
    } catch (error) {
      console.error('表单验证失败:', error);
    }
  };

  // 初始化表单值
  useEffect(() => {
    // if (visible && id) {
    // } else {
    //   form.resetFields();
    // }
  }, [visible, id]);

  return (
    <Modal
      centered
      open={visible}
      width={860}
      title={<div className={styles.modalTitle}>{id ? '编辑服务供应商' : '新增服务供应商'}</div>}
      onCancel={onCancel}
      footer={[
        <Button key="cancel" onClick={onCancel}>
          取消
        </Button>,
        <Button key="submit" type="primary" onClick={handleSubmit}>
          确认
        </Button>,
      ]}
    >
      <Form form={form} labelCol={{ span: 5 }} wrapperCol={{ span: 19 }}>
        {/* 基础信息 */}
        <div className={styles.infoName}>基础信息</div>
        <Form.Item label="服务提供商名称" name="providerName" rules={[{ required: true, message: '请输入服务提供商名称' }]}>
          <Input placeholder="请输入服务提供名称" />
        </Form.Item>

        <Form.Item label="服务提供商厂商名称" name="purchaseType" rules={[{ required: true, message: '服务提供商厂商名称' }]}>
          <Input placeholder="招商权" />
        </Form.Item>

        <Form.Item label="服务来源" name="serviceName" rules={[{ required: true, message: '请选择服务来源' }]}>
          <Select placeholder="请选择服务名称">
            {serviceOrgList.map((item) => (
              <Select.Option key={item?.key} value={item.name}>
                {item.name}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item label="服务名称" name="serviceName" rules={[{ required: true, message: '请选择服务名称' }]}>
          <Select placeholder="请选择服务名称">
            {modelList.map((item) => (
              <Select.Option key={item?.key} value={item.name}>
                {item.name}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item label="描述" name="description">
          <Input.TextArea placeholder="请输入" />
        </Form.Item>

        {/* 鉴权信息 */}
        <div className={styles.infoName}>鉴权信息</div>
        <Form.Item label="请求方法" name="requestMethod" rules={[{ required: true, message: '请选择请求方法' }]}>
          <Select placeholder="请选择">
            <Select.Option value="GET">GET</Select.Option>
            <Select.Option value="POST">POST</Select.Option>
            <Select.Option value="PUT">PUT</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item label="URL" name="endpoint" rules={[{ required: true, message: '请输入服务提供商的URL', type: 'url' }]}>
          <Input placeholder="请输入服务提供商的URL" />
        </Form.Item>

        <Form.Item label="鉴权类型" name="authType" rules={[{ required: true, message: '请输入服务提供商的鉴权类型' }]}>
          <Select placeholder="请输入服务提供商的鉴权类型">
            <Select.Option value="API_KEY">API Key</Select.Option>
            <Select.Option value="OAUTH2">OAuth 2.0</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item label="鉴权Key" name="apiKey" rules={[{ required: true, message: '请输入服务提供商的鉴权信息' }]}>
          <Input placeholder="请输入服务提供商的鉴权信息" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
