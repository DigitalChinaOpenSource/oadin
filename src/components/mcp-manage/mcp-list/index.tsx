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
}
export default function McpList(props: IMcpListProps) {
  const { mcpListData, pagination, onPageChange, handelMcpCardClick, pageLoading = false, isSelectable, showOnlySelectedMcpList, showOnlySelectedMyMcp, activeKey } = props;

  const { selectMcpList } = useSelectMcpStore();
  const [filteredData, setFilteredData] = useState<IMcpListItem[]>([]);
  useEffect(() => {
    if (activeKey === 'myMcp') {
      if (showOnlySelectedMyMcp) {
        // 根据筛选状态和搜索值过滤数据
        const _filteredData = mcpListData.filter((item) => {
          return showOnlySelectedMyMcp
            ? selectMcpList
                .map((mcpItem) => {
                  return mcpItem?.id;
                })
                .includes(item.id)
            : true;
        });
        setFilteredData(_filteredData);
      } else {
        setFilteredData(mcpListData);
      }
    } else {
      if (showOnlySelectedMcpList) {
        // 根据筛选状态和搜索值过滤数据
        const _filteredData = mcpListData.filter((item) => {
          return showOnlySelectedMcpList
            ? selectMcpList
                .map((mcpItem) => {
                  return mcpItem?.id;
                })
                .includes(item.id)
            : true;
        });
        setFilteredData(_filteredData);
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
              : filteredData.length >= 12
                ? { className: styles.mcpListPagination, align: 'end', ...pagination, pageSizeOptions: [12, 24, 48, 96], showSizeChanger: true, onChange: onPageChange }
                : false
          }
          renderItem={(item) => (
            <List.Item key={item.id}>
              <McpCard
                isSelectable={isSelectable}
                mcpData={item}
                handelMcpCardClick={handelMcpCardClick}
              />
            </List.Item>
          )}
        />
      )}
    </div>
  );
}
