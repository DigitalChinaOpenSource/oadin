import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { IModelAuth, IModelAuthType, IModelPathSpaceRes } from '../types';
import { IModelDataItem, IModelPathRes, IModelSourceType, IRequestModelParams, ModelData } from '@/types';
import { DOWNLOAD_STATUS } from '@/constants';
import { httpRequest } from '@/utils/httpRequest';
import { useDownLoad } from '@/hooks/useDownload';
import { message, Modal } from 'antd';
import { useRequest } from 'ahooks';
import { dealSmartVisionModels } from './utils';
import useModelListStore from '@/store/useModelListStore';
import { convertToMB } from '@/utils';
import useSelectedModelStore from '@/store/useSelectedModel.ts';
import useChatStore from '@/components/chat-container/store/useChatStore.ts';

export type ModelSourceType = 'local' | 'remote';

interface IPagenation {
  current: number;
  pageSize: number;
  total: number;
}
export interface IModelSquareParams {
  flavor?: string;
  // remote时需要传
  // 'dev' | 'product'
  env_type?: string;
  service_source: ModelSourceType;
  page_size?: number;
  page?: number;
  mine?: boolean;
}

export interface IUseViewModel {
  modalPathVisible: boolean;
  onModelPathVisible: () => void;

  modelAuthVisible: boolean;
  onModelAuthVisible: (data: IModelAuth) => void;
  onModelAuthSuccess: () => Promise<void>;

  isDetailVisible: boolean;
  onDetailModalVisible: (visible: boolean, modelData?: IModelDataItem) => void;
  modelAuthType: IModelAuthType;

  onDeleteConfirm: (modelData: IModelDataItem) => void;
  onDownloadConfirm: (modelData: IModelDataItem) => void;

  modelSupportLoading: boolean;
  deleteModelLoading: boolean;

  pagenationData: IModelDataItem[];
  modelSearchVal: string;
  modelSourceVal: ModelSourceType;
  onModelSearch: (val: string) => void;
  selectModelData: IModelDataItem;

  fetchModelSupport: (params: IModelSquareParams) => void;

  pagination: IPagenation;
  onPageChange: (current: number) => void;
  onShowSizeChange: (current: number, pageSize: number) => void;
  modelListStateData: IModelDataItem[];
  mine: boolean;
}

const { confirm } = Modal;
export interface IModelListContent {
  modelSearchVal: string;
  modelSourceVal: IModelSourceType;
  onModelSearch: (val: string) => void;
  mine?: boolean;
  pageType?: number; // 1 体验中心选择
}

