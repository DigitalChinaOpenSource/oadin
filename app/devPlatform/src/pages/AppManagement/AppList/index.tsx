/* eslint-disable no-promise-executor-return */
import React, { useState } from 'react';
import { Button, Input, message, Pagination, Spin } from 'antd';
import AppCard from './AppCard.tsx';
import { useRequest } from 'ahooks';
import { PlusOutlined } from '@ant-design/icons';
import CreateAppModal from './CreateAppModal.tsx';
import DeleteAppModal from './DeleteAppModal.tsx';
import styles from './index.module.scss';
import { addApplication, deleteApplication, getApplicationList } from '@/pages/AppManagement/remote';
import { NoData } from '@/pages/AppManagement/AppList/noData.tsx';

// 删除应用
const deleteApp = async (id: string) => {
  return await deleteApplication(id);
};
// 创建应用
const createApp = async (values: { name: string }) => {
  const data = await addApplication(values);
  console.info(data, '创建应用返回数据');
  return data;
};

// 查询列表数据
const fetchApps = async (params: { current: number; pageSize: number; keyword?: string }) => {
  const { current, pageSize, keyword = '' } = params;
  console.info(keyword, '搜索数据');
  return getApplicationList({
    page: current,
    size: pageSize,
    keyword,
  }).then((res) => {
    console.log('获取应用列表结果:', res);
    const filteredApps = res.list;

    return {
      data: filteredApps,
      total: res.total,
    };
  });
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
        keyword: searchText,
      }),
    {
      refreshDeps: [current, searchText],
    },
  );
  const [deletingApp, setDeletingApp] = useState<any>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [formModalVisible, setFormModalVisible] = useState(false);
  const [formMode, setFormMode] = useState<'create'>('create');
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
    setFormModalVisible(true);
  };
  const handleFormSubmit = async (values: { name: string }): Promise<void> => {
    setFormLoading(true);
    try {
      if (formMode === 'create') {
        const isCreated = await createApp(values);
        if (isCreated) {
          message.success('创建成功');
          refresh();
          setFormModalVisible(false);
        }
      }
    } finally {
      setFormLoading(false);
    }
  };

  const handleFormCancel = () => {
    setFormModalVisible(false);
    setFormMode('create');
  };

  const handleSearch = (value: string) => {
    console.info(value);
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
                  <span className={styles['total-count']}>（共 {data?.total || 0} 个）</span>
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
            onSearch={(e) => handleSearch(e.trim())}
          />
        </div>

        <Spin spinning={listLoading}>
          <div className={styles['app-list-container']}>
            {data && data.total > 0 ? (
              data?.data.map((app: any) => (
                <AppCard
                  key={app.id}
                  app={app}
                  onDelete={showDeleteModal}
                  onCopy={handleCopy}
                />
              ))
            ) : (
              <NoData showCreateModal={showCreateModal} />
            )}
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
