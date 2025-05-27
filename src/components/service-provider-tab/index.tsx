import { Table } from 'antd';
import { useViewModel } from './view-model';
import ServiceProviderDetail from './service-provider-detail';
import { useTableColumns } from './table-columns';
import noDataSvg from '@/components/icons/no-data.svg';
import styles from './index.module.scss';

export default function ServiceProviderManageTab() {
  const vm = useViewModel();
  const columns = useTableColumns({ handleDetail: vm.handleDetail });

  return (
    <div className={styles.serviceProviderTab}>
      <div className={styles.container}>
        <div className={styles.titleBlock}>服务提供商管理</div>

        <Table
          loading={vm.serviceProviderLoading}
          columns={columns}
          dataSource={vm.dataList}
          pagination={
            vm.pagination.total > 10 && {
              ...vm.pagination,
              onChange: vm.handlePageChange,
            }
          }
          locale={{
            emptyText: () => {
              return (
                <>
                  {!vm.serviceProviderLoading && (
                    <div className={styles.noData}>
                      <div className={styles.noDataIcon}>
                        <img
                          src={noDataSvg}
                          alt="no-data"
                        />
                      </div>
                      <div className={styles.noDataText}>暂无匹配的模型</div>
                    </div>
                  )}
                </>
              );
            },
          }}
        />
      </div>
      {vm.detailVisible && (
        <ServiceProviderDetail
          selectedRow={vm.selectedRow}
          onCancel={() => vm.setDetailVisible(false)}
        />
      )}
    </div>
  );
}
