import styles from './index.module.scss';
import { Table } from 'antd';
import { useViewModel } from './view-model';
import ServiceProviderDetail from './service-provider-detail';
import { useTableColumns } from './table-columns';

export default function ServiceProviderManageTab() {
  const vm = useViewModel();
  const columns = useTableColumns({ handleDetail: vm.handleDetail, handleDeleteConfirm: vm.handleDeleteConfirm });

  const pagination =
    vm.dataList.length > 10
      ? {
          ...vm.pagination,
          onChange: vm.handlePageChange,
        }
      : undefined;

  return (
    <div className={styles.serviceProviderTab}>
      <div className={styles.container}>
        <div className={styles.titleBlock}>服务提供商管理</div>

        <Table
          loading={vm.loading}
          columns={columns}
          dataSource={vm.dataList}
          pagination={pagination}
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
