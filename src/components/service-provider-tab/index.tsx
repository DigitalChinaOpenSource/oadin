import styles from './index.module.scss';
import { Table, Space, Modal } from 'antd';
import { useViewModel } from './view-model';
import ServiceProviderDetail from './service-provider-detail';

const { confirm } = Modal;

export default function ServiceProviderManageTab() {
  const vm = useViewModel();

  const columnsList = [
    ...vm.columns,
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 200,
      render: (_: any, record: any) => (
        <Space>
          <a
            className={styles.linkA}
            onClick={() => vm.handleDetail(record.id)}
          >
            查看详情
          </a>
          <a
            className={styles.linkA}
            onClick={() => vm.handleDeleteConfirm(record)}
          >
            删除
          </a>
        </Space>
      ),
    },
  ];

  return (
    <div className={styles.serviceProviderTab}>
      <div className={styles.container}>
        <div className={styles.titleBlock}>服务提供商管理</div>

        <Table
          rowSelection={vm.rowSelection}
          columns={columnsList as any}
          dataSource={vm.dataList}
          pagination={{
            ...vm.pagination,
            onChange: vm.handlePageChange,
          }}
        />
      </div>
      {vm.detailVisible && (
        <ServiceProviderDetail
          id={vm.selectId}
          onCancel={() => vm.setDetailVisible(false)}
        />
      )}
    </div>
  );
}
