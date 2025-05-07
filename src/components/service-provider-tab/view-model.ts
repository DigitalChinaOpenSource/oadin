import { useState } from 'react';
import { Modal } from 'antd';
const { confirm } = Modal;

export function useViewModel() {
  const dataSource = Array.from({ length: 46 }).map((_, i) => ({
    index: i,
    id: i,
    key: i,
    name: `Edward King ${i}`,
    age: 32,
    address: `London, Park Lane no. ${i}`,
  }));

  const [dataList, setDataList] = useState(dataSource);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [detailVisible, setDetailVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [selectId, setSelectId] = useState<string>('');
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  const handlePageChange = (current: any) => {
    console.log('---page', current);
    setPagination({ ...pagination, current });
  };

  const handleDetail = (id: string) => {
    console.log('查看详情', id);
    setSelectId(id);
    setDetailVisible(true);
  };

  const columns = [
    { title: '序号', dataIndex: 'index' },
    { title: '服务提供商名称', dataIndex: 'name' },
    { title: '服务名称', dataIndex: 'address' },
    { title: '服务来源', dataIndex: 'address' },
    { title: '模型数量', dataIndex: 'address' },
  ];

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

  return { selectId, editVisible, dataList, detailVisible, columns, pagination, rowSelection, setEditVisible, setDetailVisible, handlePageChange, handleDetail, handleDeleteConfirm };
}
