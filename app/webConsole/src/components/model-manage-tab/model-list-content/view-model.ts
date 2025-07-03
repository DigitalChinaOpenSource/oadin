import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { IModelAuth, IModelAuthType, IModelPathSpaceRes } from '../types';
import { IModelDataItem, IModelPathRes, IModelSourceType, IModelSquareParams, IRequestModelParams, ModelData, ModelSourceType } from '@/types';
import { DOWNLOAD_STATUS } from '@/constants';
import { httpRequest } from '@/utils/httpRequest';
import { useDownLoad } from '@/hooks/useDownload';
import { message, Modal } from 'antd';
import { useRequest } from 'ahooks';
import useModelListStore from '@/store/useModelListStore';
import { convertToMB } from '@/utils';
import useSelectedModelStore from '@/store/useSelectedModel.ts';

interface IPagenation {
  current: number;
  pageSize: number;
  total: number;
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
  // 自定义设置列表数据的函数
  customSetListData?: (list: IModelDataItem[]) => void;
  // 自定义模型列表数据源
  customModelListData?: IModelDataItem[];
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

  // 本地缓存的模型数据
  const [modelListStateData, setModelListStateData] = useState<IModelDataItem[]>([]);
  const [pagination, setPagination] = useState<IPagenation>({
    current: 1,
    pageSize: 12,
    total: 0,
  });
  // 选中的模型数据，暂用于配置授权
  const [selectModelData, setSelectModelData] = useState<IModelDataItem>({} as any);

  const { selectedModel, setSelectedModel } = useSelectedModelStore();

  const isPageSizeChangingRef = useRef(false);
  const { fetchDownloadStart } = useDownLoad();
  const setListData = (list: IModelDataItem[]) => {
    // 创建一个深拷贝，避免引用问题
    const listCopy = JSON.parse(JSON.stringify(list));
    
    // 如果提供了自定义的设置列表数据函数，则使用它
    if (props.customSetListData) {
      props.customSetListData(listCopy);
    } else {
      // 否则使用默认的全局状态管理
      setModelListData(listCopy);
    }
    // 初始化或重新设置数据时，都更新本地状态
    setModelListStateData(listCopy);
    // 重置实例ID标识，确保可以追踪当前组件实例
    instanceIdRef.current = `${modelSourceVal}-${mine ? 'mine' : 'all'}`;
  };
  /// 监听全局的list变化并更新本地的state
  const prevDownloadStatusMapRef = useRef<Record<string, any>>({});
  // 使用一个ref记录当前实例的模型源和mine类型，用于识别实例身份
  const instanceIdRef = useRef<string>(`${modelSourceVal}-${mine ? 'mine' : 'all'}`);
  // 使用一个ref记录是否刚删除了模型
  const justDeletedRef = useRef<boolean>(false);
  useEffect(() => {
    // 使用自定义模型列表数据源或全局模型列表数据
    const currentModelListData = props.customModelListData || modelListData;

    // 创建一个映射来跟踪当前的下载状态
    const currentDownloadStatusMap: Record<string, any> = {};
    currentModelListData.forEach((item) => {
      if (item.id) {
        currentDownloadStatusMap[item.id] = item.status;
      }
    });

    // 检查数据源实例是否变化（remote/local切换）
    const currentInstanceId = `${modelSourceVal}-${mine ? 'mine' : 'all'}`;
    const isInstanceChanged = instanceIdRef.current !== currentInstanceId;

    // 如果是刚删除模型的情况或者数据源变化，我们需要强制更新列表
    if (justDeletedRef.current || isInstanceChanged) {
      console.info('检测到模型刚被删除或数据源变化，强制更新modelListStateData');
      setModelListStateData(currentModelListData);
      justDeletedRef.current = false;
      // 如果是数据源变化，更新实例ID
      if (isInstanceChanged) {
        instanceIdRef.current = currentInstanceId;
      }
      prevDownloadStatusMapRef.current = { ...currentDownloadStatusMap };
      return;
    }

    // 检查列表长度是否变化
    const isLengthChanged = currentModelListData.length !== modelListStateData.length;
    if (isLengthChanged) {
      setModelListStateData(currentModelListData);
      prevDownloadStatusMapRef.current = { ...currentDownloadStatusMap };
      return;
    }

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
    // 如果有状态变化，更新modelListStateData
    if (hasStatusChanged) {
      // 只更新现有项的状态，不处理列表长度变化
      const updatedList = modelListStateData.map((item) => {
        const downItem = currentModelListData.find((_item) => _item.id === item.id);
        if (downItem) {
          return {
            ...item,
            status: downItem?.status,
            can_select: downItem?.can_select,
          };
        }
        return item;
      });

      if (updatedList.length > 0) {
        setModelListStateData(updatedList);
      }

      // 更新记录的状态
      prevDownloadStatusMapRef.current = { ...currentDownloadStatusMap };
    }
  }, [props.customModelListData, modelListData, modelListStateData]);

