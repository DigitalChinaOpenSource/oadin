import { Divider, Drawer, DrawerProps, Tabs, TabsProps } from 'antd';
import React, { useEffect } from 'react';
import { useMcpDetail } from './useMcpDetail.ts';
import styles from './index.module.scss';
import McpOverview from './mcp-overview';
import McpTools from './mcp-tools';
import DrawerDetailTitle from './title.tsx';

export interface IDetailDrawer extends DrawerProps {
  id: string;
}

export const DetailDrawer: React.FC<IDetailDrawer> = (options: IDetailDrawer) => {
  const { id } = options;

  const { fetchMcpDetail, mcpDetail } = useMcpDetail(id);
  useEffect(() => {
    if (id) {
      fetchMcpDetail();
    }
  }, [id]);

  const items: TabsProps['items'] = [
    {
      key: 'overView',
      label: '概览',
      children: <McpOverview markDownData={mcpDetail?.summary} />,
    },
    {
      key: 'tools',
      label: '工具',
      children: (
        <McpTools
          status={mcpDetail?.status}
          id={id}
        />
      ),
    },
  ];
  return (
    <Drawer
      size="large"
      zIndex={1500}
      closable={false}
      {...options}
      destroyOnHidden={true}
    >
      {mcpDetail && (
        <div className={styles.mcpManageDetail}>
          <DrawerDetailTitle
            {...options}
            mcpDetail={mcpDetail}
          />
          <Divider />
          <div className={styles.detailContent}>
            <Tabs
              className={styles.tabs}
              defaultActiveKey="tools"
              items={items}
            />
          </div>
        </div>
      )}
    </Drawer>
  );
};
