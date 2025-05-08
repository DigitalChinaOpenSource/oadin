import { useState, useCallback, useEffect, useRef } from 'react';
import { IModelAuthorize, IModelAuthType, ModelData, ModelDataItem, IModelSourceType } from '../types';
import { useAxios } from '../../../utils/useAxios';
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
    pageSize: 10,
    total: 0,
  });

  const lastPageSizeRef = useRef(pagination.pageSize);

  const { get } = useAxios();

  const fetchModelList = async (serviceSource: IModelSourceType) => {
    const data = await get<ModelData>('/model/support', {
      service_source: serviceSource,
      flavor: 'ollama',
    });
    return data?.chat || [];
  };

  const { loading, run } = useRequest(fetchModelList, {
    manual: true,
    onSuccess: (data) => {
      const dataWithSource = data.map(
        (item) =>
          ({
            ...item,
            source: 'local',
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

  useEffect(() => {
    setPagenationData(paginatedData(pagination));
  }, [modelListData]);

  // 添加分页数据计算逻辑
  const paginatedData = (pagination: any) => {
    console.log('pagination====>', pagination);
    const { current, pageSize } = pagination;
    const startIndex = (current - 1) * pageSize;
    const endIndex = current * pageSize;
    return modelListData.slice(startIndex, endIndex);
  };

  const onPageChange = (current: number) => {
    // 如果 pageSize 刚刚被改变，则不执行页码变更逻辑
    if (lastPageSizeRef.current !== pagination.pageSize) {
      lastPageSizeRef.current = pagination.pageSize;
      return;
    }
    setPagenationData(paginatedData({ ...pagination, current }));
    setPagination({ ...pagination, current });
  };

  const onShowSizeChange = (current: number, pageSize: number) => {
    lastPageSizeRef.current = pageSize;
    setPagenationData(paginatedData({ ...pagination, current: 1, pageSize }));
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

  const downloadConfirm = (modelData: ModelDataItem) => {
    confirm({
      title: '确认下载此模型？',
      okText: '确认下载',
      onOk() {
        console.log('OK');
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  };

  const deleteConfirm = (modelData: ModelDataItem) => {
    confirm({
      title: '是否确认删除此模型',
      content: '此操作将删除您与该模型的所有对话记录、文件信息，并终止相关进程。',
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

    deleteConfirm,
    downloadConfirm,

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
