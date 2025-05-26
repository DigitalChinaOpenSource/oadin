import { Table, Space } from 'antd';
import styles from './index.module.scss';
import { IServiceProviderDataItem } from './types';
export interface ITableColumns {
  handleDetail: (rowData: IServiceProviderDataItem) => void;
}

export function useTableColumns(props: ITableColumns) {
  const { handleDetail } = props;

  const serviceSourceEnum = {
    remote: '云端模型',
    local: '本地模型',
  } as any;

  const serviceNameEnum = {
    chat: '会话',
    embed: '词嵌入',
    text_to_image: '文生图',
    generate: '文本生成',
  } as any;

  return [
    { title: '服务提供商名称', dataIndex: 'provider_name', key: 'provider_name' },
    { title: '服务名称', dataIndex: 'service_name', key: 'service_name', render: (name: string) => <>{serviceNameEnum[name]}</> },
    { title: '服务来源', dataIndex: 'service_source', key: 'service_source', render: (source: string) => <>{serviceSourceEnum[source]}</> },
    { title: '模型数量', dataIndex: 'models', key: 'models', render: (models: string[]) => <>{models?.length || 0}</> },
    {
      title: '操作',
      key: 'action',
      fixed: 'right',
      width: 200,
      render: (_: any, record: any) => (
        <Space>
          <a
            className={styles.linkA}
            onClick={() => handleDetail(record)}
          >
            查看详情
          </a>
        </Space>
      ),
    },
  ] as any;
}
