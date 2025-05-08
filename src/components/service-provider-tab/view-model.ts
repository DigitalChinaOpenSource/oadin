import { useState, useEffect } from 'react';
import { Modal } from 'antd';
import { useAxios } from '../../utils/useAxios';

const { confirm } = Modal;

export function useViewModel() {
  // 移除了mock数据

  const { get } = useAxios();
  const [dataList, setDataList] = useState([]);
  const [loading, setLoading] = useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [detailVisible, setDetailVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [selectId, setSelectId] = useState<string>('');
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  useEffect(() => {
    fetchServiceProviders();
  }, []);

  // 获取服务提供商数据
  const fetchServiceProviders = async () => {
    setLoading(true);
    try {
      const data = await get('/service_provider');
      if (data) {
        setDataList(data);
        setPagination({ ...pagination, total: data.length });
      }
      setLoading(false);
    } catch (error) {
      setLoading(false);
      console.error('获取服务提供商数据失败:', error);
    }
  };

  const handlePageChange = (current: any) => {
    console.log('---page', current);
    setPagination({ ...pagination, current });
  };

  const handleDetail = (id: string) => {
    console.log('查看详情', id);
    setSelectId(id);
    setDetailVisible(true);
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: (keys: React.Key[]) => {
      setSelectedRowKeys(keys);
    },
  };

  const handleDeleteConfirm = (record: any) => {
    confirm({
      centered: true,
      title: '删除此服务提供商',
      content: '删除后，将同步卸载该服务提供商关联的所有模型，这可能会影响您的业务使用，确认继续吗？',
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

  return { selectId, editVisible, dataList, detailVisible, pagination, rowSelection, loading, setEditVisible, setDetailVisible, handlePageChange, handleDetail, handleDeleteConfirm };
}
