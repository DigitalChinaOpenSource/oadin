import { DrawerProps, Drawer, Tabs, TabsProps, Divider } from 'antd';
import React, { useEffect } from 'react';
import { useMcpDetail } from '@/components/mcp-manage/mcp-detail/useMcpDetail.ts';
import styles from './index.module.scss';
import McpOverview from '@/components/mcp-manage/mcp-detail/mcp-overview';
import McpTools from '@/components/mcp-manage/mcp-detail/mcp-tools';
import DrawerDetailTitle from '@/components/detail_drawer/title.tsx';

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
    // {
    //   key: 'presetInstructions',
    //   label: '预设指令',
    //   children: <McpInstructions />,
    // },
  ];
  console.info(mcpDetail, '详情数据');
  return (
    <Drawer
      size="large"
      zIndex={1500}
      closable={false}
      open={true}
      {...options}
    >
      {mcpDetail && (
        <div className={styles.mcpManageDetail}>
          <DrawerDetailTitle mcpDetail={mcpDetail} />
          <Divider />
          <div className={styles.detailContent}>
            <Tabs
              className={styles.tabs}
              defaultActiveKey="tools"
              items={items}
              // onChange={onChange}
            />
          </div>
        </div>
      )}
    </Drawer>
  );
};
