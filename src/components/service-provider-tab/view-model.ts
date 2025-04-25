import { useState } from "react";

export function useViewModel() {
    const dataSource = Array.from({ length: 46 }).map((_, i) => ({
        index: i,
        id: i,
        key: i,
        name: `Edward King ${i}`,
        age: 32,
        address: `London, Park Lane no. ${i}`,
      }));

    const [dataList, setDataList] = useState(dataSource)
    const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
    const [detailVisible, setDetailVisible] = useState(false);
    const [editVisible, setEditVisible] = useState(false);
    const [editId, setEditId] = useState<string>('');
    const [pagination, setPagination] = useState({
        current: 1,
        pageSize: 10,
        total: 0
      });

      const handlePageChange = (current: any) => {
        console.log('---page', current)
        setPagination({...pagination, current})
      };
    

    // 处理下载列表和进度相关内容，用于useContext分发

    const handleEdit = (record: any) => {
        console.log('编辑', record);
        setDetailVisible(true);
    }

    const handleDetail = (id: string) => {
        console.log('删除', id);
        setDetailVisible(true);
    }

    const handleDeleteAll = () => {
        console.log('批量删除', selectedRowKeys);
    }

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
            setSelectedRowKeys(keys)
        },
    };
    return { editId, editVisible, dataList, detailVisible, columns, pagination, rowSelection, setEditVisible, setDetailVisible, handlePageChange, handleEdit, handleDeleteAll, handleDetail };
}