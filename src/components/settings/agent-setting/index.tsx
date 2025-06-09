import React, { useState } from 'react';
import { Form, Input, Tooltip, Button, Space, Switch } from 'antd';
import styles from './index.module.scss';

// 表单数据类型定义
interface AgentSettingFormValues {
  agentIp: string;
  agentPort: string;
}

const AgentSetting: React.FC = () => {
  // 创建表单实例并指定泛型类型
  const [form] = Form.useForm<AgentSettingFormValues>();
  const [agentChecked, setAgentChecked] = useState(false);

  // 表单提交处理
  const onFinish = (values: AgentSettingFormValues) => {
    console.log('提交的表单数据:', values);
    // TODO: 调用API保存配置
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
                setAgentChecked(checked);
              }}
            />
            <span>{agentChecked ? '开' : '关'}</span>
          </div>
        </div>
        {agentChecked && (
          <Form
            form={form}
            name="agent-setting-form"
            initialValues={{
              agentIp: '',
              agentPort: '',
            }}
            layout="vertical"
            onFinish={onFinish}
          >
            <Form.Item
              label="代理IP或域名地址"
              name="agentIp"
              rules={[{ required: false, message: '请输入代理IP或域名地址' }]}
            >
              <Space>
                <Input
                  placeholder="请输入代理IP或域名地址"
                  style={{ width: 500 }}
                />
              </Space>
            </Form.Item>

            <Form.Item
              label="端口"
              name="agentPort"
            >
              <Space>
                <Input
                  placeholder="请输入端口"
                  style={{ width: 500 }}
                />
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
        )}
      </div>
    </div>
  );
};

export default AgentSetting;
