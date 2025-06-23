/* eslint-disable no-promise-executor-return */
import React, { useState } from 'react';
import { Button, Input, message, Pagination, Spin } from 'antd';
import AppCard from './AppCard';
import { useRequest } from 'ahooks';
import { PlusOutlined } from '@ant-design/icons';
import CreateAppModal from './CreateAppModal';
import DeleteAppModal from './DeleteAppModal.tsx';
import styles from './index.module.scss';

// 模拟删除接口
const deleteApp = async (id: string) => {
  console.log('Deleting app:', id);
  await new Promise((resolve) => setTimeout(resolve, 2000));
  return { success: true };
};

// 模拟创建应用接口
const createApp = async (values: { name: string }) => {
  await new Promise((resolve) => setTimeout(resolve, 2000));
  return {
    success: true,
    data: {
      id: `app-${Math.random().toString(36).substring(2, 10)}`,
      name: values.name,
      appId: `app_${Math.random().toString(36).substring(2, 10)}`,
      secretKey: `sk_${Math.random().toString(36).substring(2, 15)}`,
      modelCount: 0,
      mcpCount: 0,
      osCount: 0,
      updatedAt: new Date().toISOString(),
    },
  };
};

// 模拟更新应用接口
const updateApp = async (id: string, values: { name: string }) => {
  await new Promise((resolve) => setTimeout(resolve, 2000));
  return {
    success: true,
    data: {
      id,
      name: values.name,
      updatedAt: new Date().toISOString(),
    },
  };
};

// 模拟数据
const mockApps = [
  // 超长名称测试数据
  {
    id: 'long-1',
    name: '这是一个非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常非常长的应用名称1',
    appId: 'long1',
    secretKey: 'sk_long1',
    modelCount: 2,
    mcpCount: 1,
    osCount: 1,
    updatedAt: new Date().toISOString(),
  },
  // 其余 mock 数据
  ...Array.from({ length: 23 }).map((_, index) => ({
    id: `app-${index + 1}`,
    name: `测试应用${index + 1}`,
    appId: `${Math.random().toString(36).substring(2, 10)}`,
    secretKey: `sk_${Math.random().toString(36).substring(2, 15)}`,
    modelCount: Math.floor(Math.random() * 5) + 1,
    mcpCount: Math.floor(Math.random() * 3) + 1,
    osCount: Math.floor(Math.random() * 4) + 1,
    updatedAt: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString(),
  })),
];

// 模拟分页获取数据
const fetchApps = async (params: { current: number; pageSize: number; search?: string }) => {
  await new Promise((resolve) => setTimeout(resolve, 1000));
  const { current, pageSize, search = '' } = params;
  const filteredApps = mockApps.filter((app) => app.name.toLowerCase().includes(search.toLowerCase()));

  const start = (current - 1) * pageSize;
  const end = start + pageSize;

  return {
    data: filteredApps.slice(start, end),
    total: filteredApps.length,
    totalAll: mockApps.length,
  };
};

const AppManage: React.FC = () => {
  const [searchText, setSearchText] = useState('');
  const [current, setCurrent] = useState(1);
  const pageSize = 10;

  const {
    data,
    refresh,
    loading: listLoading,
  } = useRequest(
    () =>
      fetchApps({
        current,
        pageSize,
        search: searchText,
      }),
    {
      refreshDeps: [current, searchText],
    },
  );
  const [deletingApp, setDeletingApp] = useState<any>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [formModalVisible, setFormModalVisible] = useState(false);
  const [editingApp, setEditingApp] = useState<any>(null);
  const [formMode, setFormMode] = useState<'create' | 'edit'>('create');
  const [formLoading, setFormLoading] = useState(false);

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text);
    message.success('已复制');
  };

  const handleDelete = async () => {
    if (!deletingApp) return;

    setDeleteLoading(true);
    try {
      await deleteApp(deletingApp.id);
      message.success('删除成功');
      refresh();
    } catch (error) {
      message.error('删除失败');
    } finally {
      setDeleteLoading(false);
      setDeletingApp(null);
    }
  };

  const showDeleteModal = (app: any) => {
    setDeletingApp(app);
  };

  const showCreateModal = () => {
    setFormMode('create');
    setEditingApp(null);
    setFormModalVisible(true);
  };

  const showEditModal = (app: any) => {
    setFormMode('edit');
    setEditingApp(app);
    setFormModalVisible(true);
  };
  const handleFormSubmit = async (values: { name: string }): Promise<void> => {
    setFormLoading(true);
    try {
      if (formMode === 'create') {
        const res = await createApp(values);
        if (res.success) {
          message.success('创建成功');
          refresh();
          setFormModalVisible(false);
        }
      } else {
        const res = await updateApp(editingApp.id, values);
        if (res.success) {
          message.success('更新成功');
          refresh();
          setFormModalVisible(false);
        }
      }
    } catch (error) {
      message.error(formMode === 'create' ? '创建失败' : '更新失败');
    } finally {
      setFormLoading(false);
    }
  };

  const handleFormCancel = () => {
    setFormModalVisible(false);
    setEditingApp(null);
    setFormMode('create');
  };

  const handleSearch = (value: string) => {
    setSearchText(value);
    setCurrent(1); // 重置到第一页
  };

  return (
    <div className={styles['page-container']}>
      <div className={styles['content-container']}>
        <div className={styles['header-row']}>
          <h1 className={styles['page-title']}>应用管理</h1>
        </div>

        <div className={styles['action-row']}>
          <div className={styles['left-actions']}>
            <Button
              className={styles['create-btn']}
              type="primary"
              icon={<PlusOutlined />}
              onClick={showCreateModal}
            >
              创建应用
            </Button>
            <span className={styles['result-count']}>
              {searchText ? (
                <>
                  找到 <strong className={styles.highlight}>{data?.total || 0}</strong> 个应用
                  <span className={styles['total-count']}>（共 {data?.totalAll || 0} 个）</span>
                </>
              ) : (
                `已创建 ${data?.total || 0} 个应用`
              )}
            </span>
          </div>
          <Input.Search
            className={styles['search-input']}
            placeholder="搜索应用"
            allowClear
            value={searchText}
            onChange={(e) => handleSearch(e.target.value)}
          />
        </div>

        <Spin spinning={listLoading}>
          <div className={styles['app-list-container']}>
            {data?.data.map((app: any) => (
              <AppCard
                key={app.id}
                app={app}
                onEdit={showEditModal}
                onDelete={showDeleteModal}
                onCopy={handleCopy}
              />
            ))}
          </div>
          {data && data.total > 0 && (
            <div className={styles['pagination-container']}>
              <Pagination
                current={current}
                pageSize={pageSize}
                total={data.total}
                onChange={(page) => setCurrent(page)}
                showTotal={(total) => `共 ${total} 条`}
              />
            </div>
          )}
        </Spin>
      </div>
      <CreateAppModal
        open={formModalVisible}
        onCancel={handleFormCancel}
        onFinish={handleFormSubmit}
        confirmLoading={formLoading}
      />
      <DeleteAppModal
        deleteLoading={deleteLoading}
        deleteApp={deletingApp}
        setDeletingApp={setDeletingApp}
        onDelete={handleDelete}
      />
    </div>
  );
};

export default AppManage;
