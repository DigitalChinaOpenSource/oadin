import styles from './index.module.scss';
import { Table, Space, Modal } from 'antd';
import { useViewModel } from './view-model';
// import ServiceProviderDetail from './service-provider-detail';
// import ServiceProviderEdit from './service-provider-edit';

const { confirm } = Modal;

export default function ServiceProviderManageTab() {
  const vm = useViewModel();

  const deleteConfirm = (record: any) => {
    confirm({
      title: '删除此服务提供商',
      content:
        '删除后，将同步卸载该服务提供商关联的所有模型，这可能会影响您的业务使用，确认继续吗？',
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

          {/* <a className={styles.linkA} onClick={() => vm.setEditVisible(true)}>
            编辑
          </a> */}
          <a className={styles.linkA} onClick={() => deleteConfirm(record)}>
            删除
          </a>
        </Space>
      ),
    },
  ];

  return (
    <div className={styles.serviceProviderTab}>
      <div className={styles.container}>
        <div className={styles.titleBlock}>
          服务提供商管理
          {/* <div className={styles.btnBlock}>
            <Button type="primary" onClick={() => vm.setEditVisible(true)}  icon={<PlusSquareOutlined />}>新增服务提供商</Button>
            <Button onClick={vm.handleDeleteAll} className={styles.deleteBtn}>
              批量删除
            </Button>
          </div> */}
        </div>

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

      {/* <ServiceProviderDetail
        visible={vm.detailVisible}
        onCancel={() => vm.setDetailVisible(false)}
        onSubmit={() => {}}
      />

      <ServiceProviderEdit
        visible={vm.editVisible}
        onCancel={() => vm.setEditVisible(false)}
        id={vm.editId}
      /> */}
    </div>
  );
}
