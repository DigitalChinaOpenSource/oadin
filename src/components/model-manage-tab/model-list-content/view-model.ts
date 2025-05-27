import { useState, useCallback, useEffect, useRef } from 'react';
import { IModelAuthType, IModelAuth } from '../types';
import { ModelData, IModelDataItem, IRequestModelParams, IModelPathRes } from '@/types';
import { DOWNLOAD_STATUS } from '@/constants';
import { httpRequest } from '@/utils/httpRequest';
import { useDownLoad } from '@/hooks/useDownload';
import { Modal, message } from 'antd';
import { IModelListContent } from './index';
import { useRequest } from 'ahooks';
import { dealSmartVisionModels } from './utils';
import useModelListStore from '@/store/useModelListStore';

interface IPagenation {
  current: number;
  pageSize: number;
  total: number;
}
interface IModelSquareParams {
  flavor?: string;
  // remote时需要传
  // 'dev' | 'product'
  env_type?: string;
  service_source: 'local' | 'remote';
  page_size?: number;
  page?: number;
}

const { confirm } = Modal;

export function useViewModel(props: IModelListContent) {
  const { modelSourceVal, modelSearchVal, onModelSearch } = props;
  // 模型存储路径弹窗是否显示
  const [modalPathVisible, setModalPathVisible] = useState<boolean>(false);
  // 接口获取
  const [modelPath, setModelPath] = useState<string>('');
  // 云端模型授权弹窗是否显示
  const [modelAuthVisible, setModelAuthVisible] = useState<boolean>(false);
  // 模型详情弹窗是否显示
  const [isDetailVisible, setIsDetailVisible] = useState(false);
  // 配置 ｜ 更新授权
  const [modelAuthType, setModelAuthType] = useState<IModelAuthType>('config');
  // 模型/问学列表全量数据
  const { modelListData, setModelListData } = useModelListStore();
  // 分页数据，用于展示
  const [pagenationData, setPagenationData] = useState<IModelDataItem[]>([]);
  const [pagination, setPagination] = useState<IPagenation>({
    current: 1,
    pageSize: 12,
    total: 0,
  });
  // 选中的模型数据，暂用于配置授权
  const [selectModelData, setSelectModelData] = useState<IModelDataItem>({} as any);

  const lastPageSizeRef = useRef(pagination.pageSize);
  const { fetchDownloadStart } = useDownLoad();

  // 获取模型列表 （本地和云端）
  const { loading: modelSupportLoading, run: fetchModelSupport } = useRequest(
    async (params: IModelSquareParams) => {
      const paramsTemp = {
        ...params,
        page_size: 999,
      };
      if (params?.service_source === 'remote') {
        paramsTemp.env_type = import.meta.env.MODE || 'product';
      }
      const data = await httpRequest.get<ModelData>('/control_panel/model/square', paramsTemp);
      if (paramsTemp.service_source === 'remote') {
        // 处理问学模型列表的数据, 把推荐的模型放在前面
        const remoteData = dealSmartVisionModels(data?.data || []);
        return remoteData;
      }
      return data?.data || [];
    },
    {
      manual: true,
      onSuccess: (data) => {
        // 处理一些数据格式
        const dataWithSource = (data || []).map(
          (item, index) =>
            ({
              ...item,
              type: 0,
              id: index + 1,
              currentDownload: 0,
            } as any),
        );
        setModelListData(dataWithSource);
        setPagination({ ...pagination, total: dataWithSource.length });
      },
      onError: (error) => {
        console.error('获取模型列表失败:', error);
      },
    },
  );

  // 获取模型存储路径
  const { run: fetchModelPath } = useRequest(
    async () => {
      const res = await httpRequest.get<IModelPathRes>('/control_panel/model/filepath');
      return res || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        setModelPath(data?.path || '');
      },
      onError: (error) => {
        console.error('获取模型存储路径失败:', error);
      },
    },
  );

  useEffect(() => {
    fetchModelPath();
  }, []);

  useEffect(() => {
    if (modelSourceVal) {
      onModelSearch('');
    }
    setPagination({ ...pagination, current: 1 });

    fetchModelSupport({ service_source: modelSourceVal });
  }, [modelSourceVal]);

  // 根据搜索值和分页参数更新分页数据
  useEffect(() => {
    const filteredData = getFilteredData();
    setPagination((prev) => ({ ...prev, total: filteredData.length }));
    setPagenationData(paginatedData(pagination, filteredData));
  }, [modelListData, modelSearchVal, pagination.current, pagination.pageSize]);

  // 删除模型
  const { loading: deleteModelLoading, run: fetchDeleteModel } = useRequest(
    async (params: IRequestModelParams) => {
      await httpRequest.del('/model', params);
    },
    {
      manual: true,
      onSuccess: (data) => {
        message.success('模型删除成功');
      },
      onError: (error) => {
        message.error('模型删除失败');
        console.error('删除模型失败:', error);
      },
      onFinally: () => {
        fetchModelSupport({ service_source: modelSourceVal });
      },
    },
  );

  // 获取过滤后的数据
  const getFilteredData = () => {
    if (!modelSearchVal || modelSearchVal.trim() === '') {
      return modelListData;
    }
    return modelListData.filter((model) => model.name && model.name.toLowerCase().includes(modelSearchVal.toLowerCase()));
  };

  // 添加分页数据计算逻辑
  const paginatedData = (pagination: IPagenation, data = getFilteredData()) => {
    const { current, pageSize } = pagination;
    const startIndex = (current - 1) * pageSize;
    const endIndex = current * pageSize;
    return data.slice(startIndex, endIndex);
  };

  const onPageChange = (current: number) => {
    // 如果 pageSize 刚刚被改变，则不执行页码变更逻辑
    if (lastPageSizeRef.current !== pagination.pageSize) {
      lastPageSizeRef.current = pagination.pageSize;
      return;
    }
    setPagination({ ...pagination, current });
  };

  const onShowSizeChange = (current: number, pageSize: number) => {
    lastPageSizeRef.current = pageSize;
    setPagination({ ...pagination, current: 1, pageSize });
  };

  // 模型详情弹窗
  const onDetailModalVisible = useCallback((visible: boolean, modelData?: IModelDataItem) => {
    setIsDetailVisible(visible);
    // 弹窗关闭清空选择的模型数据
    setSelectModelData(modelData || ({} as any));
  }, []);

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
    confirm({
      title: '确认下载此模型？',
      okText: '确认下载',
      centered: true,
      okButtonProps: {
        style: { backgroundColor: '#5429ff' },
      },
      onOk() {
        fetchDownloadStart({
          ...modelData,
          type: modelData.type,
          status: DOWNLOAD_STATUS.IN_PROGRESS,
          modelType: 'local',
        });
      },
      onCancel() {
        console.log('Cancel');
      },
    });
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
        // 清空搜索框
        if (modelSourceVal) {
          onModelSearch('');
        }

        // 组装请求参数
        const params = {
          model_name: modelData.name,
          service_name: modelData.service_name || 'chat',
          service_source: modelData.source || 'local',
          provider_name: modelData.service_provider_name || 'local_ollama_chat',
        };
        fetchDeleteModel(params);
        console.log('OK');
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  };

  const fetchModalPath = () => {
    setModelPath('URL_ADDRESS.baidu.com');
  };

  // 授权成功刷新列表
  const onModelAuthSuccess = () => {
    fetchModelSupport({ service_source: modelSourceVal });
  };

  return {
    modelPath,
    modalPathVisible,
    onModelPathVisible,

    fetchModalPath,

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
    modelListData,
    modelSearchVal,
    modelSourceVal,
    onModelSearch,
    selectModelData,

    pagination,
    onPageChange,
    onShowSizeChange,
  };
}
