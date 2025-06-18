import { List } from 'antd';
import McpCard from '../mcp-card';
import { IMcpListItem } from '../mcp-square-tab/types';
import styles from './index.module.scss';
import realLoadingSvg from '@/components/icons/real-loading.svg';
import { IMcpCommonProps } from '@/components/mcp-manage/mcp-common';
import type { ListGridType } from 'antd/es/list';
import { useEffect, useState } from 'react';
import useSelectMcpStore from '@/store/useSelectMcpStore.ts';

interface IMcpListProps extends IMcpCommonProps {
  mcpListData: IMcpListItem[];
  pagination: {
    current: number;
    pageSize: number;
    total: number;
  };
  onPageChange?: (current: number, pageSize: number) => void;
  handelMcpCardClick: (serviceId: string | number) => void;
  pageLoading?: boolean;
  grid?: ListGridType;
  isSelectable?: boolean;
  handleMcpListToPage?: (page?: number) => void;
  isMyMcp?: boolean; // 是否是我的MCP
}
export default function McpList(props: IMcpListProps) {
  const {
    setSelectTemporaryMcpItems,
    selectTemporaryMcpItems,
    mcpListData,
    pagination,
    onPageChange,
    handelMcpCardClick,
    pageLoading = false,
    isSelectable,
    showOnlySelectedMcpList,
    showOnlySelectedMyMcp,
    activeKey,
    isDialog,
    handleMcpListToPage,
    isMyMcp = false,
  } = props;

  const { selectMcpList, setDrawerOpenId } = useSelectMcpStore();
  const [filteredData, setFilteredData] = useState<IMcpListItem[]>([]);
  const handleDetail = (id: string) => {
    if (isDialog) {
      setDrawerOpenId(id as string);
    } else {
      handelMcpCardClick(id);
    }
  };
  useEffect(() => {
    if (activeKey === 'myMcp') {
      if (showOnlySelectedMyMcp) {
        // 根据筛选状态和搜索值过滤数据
        setFilteredData(selectMcpList);
      } else {
        setFilteredData(mcpListData);
      }
    } else {
      if (showOnlySelectedMcpList) {
        // 根据筛选状态和搜索值过滤数据
        setFilteredData(selectMcpList);
      } else {
        setFilteredData(mcpListData);
      }
    }
  }, [showOnlySelectedMyMcp, showOnlySelectedMcpList, mcpListData, activeKey]);
  return (
    <div className={styles.mcpCardList}>
      {pageLoading ? (
        <div className={styles.loading}>
          <img
            src={realLoadingSvg}
            alt="loading"
          />
        </div>
      ) : (
        <List
          rowKey="id"
          grid={props?.grid}
          dataSource={filteredData}
          pagination={
            showOnlySelectedMyMcp || showOnlySelectedMcpList
              ? false // 筛选状态下强制关闭分页
              : pagination.total >= pagination.pageSize
                ? { className: styles.mcpListPagination, align: 'end', ...pagination, pageSizeOptions: [12, 24, 48, 96], showSizeChanger: true, onChange: onPageChange }
                : false
          }
          renderItem={(item) => (
            <List.Item key={item.id}>
              <McpCard
                setSelectTemporaryMcpItems={setSelectTemporaryMcpItems}
                selectTemporaryMcpItems={selectTemporaryMcpItems}
                isSelectable={isSelectable}
                mcpData={item}
                handelMcpCardClick={(id) => {
                  handleDetail(id as string);
                }}
                handleMcpListToPage={handleMcpListToPage}
                isMyMcp={isMyMcp}
              />
            </List.Item>
          )}
        />
      )}
    </div>
  );
}
