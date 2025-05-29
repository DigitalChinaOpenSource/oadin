import styles from './index.module.scss';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { Button, Tabs } from 'antd';
import { useEffect, useRef, useState } from 'react';
import type { TabsProps } from 'antd';
import McpOverview from '@/components/mcp-manage/mcp-detail/mcp-overview';
import McpTools from '@/components/mcp-manage/mcp-detail/mcp-tools';
import McpInstructions from '@/components/mcp-manage/mcp-detail/mcp-instructions';
import DetailDesc from '@/components/mcp-manage/mcp-detail/detail-desc';
import McpServiceConfig from '@/components/mcp-manage/mcp-detail/mcp-service-config';
import RecommendedClient from '@/components/mcp-manage/mcp-detail/recommended-client';
import { useMcpDetail } from '@/components/mcp-manage/mcp-detail/useMcpDetail.ts';
import McpAuthModal from '@/components/mcp-manage/mcp-detail/mcp-auth-modal';

export default function McpDetail() {
  const { handleGoBack, mcpDetail, handleAddMcp, handleCancelMcp, cancelMcpLoading, downMcpLoading, authMcpLoading, handleAuthMcp, showMcpModal, setShowMcpModal } = useMcpDetail();

  const detailDescRef = useRef<HTMLDivElement>(null);
  const [descHeight, setDescHeight] = useState(0);

  const items: TabsProps['items'] = [
    {
      key: 'overView',
      label: '概览',
      children: (
        <McpOverview
          markDownData={mcpDetail?.summary}
          offHeight={descHeight}
        />
      ),
    },
    {
      key: 'tools',
      label: '工具',
      children: <McpTools status={mcpDetail?.status} />,
    },
    // {
    //   key: 'presetInstructions',
    //   label: '预设指令',
    //   children: <McpInstructions />,
    // },
  ];

  useEffect(() => {
    console.log('detailDescRef.current?.showTooltip', detailDescRef.current?.offsetHeight);
    if (detailDescRef.current) {
      const height = detailDescRef.current.offsetHeight;
      setDescHeight(height);
    }
  }, [mcpDetail]);

  return (
    mcpDetail && (
      <div className={styles.mcpManageDetail}>
        <div
          className={styles.goBack}
          onClick={handleGoBack}
        >
          <ArrowLeftOutlined className={styles.backIcon} />
          <span className={styles.backText}>返回</span>
        </div>
        <div className={styles.detailTop}>
          <div className={styles.topLeft}>
            <div ref={detailDescRef}>
              <DetailDesc mcpDetail={mcpDetail} />
            </div>
          </div>
          <div className={styles.topRight}>
            {mcpDetail?.envRequired === 0 && (
              <Button
                type="primary"
                onClick={handleAddMcp}
                loading={downMcpLoading || authMcpLoading}
                disabled={mcpDetail.status === 1}
              >
                {mcpDetail.status === 0 ? '添加' : '已添加'}
              </Button>
            )}
            {mcpDetail.envRequired === 1 &&
              (mcpDetail.status === 0 ? (
                <Button
                  type="primary"
                  onClick={handleAddMcp}
                  loading={downMcpLoading || authMcpLoading}
                >
                  添加
                </Button>
              ) : (
                <div style={{ display: 'flex', gap: '8px', alignItems: 'center', justifyContent: 'flex-end' }}>
                  <Button
                    type="primary"
                    onClick={handleAddMcp}
                    loading={downMcpLoading || authMcpLoading}
                  >
                    重新添加
                  </Button>
                  <Button
                    type="default"
                    loading={cancelMcpLoading}
                    onClick={handleCancelMcp}
                    disabled={downMcpLoading || authMcpLoading}
                  >
                    取消添加
                  </Button>
                </div>
              ))}
          </div>
        </div>
        {/*分割线*/}
        <div className={styles.Line}></div>
        <div
          className={styles.detailContent}
          style={{ height: `calc(100vh - ${descHeight + 147}px)` }}
        >
          <div className={styles.contentLeft}>
            <Tabs
              className={styles.tabs}
              defaultActiveKey="overView"
              items={items}
              // onChange={onChange}
            />
          </div>
          <div className={styles.contentRight}>
            <McpServiceConfig code={JSON.stringify(mcpDetail.serverConfig || '暂无配置数据', null, 2)} />
            <RecommendedClient />
          </div>
        </div>
        <McpAuthModal
          mcpDetail={mcpDetail}
          handleAuthMcp={handleAuthMcp}
          setShowMcpModal={setShowMcpModal}
          showMcpModal={showMcpModal}
        />
      </div>
    )
  );
}
