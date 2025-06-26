import React, { Dispatch, SetStateAction, useEffect, useState } from 'react';
import { Checkbox, Form, Modal, Space, Tooltip } from 'antd';
import styles from './ModelSelectModal.module.scss';
import { transformedCard2Ids, transformedIds2Card } from '@/pages/AppManagement/AppConfig/uitls.ts';
import { ICardDeatilItem, IModelSelectCardItem, searchFunc } from '@/pages/AppManagement/AppConfig/types.ts';
import { ModelSearch } from '@/pages/AppManagement/AppConfig/ModelSelectModal/ModelSearch.tsx';
import { ITagsDataItem } from '@/types/model.ts';
import { ModelContent } from '@/pages/AppManagement/AppConfig/ModelSelectModal/ModelContent.tsx';
import McpAdvanceFilter from '@/pages/AppManagement/AppConfig/ModelSelectModal/McpAdvanceFilter';
import expandSvg from '@/components/icons/expand.svg';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';

// 模型选择弹窗
export interface ModelSelectModalProps {
  open: boolean;
  onCancel: () => void;
  onFinish: (selectedModels: IModelSelectCardItem[]) => void;
  title: string;
  confirmLoading?: boolean;
  onSearch: searchFunc;
  initialSelectedModels?: IModelSelectCardItem[];
  searchList: ICardDeatilItem[];
  setDrawerOpenId?: Dispatch<SetStateAction<string>>;
}

const ModelSelectModal: React.FC<ModelSelectModalProps> = ({ setDrawerOpenId, searchList, onSearch, open, onCancel, onFinish, title, confirmLoading = false, initialSelectedModels = [] }) => {
  const [form] = Form.useForm();
  /// 当前选中的id
  const [selectedModelIds, setSelectedModelIds] = useState<string[]>(transformedCard2Ids(initialSelectedModels) || []);
  // 仅显示已选
  const [showOnlySelecte, setShowOnlySelecte] = useState<boolean>(false);

  // 搜索栏
  const [searchText, setSearchText] = useState<string>('');
  // 内部筛选数据
  const [filterSearchList, setFilterSearchList] = useState<ICardDeatilItem[]>(searchList);
  // 右侧筛选器的筛选项
  const [checkedValues, setCheckedValues] = useState<string[]>([]);
  const [tagsData, setTagsData] = useState<ITagsDataItem[]>([{ category: '', tags: [] }]);

  // 打开收起筛选栏
  const [collapsed, setCollapsed] = useState<boolean>(false);
  // 获取所有mcp服务的筛选tags
  const { loading: tagsLoading, run: getTagsData } = useRequest(
    async () => {
      return await httpRequest.get('/mcp/categories');
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('tagsData===>', data);
        setTagsData(data || []);
        // 初始化标签数据
        const initData = data.reduce((acc: Record<string, any>, item: any) => {
          acc[item.category] = [];
          return acc;
        }, {});
        console.info(initData, 'initData');
        setCheckedValues(initData);
      },

      onError: (error) => {
        console.error('获取模型列表失败:', error);
      },
    },
  );
  useEffect(() => {
    getTagsData();
  }, []);
  // 当初始值或弹窗打开状态变化时更新选中的模型
  useEffect(() => {
    if (open) {
      setSelectedModelIds(transformedCard2Ids(initialSelectedModels));
      form.setFieldsValue({ models: transformedCard2Ids(initialSelectedModels) });
    }
  }, [initialSelectedModels, open, form]);

  useEffect(() => {
    onSearch();
  }, []);

  useEffect(() => {
    if (showOnlySelecte) {
      const _filterList = searchList.filter((item) => {
        return selectedModelIds.includes(item.id);
      });
      setFilterSearchList(_filterList);
    } else {
      setFilterSearchList(searchList);
    }
  }, [showOnlySelecte]);

  useEffect(() => {
    console.info(searchList, '腹肌数据');
    setFilterSearchList(searchList);
  }, [searchList]);

  const handleSubmit = async () => {
    try {
      const cardList = transformedIds2Card(searchList, selectedModelIds);
      console.info(cardList, 'cardListcardList');
      onFinish(cardList || []);
    } catch (error) {
      console.error('表单验证失败:', error);
    }
  };
  const handleClearTags = () => {};
  const handleTagsChange = (category: string, list: any[]) => {
    const updatedCheckedValues = {
      ...checkedValues,
      [category]: list,
    };
    setCheckedValues(updatedCheckedValues);
    // TODO 搜索
    onSearch();
  };

  return (
    <Modal
      title={title}
      open={open}
      onCancel={onCancel}
      onOk={handleSubmit}
      confirmLoading={confirmLoading}
      className={styles.choose_model}
      width={1000}
      footer={(_, { OkBtn, CancelBtn }) => (
        <div className={styles.choose_model_footer}>
          <Checkbox
            checked={showOnlySelecte}
            onChange={(e) => {
              setShowOnlySelecte(e.target.checked);
            }}
          >
            仅显示已选
          </Checkbox>
          <Space>
            <CancelBtn />
            <OkBtn />
          </Space>
        </div>
      )}
    >
      <div className={styles.modalWarp}>
        <div className={styles.seacrchWarp}>
          <ModelSearch
            searchText={searchText}
            onSearch={onSearch}
            setSearchText={setSearchText}
          />
          {collapsed ? (
            <Tooltip title="展开筛选">
              <div
                className={styles.expandIcon}
                onClick={() => setCollapsed(false)}
              >
                <img
                  src={expandSvg}
                  alt="折叠筛选面板"
                />
              </div>
            </Tooltip>
          ) : null}
        </div>
        <ModelContent
          setSelectedModelIds={setSelectedModelIds}
          setDrawerOpenId={setDrawerOpenId}
          selectedModelIds={selectedModelIds}
          onSearch={onSearch}
          showOnlySelecte={showOnlySelecte}
          filterSearchList={filterSearchList}
        />
      </div>
      <McpAdvanceFilter
        collapsed={collapsed}
        setCollapsed={setCollapsed}
        handleClearTags={handleClearTags}
        tagsData={tagsData}
        checkedValues={checkedValues}
        handleTagsChange={handleTagsChange}
      />
    </Modal>
  );
};

export default ModelSelectModal;