  // 获取模型列表 （本地和云端）
  const { loading: modelSupportLoading, run: fetchModelSupport } = useRequest(
    async (params: IModelSquareParams) => {
      const paramsTemp = {
        ...params,
        page_size: 999,
        // 只有在params中没有指定mine的情况下才使用props.mine
        mine: params.mine !== undefined ? params.mine : mine,
      };
      if (params?.service_source === 'remote') {
        paramsTemp.env_type = 'product';
      }
      const data = await httpRequest.get<ModelData>('/control_panel/model/square', paramsTemp);
      return data?.data || [];
    },
    {
      manual: true,
      onSuccess: (data) => {
        // 处理一些数据格式
        let dataWithSource = (data || []).map(
          (item) =>
            ({
              ...item,
              currentDownload: 0,
            }) as any,
        );
        if (props.pageType === 1) {
          dataWithSource = dataWithSource.filter((item) => {
            return item.class.every((c_item: string) => !c_item.includes('嵌入'));
          });
        }

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
    // 重置下载状态映射
    prevDownloadStatusMapRef.current = {};
    fetchModelSupport({
      service_source: modelSourceVal,
      mine: mine,
    });
    // 注意：instanceIdRef.current 的更新已移至监视 modelListData 的 useEffect 中，
    // 确保数据和 ID 的更新一致
  }, [modelSourceVal, mine]);

  // 添加新的 useEffect 来监听自定义模型列表数据的变化
  useEffect(() => {
    if (props.customModelListData && props.customModelListData.length > 0) {
      // 当自定义数据源发生变化时，更新本地状态
      setModelListStateData(props.customModelListData);
      
      // 同时更新分页信息中的total
      const filteredData = props.customModelListData.filter((model) => 
        !modelSearchVal || 
        (model.name && model.name.toLowerCase().includes(modelSearchVal.toLowerCase()))
      );
      
      console.log("customModelListData变化，更新pagination.total:", filteredData.length);
      setPagination(prev => ({
        ...prev,
        total: filteredData.length
      }));
    }
  }, [props.customModelListData, modelSearchVal]);

  // 获取模型存储路径
  const { run: fetchModelPath, data: modelPathData } = useRequest(
    async () => {
      const res = await httpRequest.get<IModelPathRes>('/control_panel/model/filepath');
      return res || {};
    },
    {
      manual: true,
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
      const data = await httpRequest.del('/model', params);
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data, params) => {
        message.success('模型删除成功');
        if (params[0].model_name === selectedModel?.name) {
          setSelectedModel(null);
        }
      },
      onError: (error: Error & { handled?: boolean }) => {
        if (!error?.handled) {
          message.error('模型删除失败');
        }
      },
      onFinally: async () => {
        // 保留当前页码，重新获取数据
        const currentPage = pagination.current;
        // 标记刚刚删除了模型
        justDeletedRef.current = true;
        // 等待获取新数据完成 - 传递mine参数保持当前上下文
        await fetchModelSupport({
          service_source: modelSourceVal,
          mine: mine,
        });
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
  }, [modelListStateData, modelSearchVal, pagination, props.customModelListData]);

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

  const { runAsync: onCheckPathSpace } = useRequest(
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
  const startDownload = async (modelData: IModelDataItem) => {
    const modelSizeMb = convertToMB(modelData.size || '0MB');
    const currentPathSpace = await onCheckPathSpace(modelPathData?.path || '');
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
    await fetchModelSupport({
      service_source: modelSourceVal,
      mine: mine,
    });
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
