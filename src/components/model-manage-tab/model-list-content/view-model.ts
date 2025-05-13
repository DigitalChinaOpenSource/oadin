import { useState, useCallback, useEffect, useRef } from 'react';
import { IModelAuthorize, IModelAuthType, ModelData, ModelDataItem, IModelSourceType } from '../types';
import { DOWNLOAD_STATUS } from '../../../constants';
import { httpRequest } from '../../../utils/httpRequest';
import { useDownLoad } from '../../../hooks/useDownload';
import { Modal } from 'antd';
import { useRequest } from 'ahooks';

const { confirm } = Modal;

export interface IViewModel {
  modelSourceVal: IModelSourceType;
  modelSearchVal: string;
}

export function useViewModel(props: IViewModel) {
  const { modelSourceVal, modelSearchVal } = props;
  // 模型存储路径弹窗是否显示
  const [modalPathVisible, setModalPathVisible] = useState<boolean>(false);
  // 接口获取
  const [modelPath, setModelPath] = useState<string>('');
  // 云端模型授权弹窗是否显示
  const [modelAuthVisible, setModelAuthVisible] = useState<boolean>(false);
  // 云端模型授权
  const [modelAuthorize, setModelAuthorize] = useState<IModelAuthorize>({
    apiHost: '',
    apiKey: '',
  });
  // 模型详情弹窗是否显示
  const [isDetailVisible, setIsDetailVisible] = useState(false);
  // 配置 ｜ 更新授权
  const [modelAuthType, setModelAuthType] = useState<IModelAuthType>('config');
  const [modelListData, setModelListData] = useState<ModelDataItem[]>([]);
  const [pagenationData, setPagenationData] = useState<ModelDataItem[]>([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 6,
    total: 0,
  });

  const lastPageSizeRef = useRef(pagination.pageSize);
  const { downLoadStart } = useDownLoad();

  const fetchModelSupport = async (serviceSource: IModelSourceType) => {
    const data = await httpRequest.get<ModelData>('/model/support', {
      service_source: serviceSource,
      flavor: 'ollama',
    });
    return data?.chat || [];
  };

  const { loading, run } = useRequest(fetchModelSupport, {
    manual: true,
    onSuccess: (data) => {
      const dataWithSource = data.map(
        (item, index) =>
          ({
            ...item,
            source: 'local',
            type: 0,
            id: index + 1,
          } as any),
      );
      setModelListData(dataWithSource);
      setPagination({ ...pagination, total: dataWithSource.length });
      console.log('获取模型列表成功:', data);
    },
    onError: (error) => {
      console.error('获取模型列表失败:', error);
    },
  });

  const fetchModelListData = () => {
    if (modelSourceVal === 'local' || modelSourceVal === 'all') {
      run(modelSourceVal);
    }
  };

  useEffect(() => {
    fetchModelListData();
  }, [modelSourceVal]);

  // 根据搜索值和分页参数更新分页数据
  useEffect(() => {
    const filteredData = getFilteredData();
    setPagination((prev) => ({ ...prev, total: filteredData.length }));
    setPagenationData(paginatedData(pagination, filteredData));
  }, [modelListData, modelSearchVal, pagination.current, pagination.pageSize]);

  // 获取过滤后的数据
  const getFilteredData = () => {
    if (!modelSearchVal || modelSearchVal.trim() === '') {
      return modelListData;
    }
    return modelListData.filter((model) => model.name && model.name.toLowerCase().includes(modelSearchVal.toLowerCase()));
  };

  // 添加分页数据计算逻辑
  const paginatedData = (pagination: any, data = getFilteredData()) => {
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
  const onDetailModalVisible = useCallback((visible: boolean) => {
    setIsDetailVisible(visible);
  }, []);

  // 模型存储路径弹窗
  const onModelPathVisible = useCallback(() => {
    setModalPathVisible(!modalPathVisible);
  }, [modalPathVisible]);

  const fetchModalPath = () => {
    setModelPath('URL_ADDRESS.baidu.com');
  };

  const onModelAuthVisible = useCallback((visible: boolean, type: IModelAuthType) => {
    setModelAuthVisible(visible);
    setModelAuthType(type);
  }, []);

  const onSetModelAuthorize = (authData: IModelAuthorize) => {
    setModelAuthorize(authData);
  };

  const onDownloadConfirm = (modelData: ModelDataItem) => {
    confirm({
      title: '确认下载此模型？',
      okText: '确认下载',
      centered: true,
      onOk() {
        downLoadStart({
          ...modelData,
          modelName: modelData.name,
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

  const onDeleteConfirm = (modelData: ModelDataItem) => {
    confirm({
      title: '是否确认删除此模型',
      content: '此操作将删除您与该模型的所有对话记录、文件信息，并终止相关进程。',
      centered: true,
      okButtonProps: {
        style: { backgroundColor: '#e85951' },
      },
      okText: '确认删除',
      onOk() {
        console.log('OK');
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  };

  return {
    modelPath,
    modalPathVisible,
    onModelPathVisible,

    fetchModalPath,
    modelAuthorize,
    onSetModelAuthorize,

    modelAuthVisible,
    onModelAuthVisible,

    isDetailVisible,
    onDetailModalVisible,
    modelAuthType,

    onDeleteConfirm,
    onDownloadConfirm,

    loading,
    pagenationData, // 分页后的数据
    modelListData, // 完整数据
    modelSearchVal,
    modelSourceVal,

    pagination,
    onPageChange,
    onShowSizeChange,
  };
}
