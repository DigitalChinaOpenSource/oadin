import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Button, Checkbox, Form, message, Spin, Tooltip } from 'antd';
import { ArrowLeftOutlined, CopyOutlined, EditOutlined, EyeOutlined, InfoCircleOutlined, RobotOutlined, SyncOutlined } from '@ant-design/icons';
import styles from './index.module.scss';
import CreateAppModal from '../CreateAppModal';

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
    supportedModels: ['gpt-3.5-turbo', 'gpt-4'],
    supportedMcps: ['mcp-1', 'mcp-2'],
    supportedOs: {
      windows: true,
      macos: false,
      ubuntu: true,
    },
    configs: {
      modelSettings: {
        defaultModel: 'gpt-4',
        temperature: 0.7,
        maxTokens: 2048,
      },
      apiSettings: {
        rateLimit: 100,
        timeout: 30,
      },
    },
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
  const [osSelection, setOsSelection] = useState({
    windows: false,
    macos: false,
    ubuntu: false,
  });
  const [agreement, setAgreement] = useState(false);
  const [form] = Form.useForm();

  // 编辑应用名称相关状态
  const [editNameModalVisible, setEditNameModalVisible] = useState(false);
  const [editLoading, setEditLoading] = useState(false);

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

        // 设置操作系统选择状态
        if (data.supportedOs) {
          setOsSelection({
            windows: data.supportedOs.windows || false,
            macos: data.supportedOs.macos || false,
            ubuntu: data.supportedOs.ubuntu || false,
          });
        }

        // 设置表单字段
        form.setFieldsValue({
          name: data.name,
          description: data.description,
          supportedModels: data.supportedModels,
          supportedMcps: data.supportedMcps,
          ...data.configs,
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
    // 验证协议是否勾选
    if (!agreement) {
      message.error('请阅读并同意相关协议');
      return;
    }

    // 验证是否选择了至少一个操作系统
    if (!osSelection.windows && !osSelection.macos && !osSelection.ubuntu) {
      message.error('请至少选择一个操作系统');
      return;
    }

    try {
      setSaving(true);

      const result = await updateApp(id!, {
        ...appDetail,
        supportedOs: osSelection,
        supportedModels: appDetail.supportedModels,
        supportedMcps: appDetail.supportedMcps,
      });

      if (result.success) {
        message.success('保存成功');
      } else {
        message.error('保存失败');
      }
    } catch (error) {
      console.error('Save error:', error);
      message.error('保存失败');
    } finally {
      setSaving(false);
    }
  };

  // 处理编辑应用名称
  const handleEditName = () => {
    setEditNameModalVisible(true);
  };

  // 处理编辑应用名称的提交
  const handleEditNameSubmit = async (values: { name: string }) => {
    try {
      setEditLoading(true);

      // 模拟API调用
      await new Promise((resolve) => setTimeout(resolve, 1000));

      // 更新本地状态
      setAppDetail({
        ...appDetail,
        name: values.name,
      });

      message.success('应用名称更新成功');
      setEditNameModalVisible(false);
    } catch (error) {
      console.error('Edit name error:', error);
      message.error('更新应用名称失败');
    } finally {
      setEditLoading(false);
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
      <div className={styles.contentArea}>
        {/* 固定标题栏 */}
        <div className={styles.header}>
          <div className={styles.headerLeft}>
            <Button
              type="text"
              icon={<ArrowLeftOutlined />}
              onClick={handleBack}
              className={styles.backButton}
            >
              返回
            </Button>
            <span className={styles.title}>配置应用</span>
          </div>
        </div>
        <div>
          {/* 应用信息卡片 */}
          <div className={styles.appInfoCard}>
            <div className={styles.appInfoContent}>
              <div className={styles.appIconContainer}>
                <RobotOutlined className={styles.appIcon} />
              </div>

              <div className={styles.appDetails}>
                <div className={styles.appNameRow}>
                  <Tooltip
                    title={appDetail?.name}
                    placement="topLeft"
                  >
                    <div className={styles.appName}>{appDetail?.name || '应用名称'}</div>
                  </Tooltip>{' '}
                  <Button
                    type="text"
                    icon={<EditOutlined />}
                    className={styles.editNameButton}
                    onClick={handleEditName}
                  />
                </div>

                <div className={styles.appInfoRow}>
                  <div className={styles.appInfoItem}>
                    <span>APP ID：{appDetail?.appId}</span>
                    <CopyOutlined
                      className={styles.copyIcon}
                      onClick={() => {
                        navigator.clipboard.writeText(appDetail?.appId || '');
                        message.success('已复制 APP ID');
                      }}
                    />
                  </div>

                  <div className={styles.appInfoItem}>
                    <span>Secret Key：****************</span>
                    <div className={styles.secretKeyActions}>
                      <EyeOutlined
                        className={styles.actionIcon}
                        onClick={() => message.info('查看密钥')}
                      />
                      <div className={styles.divider} />
                      <SyncOutlined
                        className={styles.actionIcon}
                        onClick={() => message.info('更新密钥')}
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* 配置面板 */}
          <div className={styles.configPanel}>
            <div className={styles.configContent}>
              <Form
                form={form}
                layout="vertical"
                className={styles.form}
              >
                {/* 配置台标题 */}
                <div className={styles.sectionTitleRow}>
                  <div className={styles.sectionTitle}>
                    <div className={styles.titleIndicator} />
                    <span>配置台</span>
                  </div>
                </div>

                {/* 支持的模型 */}
                <div className={styles.configSection}>
                  <div className={styles.configItemRow}>
                    <div className={styles.configLabel}>
                      <span className={styles.requiredMark}>*</span>
                      <span>支持的模型</span>
                    </div>
                    <Button
                      className={styles.modelSelectButton}
                      onClick={() => message.info('选择模型')}
                    >
                      <span className={styles.modelSelectIcon} />
                      <span className={styles.modelSelectText}>选择模型</span>
                    </Button>
                  </div>

                  {/* 支持的MCP服务 */}
                  <div className={styles.configItemRow}>
                    <div className={styles.configLabel}>
                      <span>支持的MCP服务</span>
                    </div>
                    <Button
                      className={styles.mcpSelectButton}
                      onClick={() => message.info('选择MCP')}
                    >
                      <span className={styles.mcpSelectIcon} />
                      <span className={styles.mcpSelectText}>选择MCP</span>
                    </Button>
                  </div>
                </div>

                {/* 操作系统 */}
                <div className={styles.osSection}>
                  <div className={styles.configLabel}>
                    <span className={styles.requiredMark}>*</span>
                    <span>操作系统</span>
                  </div>

                  <div className={styles.osCheckboxGroup}>
                    <Checkbox
                      checked={osSelection.windows}
                      onChange={(e) => setOsSelection({ ...osSelection, windows: e.target.checked })}
                    >
                      Windows
                    </Checkbox>
                    <Checkbox
                      checked={osSelection.macos}
                      onChange={(e) => setOsSelection({ ...osSelection, macos: e.target.checked })}
                    >
                      MacOS
                    </Checkbox>
                    <Checkbox
                      checked={osSelection.ubuntu}
                      onChange={(e) => setOsSelection({ ...osSelection, ubuntu: e.target.checked })}
                    >
                      Ubuntu
                    </Checkbox>
                  </div>

                  <div className={styles.tipsBox}>
                    <InfoCircleOutlined />
                    <span className={styles.tipsText}>点击获取 快速入门手册 ，您也可以在"文档-安装SDK"路径下查看</span>
                  </div>
                </div>

                {/* 协议勾选 */}
                <div className={styles.agreementSection}>
                  <Checkbox
                    className={styles.agreementCheckbox}
                    checked={agreement}
                    onChange={(e) => setAgreement(e.target.checked)}
                  >
                    <span className={styles.agreementText}>阅读并同意《腾讯位置服务开放API服务协议》和《腾讯位置服务隐私协议》</span>
                  </Checkbox>
                </div>

                {/* 底部按钮 */}
                <div className={styles.footerButtons}>
                  <Button
                    type="primary"
                    onClick={handleSave}
                    loading={saving}
                  >
                    确认
                  </Button>
                  <Button onClick={handleBack}>取消</Button>
                </div>
              </Form>
            </div>
          </div>
        </div>
      </div>

      {/* 编辑应用名称弹窗 */}
      <CreateAppModal
        open={editNameModalVisible}
        onCancel={() => setEditNameModalVisible(false)}
        onFinish={handleEditNameSubmit}
        confirmLoading={editLoading}
        initialValues={{ name: appDetail?.name || '' }}
        title="编辑名称"
      />
    </div>
  );
};

export default AppConfig;