export function useViewModel(props: IModelListContent): IUseViewModel {
  const { modelSourceVal, modelSearchVal, onModelSearch, mine } = props;
  // 模型存储路径弹窗是否显示
  const [modalPathVisible, setModalPathVisible] = useState<boolean>(false);
  // 云端模型授权弹窗是否显示
  const [modelAuthVisible, setModelAuthVisible] = useState<boolean>(false);
  // 模型详情弹窗是否显示
  const [isDetailVisible, setIsDetailVisible] = useState(false);
  // 配置 ｜ 更新授权
  const [modelAuthType, setModelAuthType] = useState<IModelAuthType>('config');
  // 模型/问学列表全量数据
  const { setModelListData, modelListData } = useModelListStore();
  const { setCurrentSessionId } = useChatStore();

  // 本地缓存的模型数据
  const [modelListStateData, setModelListStateData] = useState<IModelDataItem[]>([]);
  const [pagination, setPagination] = useState<IPagenation>({
    current: 1,
    pageSize: 12,
    total: 0,
  });
  // 选中的模型数据，暂用于配置授权
  const [selectModelData, setSelectModelData] = useState<IModelDataItem>({} as any);

  const { selectedModel, setIsSelectedModel, setSelectedModel } = useSelectedModelStore();

  const isPageSizeChangingRef = useRef(false);
  const { fetchDownloadStart } = useDownLoad();

  const setListData = (list: IModelDataItem[]) => {
    setModelListData(list);
    setModelListStateData(list);
  };

  /// 可能还有问题， 下面的Effect的作用是当全局的list发生变化的时候更新本地的state
  const prevDownloadStatusMapRef = useRef<Record<string, any>>({});
  useEffect(() => {
    // 创建一个映射来跟踪当前的下载状态
    const currentDownloadStatusMap: Record<string, any> = {};
    // 从downloadList中提取status信息
    modelListData.forEach((item) => {
      if (item.id) {
        currentDownloadStatusMap[item.id] = item.status;
      }
    });

    // 检查列表长度是否发生变化（添加或删除了模型）
    const modelCountChanged = modelListData.length !== modelListStateData.length;

    // 检查是否有状态变化
    let hasStatusChanged = false;

    // 检查新增或修改的状态
    for (const id in currentDownloadStatusMap) {
      if (prevDownloadStatusMapRef.current[id] !== currentDownloadStatusMap[id]) {
        hasStatusChanged = true;
        break;
      }
    }

    // 检查删除的状态
    if (!hasStatusChanged) {
      for (const id in prevDownloadStatusMapRef.current) {
        if (!(id in currentDownloadStatusMap)) {
          hasStatusChanged = true;
          break;
        }
      }
    }

    // 如果列表数量或状态有变化，更新modelListStateData
    if (modelCountChanged || hasStatusChanged) {
      console.info('模型列表或下载状态发生变化，更新modelListStateData');

      if (modelCountChanged) {
        // 列表数量变化，直接使用新列表
        setModelListStateData(modelListData);
      } else {
        // 只是状态变化，更新现有项的状态
        const updatedList = modelListStateData.map((item) => {
          const downItem = modelListData.find((_item) => _item.id === item.id);
          if (downItem) {
            return {
              ...item,
              status: downItem?.status,
              can_select: downItem?.can_select,
            };
          }
          return item;
        });
        console.info('更新后的模型列表数据:', updatedList);
        if (updatedList.length > 0) {
          setModelListStateData(updatedList);
        }
      }

      // 更新记录的状态
      prevDownloadStatusMapRef.current = { ...currentDownloadStatusMap };
    }
  }, [modelListData, modelListStateData]);

  // 获取模型列表 （本地和云端）
  const { loading: modelSupportLoading, run: fetchModelSupport } = useRequest(
    async (params: IModelSquareParams) => {
      const paramsTemp = {
        ...params,
        page_size: 999,
        mine,
      };
      if (params?.service_source === 'remote') {
        paramsTemp.env_type = 'product';
      }
      const data = await httpRequest.get<ModelData>('/control_panel/model/square', paramsTemp);
      if (paramsTemp.service_source === 'remote') {
        // 处理问学模型列表的数据, 把推荐的模型放在前面
        return dealSmartVisionModels(data?.data || []);
      }
      return data?.data || [];
    },
    {
      manual: true,
      onSuccess: (data) => {
        // 处理一些数据格式
        const dataWithSource = (data || []).map(
          (item) =>
            ({
              ...item,
              currentDownload: 0,
            }) as any,
        );
        console.info(dataWithSource, 'dataWithSourcedataWithSource');
        setListData(dataWithSource);
        setPagination({
          ...pagination,
          total: dataWithSource.length,
        });
      },
      onError: (error) => {
        console.error('获取模型列表失败:', error);
        setListData([]);
      },
    },
  );

  // 必须，下载时需要获取当前路径的存储空间
  useEffect(() => {
    fetchModelPath();
  }, []);

  useEffect(() => {
    onModelSearch('');
    setPagination({ ...pagination, current: 1, total: 0 });
    fetchModelSupport({ service_source: modelSourceVal });
  }, [modelSourceVal, mine]);

  // 获取模型存储路径
  const { run: fetchModelPath } = useRequest(
    async () => {
      const res = await httpRequest.get<IModelPathRes>('/control_panel/model/filepath');
      return res || {};
    },
    {
      manual: true,
      onSuccess: async (data) => {
        await onCheckPathSpace(data?.path || '');
      },
      onError: (error) => {
        console.error('获取模型存储路径失败:', error);
      },
    },
  );
  // 根据搜索值和分页参数更新分页数据
  const prevModelSearchValRef = useRef(modelSearchVal);

  useEffect(() => {
    // 只有当搜索值真正变化时才执行，避免路由切换触发
    if (prevModelSearchValRef.current !== modelSearchVal) {
      const filteredData = getFilteredData();
      setPagination({
        current: 1,
        pageSize: pagination.pageSize,
        total: filteredData.length,
      });
      prevModelSearchValRef.current = modelSearchVal;
    }
  }, [modelSearchVal]);

  // 删除模型
  const { loading: deleteModelLoading, run: fetchDeleteModel } = useRequest(
    async (params: IRequestModelParams) => {
      await httpRequest.del('/model', params);
    },
    {
      manual: true,
      onSuccess: (data, params) => {
        message.success('模型删除成功');
        if (params[0].model_name === selectedModel?.name) {
          setIsSelectedModel(false);
          setSelectedModel(null);
          setCurrentSessionId('');
        }
      },
      onError: (error: Error & { handled?: boolean }) => {
        if (!error?.handled) {
          message.error('模型删除失败');
        }
        console.error('删除模型失败:', error);
      },
      onFinally: async () => {
        // 保留当前页码，重新获取数据
        const currentPage = pagination.current;
        // 等待获取新数据完成
        await fetchModelSupport({ service_source: modelSourceVal });
        // 等待React状态更新后再获取过滤数据
        setTimeout(() => {
          const filteredData = getFilteredData();
          const totalPages = Math.ceil(filteredData.length / pagination.pageSize);
          const newPage = currentPage > totalPages ? totalPages || 1 : currentPage;
          console.info(filteredData, '删除后-过滤后的数据');
          setPagination({
            ...pagination,
            current: newPage,
            total: filteredData.length,
          });
        }, 0);
      },
    },
  );

  // 获取过滤后的数据
  const getFilteredData = () => {
    if (!modelSearchVal || modelSearchVal.trim() === '') {
      return modelListStateData;
    }
    return modelListStateData.filter((model) => model.name && model.name.toLowerCase().includes(modelSearchVal.toLowerCase()));
  };

  // 添加分页数据计算逻辑
  const paginatedData = (pagination: IPagenation, data = getFilteredData()) => {
    const { current, pageSize } = pagination;
    const startIndex = (current - 1) * pageSize;
    const endIndex = current * pageSize;
    return data.slice(startIndex, endIndex);
  };

  // 计算分页数据，过滤后的，用于渲染
  const pagenationData = useMemo(() => {
    const filteredData = getFilteredData();
    return paginatedData(pagination, filteredData);
  }, [modelListStateData, modelSearchVal, pagination]);

  const onPageChange = (current: number) => {
    if (isPageSizeChangingRef.current) {
      isPageSizeChangingRef.current = false;
      return;
    }
    setPagination({ ...pagination, current });
  };

  const onShowSizeChange = (current: number, pageSize: number) => {
    isPageSizeChangingRef.current = true;
    setPagination({ ...pagination, current: 1, pageSize });
  };

  // 模型详情弹窗
  const onDetailModalVisible = useCallback((visible: boolean, modelData?: IModelDataItem) => {
    setIsDetailVisible(visible);
    // 弹窗关闭清空选择的模型数据
    setSelectModelData(modelData || ({} as any));
  }, []);

  const { run: onCheckPathSpace, data: currentPathSpace } = useRequest(
    async (path: string) => {
      const data = await httpRequest.get<IModelPathSpaceRes>('/control_panel/path/space', { path });
      return data || {};
    },
    {
      manual: true,
    },
  );

  // 模型存储路径弹窗
  const onModelPathVisible = useCallback(() => {
    setModalPathVisible(!modalPathVisible);
  }, [modalPathVisible]);

  const onModelAuthVisible = useCallback((data: IModelAuth) => {
    setModelAuthVisible(data.visible);
    setModelAuthType(data.type);
    setSelectModelData(data.modelData);
  }, []);

  const onDownloadConfirm = (modelData: IModelDataItem) => {
    if (props.pageType === 1) {
      startDownload(modelData);
    } else {
      confirm({
        title: '确认下载此模型？',
        okText: '确认下载',
        centered: true,
        okButtonProps: {
          style: { backgroundColor: '#4f4dff' },
        },
        async onOk() {
          startDownload(modelData);
        },
        onCancel() {
          console.log('Cancel');
        },
      });
    }
  };
  const startDownload = (modelData: IModelDataItem) => {
    const modelSizeMb = convertToMB(modelData.size || '0MB');
    const freeSpaceMb = (currentPathSpace?.free_size || 0) * 1024;
    if (modelSizeMb > freeSpaceMb) {
      message.warning('当前路径下的磁盘空间不足，无法下载该模型');
      return;
    } else {
      fetchDownloadStart({
        ...modelData,
        status: DOWNLOAD_STATUS.IN_PROGRESS,
      });
    }
  };

  const onDeleteConfirm = (modelData: IModelDataItem) => {
    confirm({
      title: '是否确认删除此模型',
      content: '此操作将删除您与该模型的所有对话记录、文件信息，并终止相关进程。',
      centered: true,
      okButtonProps: {
        style: { backgroundColor: '#e85951' },
      },
      okText: '确认删除',
      onOk() {
        // 组装请求参数
        const params = {
          model_name: modelData.name,
          service_name: modelData.service_name || 'chat',
          service_source: modelData.source || 'local',
          provider_name: modelData.service_provider_name || 'local_ollama_chat',
        };
        fetchDeleteModel(params);
      },
    });
  };

  // 授权成功刷新列表
  const onModelAuthSuccess = async () => {
    await fetchModelSupport({ service_source: modelSourceVal });
    const filteredData = getFilteredData();
    setPagination({
      ...pagination,
      total: filteredData.length,
    });
  };

  return {
    modalPathVisible,
    onModelPathVisible,

    modelAuthVisible,
    onModelAuthVisible,
    onModelAuthSuccess,

    isDetailVisible,
    onDetailModalVisible,
    modelAuthType,

    onDeleteConfirm,
    onDownloadConfirm,

    modelSupportLoading,
    deleteModelLoading,

    pagenationData,
    modelSearchVal,
    modelSourceVal,
    onModelSearch,
    selectModelData,
    fetchModelSupport,

    modelListStateData,
    pagination,
    onPageChange,
    onShowSizeChange,

    mine: !!mine,
  };
}
