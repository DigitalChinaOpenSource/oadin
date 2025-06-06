import { Table } from 'antd';
import { useViewModel } from './view-model';
import ServiceProviderDetail from './service-provider-detail';
import { useTableColumns } from './table-columns';
import noDataSvg from '@/components/icons/no-data.svg';
import styles from './index.module.scss';
import { useEffect, useRef, useState } from 'react';

export default function ServiceProviderManageTab() {
  const vm = useViewModel();
  const columns = useTableColumns({ handleDetail: vm.handleDetail });
  const tableRef = useRef<HTMLDivElement>(null);
  const [scrollConfig, setScrollConfig] = useState<{ y: number | undefined }>();
  // const [tableHeight, setTableHeight] = useState(0);
  const calculateScrollConfig = () => {
    if (!tableRef.current) return;

    // 获取表格容器元素
    const tableContainer = tableRef.current.parentElement;
    if (!tableContainer) return;

    // 计算视窗高度和表格可用高度
    const windowHeight = window.innerHeight;
    console.log('Window Height:', windowHeight);
    const containerRect = tableContainer.getBoundingClientRect();
    console.log('Container Rect:', containerRect);
    const availableHeight = windowHeight - containerRect.top - 48 - 40 - 64 - 55; // 减去间距
    console.log('Available Height:', availableHeight);

    // 获取表格内容实际高度 (不包括表头)
    const tableBody = tableContainer.querySelector('.ant-table-tbody');
    console.log('Table Body:', tableBody);
    const contentHeight = tableBody ? tableBody.scrollHeight : 0;
    console.log('Content Height:', contentHeight);

    // setTableHeight(contentHeight);

    // 如果内容高度超过可用高度，设置 scroll.y
    if (contentHeight > availableHeight) {
      setScrollConfig({ y: availableHeight });
    } else {
      setScrollConfig({ y: undefined }); // 内容不足时不固定高度
    }
  };
  // 计算表格内容高度和可用视窗高度
  useEffect(() => {
    // 初始计算
    calculateScrollConfig();

    // 监听窗口大小变化
    window.addEventListener('resize', calculateScrollConfig);

    // 清理监听器
    return () => {
      window.removeEventListener('resize', calculateScrollConfig);
    };
  }, [vm.pagination]);

  return (
    <div className={styles.serviceProviderTab}>
      <div className={styles.container}>
        <div className={styles.titleBlock}>服务提供商管理</div>

        <Table
          ref={tableRef as any}
          loading={vm.serviceProviderLoading}
          columns={columns}
          scroll={scrollConfig} // 动态设置 scroll 属性
          dataSource={vm.dataList}
          pagination={
            vm.dataList.length > 10 && {
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
