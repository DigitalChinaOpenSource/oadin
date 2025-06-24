import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Button, Checkbox, Form, List, message, Spin, Tooltip } from 'antd';
import { ArrowLeftOutlined, CopyOutlined, EditOutlined, EyeOutlined, InfoCircleOutlined, PlusOutlined, RobotOutlined, SyncOutlined } from '@ant-design/icons';
import styles from './index.module.scss';
import CreateAppModal from '../AppList/CreateAppModal.tsx';
import ModelSelectModal from './ModelSelectModal/ModelSelectModal.tsx';
import ModelCard from '@/pages/AppManagement/AppConfig/ModelCard/ModelCard.tsx';
import { availableModels } from '@/pages/AppManagement/AppConfig/mock.ts';
import { transformedCard2Ids, transformedCard2Tags } from '@/pages/AppManagement/AppConfig/uitls.ts';
import { IModelSelectCardItem } from '@/pages/AppManagement/AppConfig/types.ts';
import TagFilter, { Tag } from '@/pages/AppManagement/AppConfig/TagFilter/TagFilter.tsx'; // 模拟可用的MCP服务

// 模拟可用的MCP服务
const availableMcps = [
  { value: 'mcp-1', label: 'MCP 服务 1' },
  { value: 'mcp-2', label: 'MCP 服务 2' },
  { value: 'mcp-3', label: 'MCP 服务 3' },
  { value: 'mcp-4', label: 'MCP 服务 4' },
  { value: 'mcp-5', label: 'MCP 服务 5' },
];

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
    mcpCount: 2,
    osCount: 1,
    updatedAt: new Date().toISOString(),
    description: '这是应用的详细描述信息',
    supportedModels: [],
    supportedMcps: [],
    supportedOs: {
      windows: true,
      macos: false,
      ubuntu: true,
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

  // 模型和MCP选择弹窗状态
  const [modelSelectVisible, setModelSelectVisible] = useState(false);
  const [mcpSelectVisible, setMcpSelectVisible] = useState(false);
  const [selectedModels, setSelectedModels] = useState<string[]>([]);
  const [selectedMcps, setSelectedMcps] = useState<string[]>([]);
  const [filterModelTags, setFilterModelTags] = useState<Tag[]>([]);
  const [selectedModelTag, setSelectedModelTag] = useState<string>('全部');
  const [filterModelList, setFilterModelList] = useState<IModelSelectCardItem[]>([]);

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
        } // 设置表单字段
        form.setFieldsValue({
          name: data.name,
          description: data.description,
          supportedModels: data.supportedModels || [],
          supportedMcps: data.supportedMcps || [],
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

  /// 根据外层选择的模型数据生成tags
  useEffect(() => {
    if (appDetail?.supportedModels?.length > 0) {
      const tags = transformedCard2Tags(appDetail?.supportedModels);
      setFilterModelTags(tags);
      setFilterModelList(appDetail?.supportedModels);
    }
  }, [appDetail?.supportedModels]);

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
      // 验证表单数据
      const formValues = await form.validateFields();

      setSaving(true);

      const result = await updateApp(id!, {
        ...appDetail,
        ...formValues, // 包含supportedModels和supportedMcps
        supportedOs: osSelection,
      });
      if (result.success) {
        // 更新本地状态
        setAppDetail({
          ...appDetail,
          ...formValues,
        });
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

      setEditNameModalVisible(false);
    } catch (error) {
      console.error('Edit name error:', error);
      message.error('更新应用名称失败');
    } finally {
      setEditLoading(false);
    }
  };

  // 处理打开模型选择弹窗
  const handleOpenModelSelect = () => {
    // 初始化选中的模型
    setSelectedModels(form.getFieldValue('supportedModels') || []);
    setModelSelectVisible(true);
  };

  // 处理确认模型选择
  const handleModelSelectConfirm = (models: IModelSelectCardItem[]) => {
    // 更新表单中的模型数据
    form.setFieldsValue({ supportedModels: transformedCard2Ids(models) });
    // 更新应用数据
    setAppDetail({
      ...appDetail,
      supportedModels: models,
    });

    // 关闭弹窗
    setModelSelectVisible(false);
  };
  const onRemove = (model: IModelSelectCardItem) => {
    console.info(model, '当前删除的内容');
  };

  // 处理打开MCP选择弹窗
  const handleOpenMcpSelect = () => {
    // 初始化选中的MCP
    setSelectedMcps(form.getFieldValue('supportedMcps') || []);
    setMcpSelectVisible(true);
  };

  // 处理确认MCP选择
  const handleMcpSelectConfirm = (mcps: string[]) => {
    // 更新表单中的MCP数据
    form.setFieldsValue({ supportedMcps: mcps });

    // 更新应用数据
    setAppDetail({
      ...appDetail,
      supportedMcps: mcps,
    });

    // 关闭弹窗
    setMcpSelectVisible(false);
  };

  // 选择model标签筛选
  const onModelTagSelect = (modelTagLabel: string) => {
    setSelectedModelTag(modelTagLabel);
    if (modelTagLabel !== '全部') {
      console.info(modelTagLabel, '当前选择的标签');
      const filteredModels = appDetail?.supportedModels?.filter((model: IModelSelectCardItem) => model.class?.includes(modelTagLabel)) || [];
      setFilterModelList(filteredModels);
    } else {
      setFilterModelList(appDetail?.supportedModels);
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
                      <span>支持的模型</span>

                      <Button
                        className={styles.modelSelectButton}
                        onClick={handleOpenModelSelect}
                        icon={<PlusOutlined />}
                      >
                        选择模型
                      </Button>
                    </div>
                    {filterModelTags.length > 0 ? (
                      <TagFilter
                        onTagSelect={onModelTagSelect}
                        selectedTag={selectedModelTag}
                        tags={filterModelTags}
                      />
                    ) : null}

                    {filterModelList?.length > 0 ? (
                      <List
                        style={{ width: '100%' }}
                        grid={{ gutter: 16, column: 4 }}
                        dataSource={filterModelList}
                        renderItem={(model: IModelSelectCardItem) => (
                          <List.Item>
                            <ModelCard
                              model={model}
                              onRemove={onRemove}
                            />
                          </List.Item>
                        )}
                      />
                    ) : null}
                  </div>

                  {/* 隐藏的表单项用于验证 */}
                  <Form.Item
                    name="supportedModels"
                    rules={[{ required: true, message: '请至少选择一个模型' }]}
                    hidden
                  >
                    <input type="hidden" />
                  </Form.Item>
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
        title="编辑应用名称"
      />

      {/* 模型选择弹窗 */}
      <ModelSelectModal
        open={modelSelectVisible}
        onCancel={() => setModelSelectVisible(false)}
        onFinish={handleModelSelectConfirm}
        title="选择模型"
        modelOptions={availableModels}
        initialSelectedModels={selectedModels}
      />
    </div>
  );
};

export default AppConfig;
