import { useViewModel } from './view-model';
import McpList from '../mcp-list';
import { genGrid, IMcpCommonProps, renderSearch, renderTitle, renderWarp } from '@/components/mcp-manage/mcp-common';
import { useEffect } from 'react';

export interface IMyMcpTabProps extends IMcpCommonProps {}

export default function MyMcpTab(props: IMyMcpTabProps) {
  const vm = useViewModel();
  const { isDialog, showOnlySelectedMcpList, showOnlySelectedMyMcp, activeKey, setSelectTemporaryMcpItems, selectTemporaryMcpItems } = props;
  console.log('activeKey', activeKey);
  useEffect(() => {
    if (activeKey === 'myMcp') {
      vm.handleMcpListToPage();
    }
  }, [activeKey]);
  return renderWarp({
    isDialog,
    childrenNode: (
      <McpList
        setSelectTemporaryMcpItems={setSelectTemporaryMcpItems}
        selectTemporaryMcpItems={selectTemporaryMcpItems}
        showOnlySelectedMcpList={showOnlySelectedMcpList}
        showOnlySelectedMyMcp={showOnlySelectedMyMcp}
        activeKey={activeKey}
        isSelectable={isDialog}
        isDialog={isDialog}
        grid={genGrid({ isDialog })}
        mcpListData={vm.mcpListData}
        setMcpListData={vm.setMcpListData}
        pagination={vm.pagination}
        onPageChange={vm.handlePageChange}
        handelMcpCardClick={vm.handelMcpCardClick}
        pageLoading={vm.mcpListLoading}
        handleMcpListToPage={vm.handleMcpListToPage}
        isMyMcp={true}
      />
    ),
    titleNode: (
      <>
        {isDialog
          ? null
          : renderTitle({
              title: '我的MCP',
              count: vm.pagination.total ?? 0,
            })}
        {!showOnlySelectedMyMcp && renderSearch({ vm })}
      </>
    ),
  });
}
