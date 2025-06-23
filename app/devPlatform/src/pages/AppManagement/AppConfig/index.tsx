import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button, Card, Form, Input, message, Space, Spin, Tabs } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import styles from './index.module.scss';

interface AppConfigProps {}

// 模拟获取应用详情的接口
const fetchAppDetail = async (id: string) => {
  console.log('Fetching app details for:', id);
  // 模拟API调用延迟
  await new Promise((resolve) => setTimeout(resolve, 1000));
  
  // 模拟返回数据
  return {
    id,
    name: `应用${id}`,
    appId: `app_${id.slice(0, 8)}`,
    secretKey: `sk_${id.slice(0, 12)}`,
    modelCount: 3,
    mcpCount: 2,
    osCount: 1,
    updatedAt: new Date().toISOString(),
    description: '这是应用的详细描述信息',
    configs: {
      modelSettings: {
        defaultModel: 'gpt-4',
        temperature: 0.7,
        maxTokens: 2048
      },
      apiSettings: {
        rateLimit: 100,
        timeout: 30
      }
    }
  };
};

// 模拟更新应用的接口
const updateApp = async (id: string, data: any) => {
  console.log('Updating app:', id, data);
  await new Promise((resolve) => setTimeout(resolve, 1000));
  return { success: true };
};

const AppConfig: React.FC<AppConfigProps> = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [appDetail, setAppDetail] = useState<any>(null);
  const [form] = Form.useForm();
  
  // 获取应用详情
  useEffect(() => {
    if (!id) {
      message.error('应用ID不存在');
      navigate('/app-management');
      return;
    }
    
    const loadAppDetail = async () => {
      try {
        setLoading(true);
        const data = await fetchAppDetail(id);
        setAppDetail(data);
        form.setFieldsValue({
          name: data.name,
          description: data.description,
          ...data.configs
        });
      } catch (error) {
        message.error('获取应用详情失败');
        console.error('Error fetching app details:', error);
      } finally {
        setLoading(false);
      }
    };
    
    loadAppDetail();
  }, [id, navigate, form]);
  
  const handleBack = () => {
    navigate('/app-management');
  };
  
  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      setSaving(true);
      
      const result = await updateApp(id!, {
        ...appDetail,
        name: values.name,
        description: values.description,
        configs: {
          modelSettings: {
            defaultModel: values.modelSettings?.defaultModel,
            temperature: values.modelSettings?.temperature,
            maxTokens: values.modelSettings?.maxTokens
          },
          apiSettings: {
            rateLimit: values.apiSettings?.rateLimit,
            timeout: values.apiSettings?.timeout
          }
        }
      });
      
      if (result.success) {
        message.success('保存成功');
      } else {
        message.error('保存失败');
      }
    } catch (error) {
      console.error('Save error:', error);
      message.error('表单验证失败');
    } finally {
      setSaving(false);
    }
  };
  
  if (loading) {
    return (
      <div className={styles.loadingContainer}>
        <Spin size="large" />
      </div>
    );
  }
  
  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <Button 
          type="text" 
          icon={<ArrowLeftOutlined />} 
          onClick={handleBack}
          className={styles.backButton}
        >
          返回
        </Button>
        <h1 className={styles.title}>配置应用</h1>
      </div>
      
      <div className={styles.content}>
        <Form
          form={form}
          layout="vertical"
          className={styles.form}
        >
          <Card className={styles.card}>
            <h2 className={styles.sectionTitle}>基本信息</h2>
            <Form.Item 
              name="name" 
              label="应用名称" 
              rules={[
                { required: true, message: '请输入应用名称' },
                { max: 50, message: '应用名称不能超过50个字符' }
              ]}
            >
              <Input placeholder="请输入应用名称" maxLength={50} />
            </Form.Item>
            
            <Form.Item
              name="description"
              label="应用描述"
              rules={[{ max: 200, message: '应用描述不能超过200个字符' }]}
            >
              <Input.TextArea 
                placeholder="请输入应用描述" 
                rows={4} 
                maxLength={200}
              />
            </Form.Item>
          </Card>

          <Tabs
            defaultActiveKey="modelConfig"
            className={styles.tabs}
            items={[
              {
                key: 'modelConfig',
                label: '模型配置',
                children: (
                  <Card className={styles.card}>
                    <Form.Item
                      name={['modelSettings', 'defaultModel']}
                      label="默认模型"
                      rules={[{ required: true, message: '请选择默认模型' }]}
                    >
                      <Input placeholder="请选择默认模型" />
                    </Form.Item>
                    
                    <Form.Item
                      name={['modelSettings', 'temperature']}
                      label="温度"
                      rules={[{ required: true, message: '请输入温度值' }]}
                    >
                      <Input type="number" placeholder="请输入温度值" min={0} max={1} step={0.1} />
                    </Form.Item>
                    
                    <Form.Item
                      name={['modelSettings', 'maxTokens']}
                      label="最大Token数"
                      rules={[{ required: true, message: '请输入最大Token数' }]}
                    >
                      <Input type="number" placeholder="请输入最大Token数" min={1} />
                    </Form.Item>
                  </Card>
                )
              },
              {
                key: 'apiConfig',
                label: 'API配置',
                children: (
                  <Card className={styles.card}>
                    <Form.Item
                      name={['apiSettings', 'rateLimit']}
                      label="速率限制(请求/分钟)"
                      rules={[{ required: true, message: '请设置速率限制' }]}
                    >
                      <Input type="number" placeholder="请设置速率限制" min={1} />
                    </Form.Item>
                    
                    <Form.Item
                      name={['apiSettings', 'timeout']}
                      label="超时时间(秒)"
                      rules={[{ required: true, message: '请设置超时时间' }]}
                    >
                      <Input type="number" placeholder="请设置超时时间" min={1} />
                    </Form.Item>
                  </Card>
                )
              }
            ]}
          />
          
          <div className={styles.footer}>
            <Space>
              <Button onClick={handleBack}>取消</Button>
              <Button type="primary" onClick={handleSave} loading={saving}>
                保存
              </Button>
            </Space>
          </div>
        </Form>
      </div>
    </div>
  );
};

export default AppConfig;
