import React, { useEffect, useState } from 'react';
import { Form, Input, Tooltip, Button, Space, Switch } from 'antd';
import styles from './index.module.scss';
import { useAgentSettingViewModel } from './view-module';

// 表单数据类型定义
export interface AgentSettingFormValues {
  endpoint: string;
  username?: string;
  password?: string;
}

const AgentSetting: React.FC = () => {
  // 创建表单实例并指定泛型类型
  const [form] = Form.useForm<AgentSettingFormValues>();
  const { agentChecked, systemProxy, changeProxy, saveProxy, saveProxyLoading } = useAgentSettingViewModel();

  // 监听systemProxy变化，回显到表单
  useEffect(() => {
    form.setFieldsValue(systemProxy);
  }, [systemProxy, form]);

  // 表单提交处理
  const onFinish = (values: AgentSettingFormValues) => {
    console.log('提交的表单数据:', values);
    saveProxy(values);
  };

  return (
    <div className={styles.agentSetting}>
      <div className={styles.settingTitle}>代理设置</div>
      <div className={styles.settingContent}>
        <div className={styles.contentHeader}>
          <div className={styles.headerTitle}>代理服务器</div>
          <div className={styles.headerControl}>
            <Switch
              checkedChildren="ON"
              unCheckedChildren="OFF"
              checked={agentChecked}
              onClick={(checked, e) => {
                e.stopPropagation();
                changeProxy(checked);
              }}
            />
            <span>{agentChecked ? '开' : '关'}</span>
          </div>
        </div>
        <Form
          form={form}
          name="agent-setting-form"
          initialValues={systemProxy}
          layout="vertical"
          onFinish={onFinish}
        >
          <Form.Item
            label="代理地址"
            name="endpoint"
            rules={[{ required: true, message: '请输入代理地址' }]}
            tooltip={'请输入代理地址'}
          >
            <Input
              autoComplete={'off'}
              placeholder="请输入代理地址"
              style={{ width: 400 }}
              allowClear={true}
            />
          </Form.Item>
          <Form.Item>
            <Space size={40}>
              <Form.Item
                label="用户名"
                name="username"
                rules={[{ required: false, message: '请输入用户名' }]}
                style={{ marginBottom: 0 }}
              >
                <Input
                  autoComplete={'off'}
                  placeholder="请输入用户名"
                  style={{ width: 400 }}
                  allowClear={true}
                />
              </Form.Item>
              <Form.Item
                label="密码"
                name="password"
                rules={[{ required: false, message: '请输入密码' }]}
                style={{ marginBottom: 0 }}
              >
                <Input.Password
                  autoComplete={'off'}
                  placeholder="请输入密码"
                  style={{ width: 400 }}
                  allowClear={true}
                />
              </Form.Item>
            </Space>
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={saveProxyLoading}
            >
              保存
            </Button>
          </Form.Item>
        </Form>
      </div>
    </div>
  );
};

export default AgentSetting;
