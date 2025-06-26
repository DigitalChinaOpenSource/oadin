import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Button, Checkbox, Form, List, message, Spin, Tooltip } from 'antd';
import { ArrowLeftOutlined, CopyOutlined, EditOutlined, EyeOutlined, InfoCircleOutlined, PlusOutlined, RobotOutlined, SyncOutlined } from '@ant-design/icons';
import styles from './index.module.scss';
import CreateAppModal from '../AppList/CreateAppModal.tsx';
import ModelSelectModal from './ModelSelectModal/ModelSelectModal.tsx';
import ModelCard from '@/pages/AppManagement/AppConfig/ModelCard/ModelCard.tsx';
import { availableMcps, availableModels } from '@/pages/AppManagement/AppConfig/mock.ts';
import { transformedCard2Ids, transformedCard2Tags, transformedMcp2Card, transformedModel2Card } from '@/pages/AppManagement/AppConfig/uitls.ts';
import { ICardDeatilItem, IModelSelectCardItem, searchFunc, SearchParams } from '@/pages/AppManagement/AppConfig/types.ts';
import TagFilter, { Tag } from '@/pages/AppManagement/AppConfig/TagFilter/TagFilter.tsx';

interface AppConfigProps {}

export const defaultTag = '全部';

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
  // 当前选中的模型
  const [selectedModels, setSelectedModels] = useState<IModelSelectCardItem[]>([]);
  const [filterModelTags, setFilterModelTags] = useState<Tag[]>([]);
  const [selectedModelTag, setSelectedModelTag] = useState<string>(defaultTag);
  const [filterModelList, setFilterModelList] = useState<IModelSelectCardItem[]>([]);
  /// 搜索结果
  const [searchList, setSearchList] = useState<ICardDeatilItem[]>([]);
  // 当前选中的MCP
  const [selectedMcps, setSelectedMcps] = useState<IModelSelectCardItem[]>([]);
  const [filterMcpTags, setFilterMcpTags] = useState<Tag[]>([]);
  const [selectedMcpTag, setSelectedMcpTag] = useState<string>(defaultTag);
  const [filterMcpList, setFilterMcpList] = useState<IModelSelectCardItem[]>([]);
  const [searchMcpList, setSearchMcpList] = useState<ICardDeatilItem[]>([]);
  /// 搜索条件
  const [searchParamsState, setSearchParamsState] = useState<SearchParams>({});

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

  /// 根据详情数据初始化内部是否已选中数据
  useEffect(() => {
    if (appDetail?.supportedModels?.length > 0) {
      setSelectModelsToForm(appDetail?.supportedModels);
    }
  }, [appDetail?.supportedModels]);
  /// 根据详情数据初始化内部是否已选中数据
  useEffect(() => {
    if (appDetail?.supportedMcps?.length > 0) {
      setSelectMcpToForm(appDetail?.supportedMcps);
    }
  }, [appDetail?.supportedMcps]);

  /// 根据选中的模型初始化需要显示的模型
  useEffect(() => {
    console.info(selectedModels, 'selectedModelsselectedModels');

    const tags = transformedCard2Tags(selectedModels);
    setFilterModelTags(tags);
    setFilterModelList(selectedModels);
  }, [selectedModels]);
  /// 根据选中的MCP初始化需要显示的模型
  useEffect(() => {
    const tags = transformedCard2Tags(selectedMcps);
    setFilterMcpTags(tags);
    setFilterMcpList(selectedMcps);
  }, [selectedMcps]);

  const setSelectModelsToForm = (models: IModelSelectCardItem[]) => {
    setSelectedModels(models);
    form.setFieldsValue({ supportedModels: transformedCard2Ids(models) });
  };
  const setSelectMcpToForm = (models: IModelSelectCardItem[]) => {
    setSelectedMcps(models);
    form.setFieldsValue({ supportedMcps: transformedCard2Ids(models) });
  };

  const handleBack = () => {
    navigate('/app-management');
  };
  //  保存
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
    setModelSelectVisible(true);
  };

  // 处理打开模型选择弹窗
  const handleOpenMcpSelect = () => {
    setMcpSelectVisible(true);
  };

  // 处理确认模型选择
  const handleModelSelectConfirm = (models: IModelSelectCardItem[]) => {
    // 更新表单中的模型数据
    setSelectModelsToForm(models);
    setSelectedModelTag(defaultTag);
    // 关闭弹窗
    setModelSelectVisible(false);
  };
  // 删除模型
  const onRemoveModel = (model: IModelSelectCardItem) => {
    const filterList = selectedModels.filter((item) => {
      return item.id !== model.id;
    });
    setSelectModelsToForm(filterList);
  };

  // 处理确认模型选择
  const handleMcpSelectConfirm = (models: IModelSelectCardItem[]) => {
    // 更新表单中的模型数据
    setSelectMcpToForm(models);
    setSelectedMcpTag(defaultTag);
    // 关闭弹窗
    setMcpSelectVisible(false);
  };
  // 删除模型
  const onRemoveMcp = (model: IModelSelectCardItem) => {
    const filterList = selectedMcps.filter((item) => {
      return item.id !== model.id;
    });
    setSelectMcpToForm(filterList);
  };

  // 选择model标签筛选
  const onModelTagSelect = (modelTagLabel: string) => {
    setSelectedModelTag(modelTagLabel);
    if (modelTagLabel !== defaultTag) {
      console.info(modelTagLabel, '当前选择的标签');
      const filteredModels = selectedModels?.filter((model: IModelSelectCardItem) => model.class?.includes(modelTagLabel)) || [];
      setFilterModelList(filteredModels);
    } else {
      setFilterModelList(selectedModels);
    }
  };
  // 选择mcp标签筛选
  const onMcpTagSelect = (mcpTagLabel: string) => {
    setSelectedMcpTag(mcpTagLabel);
    if (mcpTagLabel !== defaultTag) {
      const filteredMcps = selectedMcps?.filter((model: IModelSelectCardItem) => model.class?.includes(mcpTagLabel)) || [];
      setFilterMcpList(filteredMcps);
    } else {
      setFilterMcpList(selectedMcps);
    }
  };

  const onSearch: searchFunc = async (params?: SearchParams) => {
    const _searchParams = { ...searchParamsState, ...params };
    setSearchParamsState(_searchParams);
    message.info(`搜索条件为:${JSON.stringify(params)}`);
    /// TODO 接口调用
    /// 接口返回的数据进行转换
    const dealList = transformedModel2Card(availableModels);
    setSearchList(dealList);
    return dealList;
  };
  const onMcpSearch: searchFunc = async (params?: SearchParams) => {
    const _searchParams = { ...searchParamsState, ...params };
    setSearchParamsState(_searchParams);
    message.info(`搜索条件为:${JSON.stringify(params)}`);
    /// TODO 接口调用

    const dealList = transformedMcp2Card(availableMcps);
    setSearchMcpList(dealList);
    return dealList;
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
                      <span>选择模型</span>
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
                              onRemove={onRemoveModel}
                            />
                          </List.Item>
                        )}
                      />
                    ) : null}
                  </div>

                  <div className={styles.configItemRow}>
                    <div className={styles.configLabel}>
                      <span>选择MCP服务</span>
                      <Button
                        className={styles.modelSelectButton}
                        onClick={handleOpenMcpSelect}
                        icon={<PlusOutlined />}
                      >
                        选择MCP
                      </Button>
                    </div>
                    {filterMcpTags.length > 0 ? (
                      <TagFilter
                        onTagSelect={onMcpTagSelect}
                        selectedTag={selectedMcpTag}
                        tags={filterMcpTags}
                      />
                    ) : null}

                    {filterMcpList?.length > 0 ? (
                      <List
                        style={{ width: '100%' }}
                        grid={{ gutter: 16, column: 4 }}
                        dataSource={filterMcpList}
                        renderItem={(model: IModelSelectCardItem) => (
                          <List.Item>
                            <ModelCard
                              model={model}
                              onRemove={onRemoveMcp}
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
                  {/* 隐藏的表单项用于验证 */}
                  <Form.Item
                    name="supportedMcps"
                    rules={[{ required: true, message: '请至少选择一个Mcp' }]}
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
      {modelSelectVisible ? (
        <ModelSelectModal
          open={modelSelectVisible}
          onCancel={() => setModelSelectVisible(false)}
          onFinish={handleModelSelectConfirm}
          title="选择模型"
          onSearch={onSearch}
          searchList={searchList}
          initialSelectedModels={selectedModels}
        />
      ) : null}

      {/* MCP选择弹窗 */}
      {mcpSelectVisible ? (
        <ModelSelectModal
          open={mcpSelectVisible}
          onCancel={() => setMcpSelectVisible(false)}
          onFinish={handleMcpSelectConfirm}
          title="选择模型"
          onSearch={onMcpSearch}
          searchList={searchMcpList}
          initialSelectedModels={selectedMcps}
        />
      ) : null}
    </div>
  );
};

export default AppConfig;
