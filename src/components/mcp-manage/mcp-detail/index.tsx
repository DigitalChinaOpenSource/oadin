import Styles from './index.module.scss';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { Button, Tabs } from 'antd';
import { useState } from 'react';
import type { TabsProps } from 'antd';
import McpOverview from '@/components/mcp-manage/mcp-detail/mcp-overview';
import McpTools from '@/components/mcp-manage/mcp-detail/mcp-tools';
import McpInstructions from '@/components/mcp-manage/mcp-detail/mcp-instructions';
import DetailDesc from '@/components/mcp-manage/mcp-detail/detail-desc';
import McpServiceConfig from '@/components/mcp-manage/mcp-detail/mcp-service-config';
import RecommendedClient from '@/components/mcp-manage/mcp-detail/recommended-client';

export default function McpDetail() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const serviceId = searchParams.get('serviceId');
  const mcpFrom = searchParams.get('mcpFrom');
  const [loading, setLoading] = useState<boolean>(false);
  const items: TabsProps['items'] = [
    {
      key: 'overView',
      label: '概览',
      children: <McpOverview />,
    },
    {
      key: 'tools',
      label: '工具',
      children: <McpTools />,
    },
    {
      key: 'presetInstructions',
      label: '预设指令',
      children: <McpInstructions />,
    },
  ];
  const handledGoBack = (): void => {
    navigate(`/mcp-service?mcpFrom=${mcpFrom}`);
  };
  return (
    <div className={Styles.mcpManageDetail}>
      <div
        className={Styles.goBack}
        onClick={handledGoBack}
      >
        <ArrowLeftOutlined className={Styles.backIcon} />
        <span className={Styles.backText}>返回</span>
      </div>
      <div className={Styles.detailTop}>
        <div className={Styles.topLeft}>
          <DetailDesc />
        </div>
        <div className={Styles.topRight}>
          <Button type="primary">添加</Button>
        </div>
      </div>
      {/*分割线*/}
      <div className={Styles.Line}></div>
      <div className={Styles.detailContent}>
        <div className={Styles.contentLeft}>
          <Tabs
            className={Styles.tabs}
            defaultActiveKey="overView"
            items={items}
            // onChange={onChange}
          />
        </div>
        <div className={Styles.contentRight}>
          <McpServiceConfig />
          <RecommendedClient />
        </div>
      </div>
    </div>
  );
}
