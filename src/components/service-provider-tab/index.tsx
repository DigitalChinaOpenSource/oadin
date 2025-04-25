import styles from './index.module.scss';
import { Button, Table, Space } from 'antd';
import { PlusSquareOutlined } from '@ant-design/icons';
import { useViewModel } from './view-model';
import ServiceProviderDetail from './service-provider-detail';
import ServiceProviderEdit from './service-provider-edit';




export default function ServiceProviderManageTab() {
  const vm = useViewModel();

  const columnsList = [
    ...vm.columns,
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 200,
      render: (_, record) => (
        <Space>
          <a className={styles.linkA} onClick={() => vm.handleDetail(record.id)} >查看详情</a>

          <a className={styles.linkA} onClick={() => vm.setEditVisible(true)}>编辑</a>
        </Space>
      )
    }
  ]

  return (
    <div className={styles.container}>
      <div className={styles.titleBlock}>
        <span>服务提供商管理</span>
        <span>(12)</span>
        <div className={styles.btnBlock}>
          <Button type="primary" onClick={() => vm.setEditVisible(true)}  icon={<PlusSquareOutlined />}>新增服务提供商</Button>
          <Button onClick={vm.handleDeleteAll} className={styles.deleteBtn}>批量删除</Button>
        </div>
      </div>

      <Table
        rowSelection={vm.rowSelection}
        columns={columnsList}
        dataSource={vm.dataList}
        pagination={{
          ...vm.pagination,
          onChange: vm.handlePageChange
        }}  />

      <ServiceProviderDetail
         visible={vm.detailVisible}
         onCancel={() => vm.setDetailVisible(false)}
         onSubmit={() => {}}
       />


      <ServiceProviderEdit
         visible={vm.editVisible}
         onCancel={() => vm.setEditVisible(false)}
         id={vm.editId}
      />
    </div>
  );
};