import { useState, useEffect } from 'react';
import { Modal, message } from 'antd';
import { useRequest } from 'ahooks';
import { httpRequest } from '../../utils/httpRequest';
import { IServiceProviderDataItem } from './types';

const { confirm } = Modal;

export function useViewModel() {
  const [dataList, setDataList] = useState<IServiceProviderDataItem[]>([]);
  const [selectedRow, setSelectedRow] = useState<IServiceProviderDataItem>({} as IServiceProviderDataItem);
  const [detailVisible, setDetailVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  useEffect(() => {
    fetchServiceProviders();
  }, []);

  const { loading: serviceProviderLoading, run: fetchServiceProviders } = useRequest(
    async () => {
      const data = await httpRequest.get<IServiceProviderDataItem[]>('/service_provider');
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        if (!data) return;
        setDataList(data);
        setPagination({ ...pagination, total: data.length });
      },
      onError: (error) => {
        console.error('获取服务提供商数据失败:', error);
      },
    },
  );

  const handlePageChange = (current: number) => {
    setPagination({ ...pagination, current });
  };

  const handleDetail = (rowData: IServiceProviderDataItem) => {
    setSelectedRow(rowData);
    setDetailVisible(true);
  };

  const { loading: deleteProviderLoading, run: fetchDeleteProvider } = useRequest(
    async (params: { provider_name: string }) => {
      const data = await httpRequest.del('/service_provider', params);
      return data;
    },
    {
      manual: true,
      onSuccess: (data) => {
        message.success('删除服务商成功');
        fetchServiceProviders();
      },
      onError: (error: Error & { handled?: boolean }) => {
        if (!error?.handled) {
          message.error('删除服务商失败，请重试');
        }
        console.error('删除服务商失败:', error);
      },
    },
  );

  const handleDeleteConfirm = (record: IServiceProviderDataItem) => {
    confirm({
      centered: true,
      title: '删除此服务提供商',
      content: '删除后，将同步卸载该服务提供商关联的所有模型，这可能会影响您的业务使用，确认继续吗？',
      okButtonProps: {
        style: { backgroundColor: '#e85951' },
        loading: deleteProviderLoading,
      },
      okText: '确认删除',
      onOk() {
        fetchDeleteProvider({ provider_name: record.provider_name });
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  };

  return { editVisible, dataList, detailVisible, pagination, serviceProviderLoading, setEditVisible, setDetailVisible, handlePageChange, handleDetail, handleDeleteConfirm, selectedRow };
}
